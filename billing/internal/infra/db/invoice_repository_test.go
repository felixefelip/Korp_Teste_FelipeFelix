package db_test

import (
	"testing"

	"billing/internal/infra/db"
	"billing/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestGetInvoicesReturnsEverythingStored(t *testing.T) {
	repository := newRepository(t)

	_, err := repository.CreateInvoice(model.Invoice{Series: 1, Number: 1, Status: "OPEN"})
	require.NoError(t, err)

	_, err = repository.CreateInvoice(model.Invoice{Series: 1, Number: 2, Status: "CLOSED"})
	require.NoError(t, err)

	invoices, err := repository.GetInvoices()
	require.NoError(t, err)

	require.Len(t, invoices, 2)
	assert.ElementsMatch(t,
		[]int{1, 2},
		[]int{invoices[0].Number, invoices[1].Number},
	)
}

func TestGetInvoicesReturnsTheStoredFields(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(model.Invoice{
		Series: 1, Number: 1, Type: model.InvoiceTypeOut, Status: "CLOSED",
	})
	require.NoError(t, err)

	invoices, err := repository.GetInvoices()
	require.NoError(t, err)

	require.Len(t, invoices, 1)
	assert.Equal(t, model.Invoice{
		ID:     id,
		Series: 1, Number: 1,
		Type:      model.InvoiceTypeOut,
		Status:    "CLOSED",
		Items:     []model.InvoiceItem{},
		Shortages: []model.InvoiceShortage{},
	}, invoices[0])
}

func TestGetInvoicesWithNothingStoredReturnsAnEmptyList(t *testing.T) {
	repository := newRepository(t)

	invoices, err := repository.GetInvoices()

	require.NoError(t, err)
	assert.Empty(t, invoices)
}

func TestGetInvoiceByIDReturnsTheInvoice(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(model.Invoice{Series: 1, Number: 1, Status: "OPEN"})
	require.NoError(t, err)

	invoice, err := repository.GetInvoiceByID(id)

	require.NoError(t, err)
	assert.Equal(t, model.Invoice{
		ID:     id,
		Series: 1, Number: 1,
		Type:      model.InvoiceTypeOut,
		Status:    "OPEN",
		Items:     []model.InvoiceItem{},
		Shortages: []model.InvoiceShortage{},
	}, invoice)
}

func TestGetInvoiceByIDWhenMissingReturnsErrRecordNotFound(t *testing.T) {
	repository := newRepository(t)

	_, err := repository.GetInvoiceByID(404)

	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestCreateInvoice(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(model.Invoice{Series: 1, Number: 1, Status: "OPEN"})
	require.NoError(t, err)
	assert.NotZero(t, id, "the database should have generated an id")

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)

	assert.Equal(t, 1, saved.Number)
	assert.Equal(t, "OPEN", saved.Status)
}

func TestCreateInvoiceKeepsTheClosedStatus(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(model.Invoice{Series: 1, Number: 2, Status: "CLOSED"})
	require.NoError(t, err)

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)

	assert.Equal(t, "CLOSED", saved.Status)
}

func TestCreateInvoiceGeneratesSequentialIDs(t *testing.T) {
	repository := newRepository(t)

	first, err := repository.CreateInvoice(model.Invoice{Series: 1, Number: 1, Status: "OPEN"})
	require.NoError(t, err)

	second, err := repository.CreateInvoice(model.Invoice{Series: 1, Number: 2, Status: "OPEN"})
	require.NoError(t, err)

	assert.Greater(t, second, first, "each invoice should get its own id")
}

func TestCreateInvoiceRefusesADuplicateNumberInTheSameSeries(t *testing.T) {
	repository := newRepository(t)

	_, err := repository.CreateInvoice(model.Invoice{Series: 1, Number: 1, Status: "OPEN"})
	require.NoError(t, err)

	_, err = repository.CreateInvoice(model.Invoice{Series: 1, Number: 1, Status: "OPEN"})

	assert.ErrorIs(t, err, model.ErrInvoiceDuplicated)
}

