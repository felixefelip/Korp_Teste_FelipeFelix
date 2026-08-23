package db

import (
	"errors"
	"time"

	"billing/internal/model"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func asDuplicated(err error) error {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return model.ErrInvoiceDuplicated
	}

	return err
}

type InvoiceRepository struct {
	connection *gorm.DB
}

func NewInvoiceRepository(connection *gorm.DB) *InvoiceRepository {
	return &InvoiceRepository{
		connection: connection,
	}
}

func (ir *InvoiceRepository) GetInvoices() ([]model.Invoice, error) {
	var invoiceList []model.Invoice

	err := withItems(ir.connection).Order("series desc, number desc").Find(&invoiceList).Error
	if err != nil {
		return []model.Invoice{}, err
	}

	if err := fillProcessingSince(ir.connection, invoiceList); err != nil {
		return []model.Invoice{}, err
	}

	return invoiceList, nil
}

func (ir *InvoiceRepository) GetInvoiceByID(id int) (model.Invoice, error) {
	var invoice model.Invoice

	err := withItems(ir.connection).First(&invoice, id).Error
	if err != nil {
		return model.Invoice{}, err
	}

	single := []model.Invoice{invoice}

	if err := fillProcessingSince(ir.connection, single); err != nil {
		return model.Invoice{}, err
	}

	return single[0], nil
}

type lastRequest struct {
	AggregateID int
	CreatedAt   time.Time
}

func fillProcessingSince(connection *gorm.DB, invoices []model.Invoice) error {
	ids := make([]int, 0, len(invoices))

	for _, invoice := range invoices {
		if invoice.Processing() {
			ids = append(ids, invoice.ID)
		}
	}

	if len(ids) == 0 {
		return nil
	}

	var requests []lastRequest

	err := connection.
		Model(&model.OutboxEvent{}).
		Select("aggregate_id, MAX(created_at) AS created_at").
		Where("aggregate_type = ? AND aggregate_id IN ?", model.OutboxAggregateInvoice, ids).
		Group("aggregate_id").
		Scan(&requests).Error
	if err != nil {
		return err
	}

	requestedAt := make(map[int]time.Time, len(requests))

	for _, request := range requests {
		requestedAt[request.AggregateID] = request.CreatedAt
	}

	for index := range invoices {
		if moment, asked := requestedAt[invoices[index].ID]; asked {
			invoices[index].ProcessingSince = &moment
		}
	}

	return nil
}

func (ir *InvoiceRepository) CreateInvoice(invoice model.Invoice) (int, error) {
	err := ir.connection.Transaction(func(tx *gorm.DB) error {
		items := invoice.Items
		invoice.Items = nil

		if err := tx.Create(&invoice).Error; err != nil {
			return err
		}

		return saveItems(tx, invoice.ID, items)
	})
	if err != nil {
		return 0, asDuplicated(err)
	}

	return invoice.ID, nil
}

func (ir *InvoiceRepository) UpdateInvoice(invoice model.Invoice) error {
	return asDuplicated(ir.connection.Transaction(func(tx *gorm.DB) error {
		result := tx.
			Model(&model.Invoice{ID: invoice.ID}).
			Updates(map[string]any{"series": invoice.Series, "number": invoice.Number})
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		if invoice.Items == nil {
			return nil
		}

		if err := deleteInvoiceItemsByInvoiceID(tx, invoice.ID); err != nil {
			return err
		}

		return saveItems(tx, invoice.ID, invoice.Items)
	}))
}

func (ir *InvoiceRepository) CloseInvoice(id int, event model.OutboxEvent) error {
	return ir.connection.Transaction(func(tx *gorm.DB) error {
		if err := startTransition(tx, id, model.InvoiceStatusClosing); err != nil {
			return err
		}

		return tx.Create(&event).Error
	})
}

func (ir *InvoiceRepository) ConfirmClose(id int) (bool, error) {
	return ir.moveFrom(id, model.InvoiceStatusClosing, model.InvoiceStatusClosed, "", nil)
}

func (ir *InvoiceRepository) RejectClose(
	id int,
	reason string,
	shortages []model.InvoiceShortage,
) (bool, error) {
	return ir.moveFrom(id, model.InvoiceStatusClosing, model.InvoiceStatusOpen, reason, shortages)
}