func TestCreateInvoiceAcceptsTheSameNumberInAnotherSeries(t *testing.T) {
	repository := newRepository(t)

	_, err := repository.CreateInvoice(model.Invoice{Series: 1, Number: 1, Status: "OPEN"})
	require.NoError(t, err)

	_, err = repository.CreateInvoice(model.Invoice{Series: 2, Number: 1, Status: "OPEN"})

	assert.NoError(t, err, "the number is unique within its series, not globally")
}

func TestUpdateInvoiceRefusesANumberAnotherInvoiceAlreadyUses(t *testing.T) {
	repository := newRepository(t)

	_, err := repository.CreateInvoice(model.Invoice{Series: 1, Number: 1, Status: "OPEN"})
	require.NoError(t, err)

	second, err := repository.CreateInvoice(model.Invoice{Series: 1, Number: 2, Status: "OPEN"})
	require.NoError(t, err)

	err = repository.UpdateInvoice(model.Invoice{ID: second, Series: 1, Number: 1})

	assert.ErrorIs(t, err, model.ErrInvoiceDuplicated)
}

func TestUpdateInvoiceChangesEveryEditableField(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(model.Invoice{
		Series: 1, Number: 1, Type: model.InvoiceTypeOut, Status: "OPEN",
	})
	require.NoError(t, err)

	err = repository.UpdateInvoice(model.Invoice{
		ID: id, Series: 1, Number: 2, Status: model.InvoiceStatusClosed,
	})
	require.NoError(t, err)

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)

	assert.Equal(t, model.Invoice{
		ID:     id,
		Series: 1, Number: 2,
		Type:   model.InvoiceTypeOut,
		Status: model.InvoiceStatusOpen,
	}, saved, "only CloseInvoice moves the status")
}

func TestUpdateInvoiceWhenMissingReturnsErrRecordNotFound(t *testing.T) {
	repository := newRepository(t)

	err := repository.UpdateInvoice(model.Invoice{ID: 9999, Series: 1, Number: 1, Status: "OPEN"})

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUpdateInvoiceLeavesTheOtherInvoicesAlone(t *testing.T) {
	repository := newRepository(t)

	first, err := repository.CreateInvoice(model.Invoice{Series: 1, Number: 1, Status: "OPEN"})
	require.NoError(t, err)

	second, err := repository.CreateInvoice(model.Invoice{Series: 1, Number: 2, Status: "OPEN"})
	require.NoError(t, err)

	require.NoError(t, repository.UpdateInvoice(model.Invoice{ID: first, Series: 1, Number: 1, Status: "CLOSED"}))

	var untouched model.Invoice
	require.NoError(t, testConnection.First(&untouched, second).Error)

	assert.Equal(t, 2, untouched.Number)
	assert.Equal(t, "OPEN", untouched.Status)
}

func itemOf(inventoryID int, code, name string, quantity int, price float64) model.InvoiceItem {
	return model.InvoiceItem{
		ProductCode: code,
		ProductName: name,
		Unit:        "UN",
		Quantity:    quantity,
		UnitPrice:   price,
		Product: model.Product{
			InventoryID: inventoryID,
			Code:        code,
			Name:        name,
			Unit:        "UN",
			Price:       price,
		},
	}
}

var fixtureNumber = 100

func invoiceWithItems(items ...model.InvoiceItem) model.Invoice {
	fixtureNumber++

	return model.Invoice{
		Series: 1, Number: fixtureNumber, Type: model.InvoiceTypeOut, Status: "OPEN", Items: items,
	}
}

func TestCreateInvoiceStoresTheItems(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(invoiceWithItems(
		itemOf(11, "PRD-0001", "Camiseta", 2, 30.99),
		itemOf(12, "PRD-0002", "Caneca", 1, 19.9),
	))
	require.NoError(t, err)

	invoice, err := repository.GetInvoiceByID(id)
	require.NoError(t, err)

	require.Len(t, invoice.Items, 2)
	assert.Equal(t, "PRD-0001", invoice.Items[0].ProductCode)
	assert.Equal(t, "Camiseta", invoice.Items[0].ProductName)
	assert.Equal(t, 2, invoice.Items[0].Quantity)
	assert.Equal(t, 30.99, invoice.Items[0].UnitPrice)
	assert.Equal(t, id, invoice.Items[0].InvoiceID)
	assert.Equal(t, "Caneca", invoice.Items[1].ProductName)
}

func TestCreateInvoiceRegistersTheProductOfEachItem(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(invoiceWithItems(itemOf(11, "PRD-0001", "Camiseta", 2, 30.99)))
	require.NoError(t, err)

	invoice, err := repository.GetInvoiceByID(id)
	require.NoError(t, err)

	require.Len(t, invoice.Items, 1)

	var products []model.Product
	require.NoError(t, testConnection.Find(&products).Error)

	require.Len(t, products, 1)
	assert.NotZero(t, products[0].ID, "the billing product has an id of its own")
	assert.Equal(t, 11, products[0].InventoryID)
	assert.Equal(t, "PRD-0001", products[0].Code)
	assert.Equal(t, products[0].ID, invoice.Items[0].ProductID, "the item points at the local id")
	assert.Equal(t, 11, invoice.Items[0].Product.InventoryID)
}

func TestCreateInvoiceReusesTheProductAlreadyRegistered(t *testing.T) {
	repository := newRepository(t)

	first, err := repository.CreateInvoice(invoiceWithItems(itemOf(11, "PRD-0001", "Camiseta", 2, 30.99)))
	require.NoError(t, err)

	second, err := repository.CreateInvoice(invoiceWithItems(itemOf(11, "PRD-0001", "Camiseta", 1, 30.99)))
	require.NoError(t, err)

	var products []model.Product
	require.NoError(t, testConnection.Find(&products).Error)
	assert.Len(t, products, 1, "the same inventory product must not be registered twice")

	firstInvoice, err := repository.GetInvoiceByID(first)
	require.NoError(t, err)

	secondInvoice, err := repository.GetInvoiceByID(second)
	require.NoError(t, err)

	assert.Equal(t, firstInvoice.Items[0].ProductID, secondInvoice.Items[0].ProductID)
}

func TestCreateInvoiceLeavesAnAlreadySyncedProductAlone(t *testing.T) {
	repository := newRepository(t)

	require.NoError(t, testConnection.Create(&model.Product{
		InventoryID: 11, Code: "PRD-0001", Name: "Camiseta", Unit: "UN", Price: 30.99, Active: true,
	}).Error)

	_, err := repository.CreateInvoice(invoiceWithItems(
		itemOf(11, "PRD-0001", "Camiseta polo", 1, 45.5),
	))
	require.NoError(t, err)

	var product model.Product
	require.NoError(t, testConnection.Where("inventory_id = ?", 11).First(&product).Error)

	assert.Equal(t, "Camiseta", product.Name, "the catalog belongs to the inventory sync")
	assert.Equal(t, 30.99, product.Price)
}

func TestCreateInvoiceRegistersAProductTheSyncHasNotBroughtYet(t *testing.T) {
	repository := newRepository(t)

	_, err := repository.CreateInvoice(invoiceWithItems(
		itemOf(11, "PRD-0001", "Camiseta", 2, 30.99),
	))
	require.NoError(t, err)

	var product model.Product
	require.NoError(t, testConnection.Where("inventory_id = ?", 11).First(&product).Error)

	assert.Equal(t, "PRD-0001", product.Code, "the item still needs a product row to point at")
}

func TestInvoiceItemKeepsItsOwnSnapshotOfTheProduct(t *testing.T) {
	repository := newRepository(t)

	require.NoError(t, testConnection.Create(&model.Product{
		InventoryID: 11, Code: "PRD-0001", Name: "Camiseta", Unit: "UN", Price: 30.99, Active: true,
	}).Error)

	id, err := repository.CreateInvoice(invoiceWithItems(
		itemOf(11, "PRD-0001", "Camiseta polo", 1, 45.5),
	))
	require.NoError(t, err)

	invoice, err := repository.GetInvoiceByID(id)
	require.NoError(t, err)

	require.Len(t, invoice.Items, 1)
	assert.Equal(t, "Camiseta polo", invoice.Items[0].ProductName,
		"the item froze what was emitted, the replica keeps the catalog")
	assert.Equal(t, 45.5, invoice.Items[0].UnitPrice)
}

func TestCreateInvoiceKeepsTheItemAsItWasSold(t *testing.T) {
	repository := newRepository(t)

	sold, err := repository.CreateInvoice(invoiceWithItems(itemOf(11, "PRD-0001", "Camiseta", 2, 30.99)))
	require.NoError(t, err)

	_, err = repository.CreateInvoice(invoiceWithItems(itemOf(11, "PRD-0001", "Camiseta polo", 1, 45.5)))
	require.NoError(t, err)

	invoice, err := repository.GetInvoiceByID(sold)
	require.NoError(t, err)

	require.Len(t, invoice.Items, 1)
	assert.Equal(t, "Camiseta", invoice.Items[0].ProductName, "the invoice keeps what was sold")
	assert.Equal(t, 30.99, invoice.Items[0].UnitPrice, "the price of the sale must not follow the product")
}

func TestGetInvoicesBringsTheItemsOfEachInvoice(t *testing.T) {
	repository := newRepository(t)

	newest := invoiceWithItems(itemOf(11, "PRD-0001", "Camiseta", 2, 30.99))
	newest.Number = 2

	_, err := repository.CreateInvoice(newest)
	require.NoError(t, err)

	_, err = repository.CreateInvoice(model.Invoice{Series: 1, Number: 1, Status: "OPEN"})
	require.NoError(t, err)

	invoices, err := repository.GetInvoices()
	require.NoError(t, err)

	require.Len(t, invoices, 2)
	assert.Len(t, invoices[0].Items, 1)
	assert.Equal(t, 11, invoices[0].Items[0].Product.InventoryID)
	assert.Empty(t, invoices[1].Items)
}

func TestUpdateInvoiceReplacesTheItems(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(invoiceWithItems(
		itemOf(11, "PRD-0001", "Camiseta", 2, 30.99),
		itemOf(12, "PRD-0002", "Caneca", 1, 19.9),
	))
	require.NoError(t, err)

	err = repository.UpdateInvoice(model.Invoice{
		ID:     id,
		Series: 1, Number: 1,
		Status: "CLOSED",
		Items:  []model.InvoiceItem{itemOf(13, "PRD-0003", "Mochila", 3, 99.9)},
	})
	require.NoError(t, err)

	invoice, err := repository.GetInvoiceByID(id)
	require.NoError(t, err)

	require.Len(t, invoice.Items, 1)
	assert.Equal(t, "Mochila", invoice.Items[0].ProductName)
	assert.Equal(t, 3, invoice.Items[0].Quantity)

	var stored []model.InvoiceItem
	require.NoError(t, testConnection.Find(&stored).Error)
	assert.Len(t, stored, 1, "the replaced items must not stay behind")
}

func TestUpdateInvoiceWithAnEmptyListEmptiesTheInvoice(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(invoiceWithItems(itemOf(11, "PRD-0001", "Camiseta", 2, 30.99)))
	require.NoError(t, err)

	err = repository.UpdateInvoice(model.Invoice{
		ID: id, Series: 1, Number: 1, Status: "OPEN", Items: []model.InvoiceItem{},
	})
	require.NoError(t, err)

	invoice, err := repository.GetInvoiceByID(id)
	require.NoError(t, err)

	assert.Empty(t, invoice.Items)
}

func TestUpdateInvoiceWithoutAListKeepsTheItemsAlreadyThere(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(invoiceWithItems(itemOf(11, "PRD-0001", "Camiseta", 2, 30.99)))
	require.NoError(t, err)

	err = repository.UpdateInvoice(model.Invoice{ID: id, Series: 1, Number: 9})
	require.NoError(t, err)

	invoice, err := repository.GetInvoiceByID(id)
	require.NoError(t, err)

	assert.Equal(t, 9, invoice.Number)
	require.Len(t, invoice.Items, 1, "an update that says nothing about the items must not erase them")
	assert.Equal(t, "Camiseta", invoice.Items[0].ProductName)
}

func TestUpdateInvoiceKeepsTheProductAlreadyRegistered(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(invoiceWithItems(itemOf(11, "PRD-0001", "Camiseta", 2, 30.99)))
	require.NoError(t, err)

	err = repository.UpdateInvoice(model.Invoice{
		ID:     id,
		Series: 1, Number: 1,
		Status: "OPEN",
		Items:  []model.InvoiceItem{itemOf(11, "PRD-0001", "Camiseta", 5, 30.99)},
	})
	require.NoError(t, err)

	var products []model.Product
	require.NoError(t, testConnection.Find(&products).Error)
	assert.Len(t, products, 1)
}

func TestUpdateInvoiceOfAMissingInvoiceStoresNoItem(t *testing.T) {
	repository := newRepository(t)

	err := repository.UpdateInvoice(model.Invoice{
		ID:     9999,
		Series: 1, Number: 1,
		Status: "OPEN",
		Items:  []model.InvoiceItem{itemOf(11, "PRD-0001", "Camiseta", 2, 30.99)},
	})

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	var stored []model.InvoiceItem
	require.NoError(t, testConnection.Find(&stored).Error)
	assert.Empty(t, stored, "the transaction must have rolled back")

	var products []model.Product
	require.NoError(t, testConnection.Find(&products).Error)
	assert.Empty(t, products)
}

func TestUpdateInvoiceLeavesTheItemsOfTheOtherInvoicesAlone(t *testing.T) {
	repository := newRepository(t)

	first, err := repository.CreateInvoice(invoiceWithItems(itemOf(11, "PRD-0001", "Camiseta", 2, 30.99)))
	require.NoError(t, err)

	second, err := repository.CreateInvoice(invoiceWithItems(itemOf(12, "PRD-0002", "Caneca", 1, 19.9)))
	require.NoError(t, err)

	err = repository.UpdateInvoice(model.Invoice{
		ID: first, Series: 1, Number: 1, Status: "CLOSED", Items: []model.InvoiceItem{},
	})
	require.NoError(t, err)

	untouched, err := repository.GetInvoiceByID(second)
	require.NoError(t, err)

	require.Len(t, untouched.Items, 1)
	assert.Equal(t, "Caneca", untouched.Items[0].ProductName)
}

func TestUpdateInvoiceNeverFlipsTheDirection(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(model.Invoice{
		Series: 1, Number: 1, Type: model.InvoiceTypeIn, Status: "OPEN",
	})
	require.NoError(t, err)

	err = repository.UpdateInvoice(model.Invoice{
		ID: id, Series: 1, Number: 1, Type: model.InvoiceTypeOut, Status: "CLOSED",
	})
	require.NoError(t, err)

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)

	assert.Equal(t, model.InvoiceTypeIn, saved.Type,
		"the direction is settled at issue; it is not a field an edit can swing")
}