func (ir *InvoiceRepository) ReopenInvoice(id int, event model.OutboxEvent) error {
	return ir.connection.Transaction(func(tx *gorm.DB) error {
		if err := startTransition(tx, id, model.InvoiceStatusReopening); err != nil {
			return err
		}

		return tx.Create(&event).Error
	})
}

func (ir *InvoiceRepository) RetryInvoice(id int, status string, event model.OutboxEvent) error {
	return ir.connection.Transaction(func(tx *gorm.DB) error {
		var processing int64

		err := tx.
			Model(&model.Invoice{}).
			Where("id = ? AND status = ?", id, status).
			Count(&processing).Error
		if err != nil {
			return err
		}

		if processing == 0 {
			return model.ErrInvoiceNotProcessing
		}

		return tx.Create(&event).Error
	})
}

func (ir *InvoiceRepository) ConfirmReopen(id int) (bool, error) {
	return ir.moveFrom(id, model.InvoiceStatusReopening, model.InvoiceStatusOpen, "", nil)
}

func (ir *InvoiceRepository) RejectReopen(
	id int,
	reason string,
	shortages []model.InvoiceShortage,
) (bool, error) {
	return ir.moveFrom(id, model.InvoiceStatusReopening, model.InvoiceStatusClosed, reason, shortages)
}

func (ir *InvoiceRepository) moveFrom(
	id int,
	from, to, reason string,
	shortages []model.InvoiceShortage,
) (bool, error) {
	moved := false

	err := ir.connection.Transaction(func(tx *gorm.DB) error {
		result := tx.
			Model(&model.Invoice{}).
			Where("id = ? AND status = ?", id, from).
			Updates(map[string]any{"status": to, "failure_reason": reason})
		if result.Error != nil {
			return result.Error
		}

		moved = result.RowsAffected > 0

		if !moved || len(shortages) == 0 {
			return nil
		}

		for index := range shortages {
			shortages[index].ID = 0
			shortages[index].InvoiceID = id
		}

		return tx.Create(&shortages).Error
	})

	return moved, err
}

func startTransition(tx *gorm.DB, id int, status string) error {
	result := tx.
		Model(&model.Invoice{ID: id}).
		Updates(map[string]any{"status": status, "failure_reason": ""})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return deleteShortagesByInvoiceID(tx, id)
}

func deleteShortagesByInvoiceID(tx *gorm.DB, invoiceID int) error {
	return tx.Where("invoice_id = ?", invoiceID).Delete(&model.InvoiceShortage{}).Error
}

func (ir *InvoiceRepository) DeleteInvoice(id int) error {
	return ir.connection.Transaction(func(tx *gorm.DB) error {
		if err := deleteShortagesByInvoiceID(tx, id); err != nil {
			return err
		}

		if err := deleteInvoiceItemsByInvoiceID(tx, id); err != nil {
			return err
		}

		result := tx.Delete(&model.Invoice{}, id)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
}

func deleteInvoiceItemsByInvoiceID(tx *gorm.DB, invoiceID int) error {
	return tx.Where("invoice_id = ?", invoiceID).Delete(&model.InvoiceItem{}).Error
}

func withItems(connection *gorm.DB) *gorm.DB {
	return connection.
		Preload("Items", func(items *gorm.DB) *gorm.DB {
			return items.Order("invoice_item.id")
		}).
		Preload("Items.Product").
		Preload("Shortages", func(shortages *gorm.DB) *gorm.DB {
			return shortages.Order("invoice_shortage.product_code")
		})
}

func saveItems(tx *gorm.DB, invoiceID int, items []model.InvoiceItem) error {
	for _, item := range items {
		product, err := ensureProduct(tx, item.Product)
		if err != nil {
			return err
		}

		item.ID = 0
		item.InvoiceID = invoiceID
		item.ProductID = product.ID

		if err := tx.Omit("Product").Create(&item).Error; err != nil {
			return err
		}
	}

	return nil
}

func ensureProduct(tx *gorm.DB, product model.Product) (model.Product, error) {
	err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "inventory_id"}},
		DoNothing: true,
	}).Create(&product).Error
	if err != nil {
		return model.Product{}, err
	}

	if product.ID != 0 {
		return product, nil
	}

	var stored model.Product
	if err := tx.Where("inventory_id = ?", product.InventoryID).First(&stored).Error; err != nil {
		return model.Product{}, err
	}

	return stored, nil
}