func TestDeleteInvoiceRemovesTheRowAndItsItems(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(invoiceWithItems(
		itemOf(11, "PRD-0001", "Camiseta", 2, 30.99),
	))
	require.NoError(t, err)

	require.NoError(t, repository.DeleteInvoice(id))

	var invoices []model.Invoice
	require.NoError(t, testConnection.Find(&invoices).Error)
	assert.Empty(t, invoices)

	var items []model.InvoiceItem
	require.NoError(t, testConnection.Find(&items).Error)
	assert.Empty(t, items, "an item without its invoice is an orphan")
}

func TestDeleteInvoiceKeepsTheProductReplica(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(invoiceWithItems(
		itemOf(11, "PRD-0001", "Camiseta", 2, 30.99),
	))
	require.NoError(t, err)

	require.NoError(t, repository.DeleteInvoice(id))

	var products []model.Product
	require.NoError(t, testConnection.Find(&products).Error)
	assert.Len(t, products, 1, "the replica is shared by every invoice")
}

func TestDeleteInvoiceLeavesTheOtherInvoicesAlone(t *testing.T) {
	repository := newRepository(t)

	first, err := repository.CreateInvoice(invoiceWithItems(
		itemOf(11, "PRD-0001", "Camiseta", 2, 30.99),
	))
	require.NoError(t, err)

	second, err := repository.CreateInvoice(invoiceWithItems(
		itemOf(12, "PRD-0002", "Caneca", 1, 19.9),
	))
	require.NoError(t, err)

	require.NoError(t, repository.DeleteInvoice(first))

	var invoices []model.Invoice
	require.NoError(t, testConnection.Find(&invoices).Error)
	require.Len(t, invoices, 1)
	assert.Equal(t, second, invoices[0].ID)

	var items []model.InvoiceItem
	require.NoError(t, testConnection.Find(&items).Error)
	assert.Len(t, items, 1)
}

func TestDeleteInvoiceWhenMissingReturnsErrRecordNotFound(t *testing.T) {
	repository := newRepository(t)

	err := repository.DeleteInvoice(9999)

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestCloseInvoiceMovesTheStatus(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(model.Invoice{
		Series: 1, Number: 1, Type: model.InvoiceTypeOut, Status: model.InvoiceStatusOpen,
	})
	require.NoError(t, err)

	require.NoError(t, repository.CloseInvoice(id, closeEventFor(t, id)))

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)

	assert.Equal(t, model.InvoiceStatusClosing, saved.Status,
		"the invoice waits for the inventory before being closed")
	assert.Equal(t, 1, saved.Number, "closing touches the status and nothing else")
}

func TestCloseInvoiceKeepsTheItems(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(invoiceWithItems(
		itemOf(11, "PRD-0001", "Camiseta", 2, 30.99),
	))
	require.NoError(t, err)

	require.NoError(t, repository.CloseInvoice(id, closeEventFor(t, id)))

	invoice, err := repository.GetInvoiceByID(id)
	require.NoError(t, err)

	assert.Len(t, invoice.Items, 1)
}

func TestCloseInvoiceWhenMissingReturnsErrRecordNotFound(t *testing.T) {
	repository := newRepository(t)

	err := repository.CloseInvoice(9999, closeEventFor(t, 9999))

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestCloseInvoiceRecordsTheEventAlongsideTheStatus(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(model.Invoice{
		Series: 1, Number: 1, Type: model.InvoiceTypeOut, Status: model.InvoiceStatusOpen,
	})
	require.NoError(t, err)

	require.NoError(t, repository.CloseInvoice(id, closeEventFor(t, id)))

	var events []model.OutboxEvent
	require.NoError(t, testConnection.Find(&events).Error)

	require.Len(t, events, 1)
	assert.Equal(t, model.InvoiceCloseRequestedKey, events[0].RoutingKey)
	assert.Equal(t, id, events[0].AggregateID)
	assert.Nil(t, events[0].PublishedAt, "it is the relay that publishes")
}

func TestCloseInvoiceWhenMissingRecordsNoEvent(t *testing.T) {
	repository := newRepository(t)

	err := repository.CloseInvoice(9999, closeEventFor(t, 9999))
	require.Error(t, err)

	var events []model.OutboxEvent
	require.NoError(t, testConnection.Find(&events).Error)

	assert.Empty(t, events, "status and event are written in the same transaction")
}

func TestConfirmCloseMovesFromClosingToClosed(t *testing.T) {
	repository := newRepository(t)
	id := closingInvoice(t, repository)

	moved, err := repository.ConfirmClose(id)
	require.NoError(t, err)
	assert.True(t, moved)

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)
	assert.Equal(t, model.InvoiceStatusClosed, saved.Status)
}

func TestConfirmCloseDoesNothingWhenTheInvoiceIsNotClosing(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(model.Invoice{
		Series: 1, Number: 1, Type: model.InvoiceTypeOut, Status: model.InvoiceStatusOpen,
	})
	require.NoError(t, err)

	moved, err := repository.ConfirmClose(id)
	require.NoError(t, err)
	assert.False(t, moved, "a repeated or late result must not move an open invoice")

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)
	assert.Equal(t, model.InvoiceStatusOpen, saved.Status)
}

func TestConfirmCloseIsIdempotent(t *testing.T) {
	repository := newRepository(t)
	id := closingInvoice(t, repository)

	first, err := repository.ConfirmClose(id)
	require.NoError(t, err)
	require.True(t, first)

	second, err := repository.ConfirmClose(id)
	require.NoError(t, err)
	assert.False(t, second, "the second delivery of the same result changes nothing")
}

func TestRejectCloseSendsTheInvoiceBackToOpenWithTheReason(t *testing.T) {
	repository := newRepository(t)
	id := closingInvoice(t, repository)

	moved, err := repository.RejectClose(id, "INSUFFICIENT_STOCK", nil)
	require.NoError(t, err)
	assert.True(t, moved)

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)
	assert.Equal(t, model.InvoiceStatusOpen, saved.Status)
	assert.Equal(t, "INSUFFICIENT_STOCK", saved.FailureReason)
}

func TestCloseInvoiceClearsThePreviousFailure(t *testing.T) {
	repository := newRepository(t)
	id := closingInvoice(t, repository)

	_, err := repository.RejectClose(id, "INSUFFICIENT_STOCK", nil)
	require.NoError(t, err)

	require.NoError(t, repository.CloseInvoice(id, closeEventFor(t, id)))

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)
	assert.Empty(t, saved.FailureReason, "a new attempt starts without the old reason")
}

func closingInvoice(t *testing.T, repository *db.InvoiceRepository) int {
	t.Helper()

	id, err := repository.CreateInvoice(model.Invoice{
		Series: 1, Number: 1, Type: model.InvoiceTypeOut, Status: model.InvoiceStatusOpen,
	})
	require.NoError(t, err)
	require.NoError(t, repository.CloseInvoice(id, closeEventFor(t, id)))

	return id
}

func TestConfirmReopenMovesFromReopeningToOpen(t *testing.T) {
	repository := newRepository(t)
	id := closingInvoice(t, repository)

	_, err := repository.ConfirmClose(id)
	require.NoError(t, err)
	require.NoError(t, repository.ReopenInvoice(id, reopenEventFor(t, id)))

	moved, err := repository.ConfirmReopen(id)
	require.NoError(t, err)
	assert.True(t, moved)

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)
	assert.Equal(t, model.InvoiceStatusOpen, saved.Status)
}

func TestRejectReopenSendsTheInvoiceBackToClosedWithTheReason(t *testing.T) {
	repository := newRepository(t)
	id := closingInvoice(t, repository)

	_, err := repository.ConfirmClose(id)
	require.NoError(t, err)
	require.NoError(t, repository.ReopenInvoice(id, reopenEventFor(t, id)))

	moved, err := repository.RejectReopen(id, "STOCK_ALREADY_USED", nil)
	require.NoError(t, err)
	assert.True(t, moved)

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)
	assert.Equal(t, model.InvoiceStatusClosed, saved.Status,
		"a refused revert leaves the invoice closed, as it was")
	assert.Equal(t, "STOCK_ALREADY_USED", saved.FailureReason)
}

func TestConfirmReopenDoesNothingWhenTheInvoiceIsNotReopening(t *testing.T) {
	repository := newRepository(t)
	id := closingInvoice(t, repository)

	moved, err := repository.ConfirmReopen(id)
	require.NoError(t, err)
	assert.False(t, moved, "a closing invoice must not be moved by a revert result")
}

func reopenEventFor(t *testing.T, id int) model.OutboxEvent {
	t.Helper()

	event, err := model.NewInvoiceReopenRequested(model.Invoice{ID: id, Series: 1, Number: 1})
	require.NoError(t, err)

	return event
}

func closeEventFor(t *testing.T, id int) model.OutboxEvent {
	t.Helper()

	event, err := model.NewInvoiceCloseRequested(model.Invoice{
		ID: id, Series: 1, Number: 1, Type: model.InvoiceTypeOut,
	})
	require.NoError(t, err)

	return event
}

func TestReopenInvoiceMovesTheStatusBack(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(model.Invoice{
		Series: 1, Number: 1, Type: model.InvoiceTypeOut, Status: model.InvoiceStatusClosed,
	})
	require.NoError(t, err)

	require.NoError(t, repository.ReopenInvoice(id, reopenEventFor(t, id)))

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)

	assert.Equal(t, model.InvoiceStatusReopening, saved.Status,
		"the invoice waits for the inventory to give the stock back")
	assert.Equal(t, 1, saved.Number, "reopening touches the status and nothing else")
}

func TestReopenInvoiceKeepsTheItems(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(invoiceWithItems(
		itemOf(11, "PRD-0001", "Camiseta", 2, 30.99),
	))
	require.NoError(t, err)

	require.NoError(t, repository.CloseInvoice(id, closeEventFor(t, id)))
	require.NoError(t, repository.ReopenInvoice(id, reopenEventFor(t, id)))

	invoice, err := repository.GetInvoiceByID(id)
	require.NoError(t, err)

	assert.Len(t, invoice.Items, 1)
}

func TestReopenInvoiceWhenMissingReturnsErrRecordNotFound(t *testing.T) {
	repository := newRepository(t)

	err := repository.ReopenInvoice(9999, reopenEventFor(t, 9999))

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func shortagesOf(productCode string) []model.InvoiceShortage {
	return []model.InvoiceShortage{
		{InventoryID: 42, ProductCode: productCode, ProductName: "Parafuso", Required: 50, Available: 42},
	}
}

func TestRejectCloseStoresWhatWasMissing(t *testing.T) {
	repository := newRepository(t)
	id := closingInvoice(t, repository)

	moved, err := repository.RejectClose(id, "INSUFFICIENT_STOCK", shortagesOf("PROD-1"))
	require.NoError(t, err)
	require.True(t, moved)

	invoice, err := repository.GetInvoiceByID(id)
	require.NoError(t, err)

	require.Len(t, invoice.Shortages, 1)
	assert.Equal(t, "PROD-1", invoice.Shortages[0].ProductCode)
	assert.Equal(t, 50, invoice.Shortages[0].Required)
	assert.Equal(t, 42, invoice.Shortages[0].Available)
}

func TestANewCloseAttemptClearsTheOldShortages(t *testing.T) {
	repository := newRepository(t)
	id := closingInvoice(t, repository)

	_, err := repository.RejectClose(id, "INSUFFICIENT_STOCK", shortagesOf("PROD-1"))
	require.NoError(t, err)

	require.NoError(t, repository.CloseInvoice(id, closeEventFor(t, id)))

	invoice, err := repository.GetInvoiceByID(id)
	require.NoError(t, err)

	assert.Empty(t, invoice.Shortages, "the new attempt starts without the old report")
	assert.Empty(t, invoice.FailureReason)
}

func TestRejectCloseOnAnInvoiceThatMovedAlreadyStoresNothing(t *testing.T) {
	repository := newRepository(t)
	id := closingInvoice(t, repository)

	_, err := repository.ConfirmClose(id)
	require.NoError(t, err)

	moved, err := repository.RejectClose(id, "INSUFFICIENT_STOCK", shortagesOf("PROD-1"))
	require.NoError(t, err)
	require.False(t, moved)

	invoice, err := repository.GetInvoiceByID(id)
	require.NoError(t, err)

	assert.Empty(t, invoice.Shortages, "a late result must not report over a closed invoice")
}

func TestDeleteInvoiceRemovesItsShortages(t *testing.T) {
	repository := newRepository(t)
	id := closingInvoice(t, repository)

	_, err := repository.RejectClose(id, "INSUFFICIENT_STOCK", shortagesOf("PROD-1"))
	require.NoError(t, err)
	require.NoError(t, repository.DeleteInvoice(id))

	var stored []model.InvoiceShortage
	require.NoError(t, testConnection.Find(&stored).Error)

	assert.Empty(t, stored)
}

func TestGetInvoicesComesWithTheMostRecentDocumentFirst(t *testing.T) {
	repository := newRepository(t)

	for _, document := range [][2]int{{2, 1}, {1, 57}, {1, 6}, {1, 7}} {
		_, err := repository.CreateInvoice(model.Invoice{
			Series: document[0], Number: document[1], Status: model.InvoiceStatusOpen,
		})
		require.NoError(t, err)
	}

	invoices, err := repository.GetInvoices()
	require.NoError(t, err)

	documents := make([]string, 0, len(invoices))
	for _, invoice := range invoices {
		documents = append(documents, invoice.FormattedNumber())
	}

	assert.Equal(t, []string{"002/000001", "001/000057", "001/000007", "001/000006"}, documents,
		"the newest number comes first, and 57 still outranks 7")
}
