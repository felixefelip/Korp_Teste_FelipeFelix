package db_test

import (
	"testing"

	"billing/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestGetInvoicesReturnsEverythingStored(t *testing.T) {
	repository := newRepository(t)

	_, err := repository.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "OPEN"})
	require.NoError(t, err)

	_, err = repository.CreateInvoice(model.Invoice{Number: "NF-0002", Status: "CLOSED"})
	require.NoError(t, err)

	invoices, err := repository.GetInvoices()
	require.NoError(t, err)

	require.Len(t, invoices, 2)
	assert.ElementsMatch(t,
		[]string{"NF-0001", "NF-0002"},
		[]string{invoices[0].Number, invoices[1].Number},
	)
}

func TestGetInvoicesReturnsTheStoredFields(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "CLOSED"})
	require.NoError(t, err)

	invoices, err := repository.GetInvoices()
	require.NoError(t, err)

	require.Len(t, invoices, 1)
	assert.Equal(t, model.Invoice{
		ID: id, Number: "NF-0001", Status: "CLOSED", Items: []model.InvoiceItem{},
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

	id, err := repository.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "OPEN"})
	require.NoError(t, err)

	invoice, err := repository.GetInvoiceByID(id)

	require.NoError(t, err)
	assert.Equal(t, model.Invoice{
		ID: id, Number: "NF-0001", Status: "OPEN", Items: []model.InvoiceItem{},
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

	id, err := repository.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "OPEN"})
	require.NoError(t, err)
	assert.NotZero(t, id, "the database should have generated an id")

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)

	assert.Equal(t, "NF-0001", saved.Number)
	assert.Equal(t, "OPEN", saved.Status)
}

func TestCreateInvoiceKeepsTheClosedStatus(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(model.Invoice{Number: "NF-0002", Status: "CLOSED"})
	require.NoError(t, err)

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)

	assert.Equal(t, "CLOSED", saved.Status)
}

func TestCreateInvoiceGeneratesSequentialIDs(t *testing.T) {
	repository := newRepository(t)

	first, err := repository.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "OPEN"})
	require.NoError(t, err)

	second, err := repository.CreateInvoice(model.Invoice{Number: "NF-0002", Status: "OPEN"})
	require.NoError(t, err)

	assert.Greater(t, second, first, "each invoice should get its own id")
}

func TestCreateInvoiceAcceptsADuplicateNumber(t *testing.T) {
	repository := newRepository(t)

	_, err := repository.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "OPEN"})
	require.NoError(t, err)

	_, err = repository.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "OPEN"})
	require.NoError(t, err, "a duplicate number must be accepted")
}

func TestUpdateInvoiceChangesEveryField(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "OPEN"})
	require.NoError(t, err)

	err = repository.UpdateInvoice(model.Invoice{ID: id, Number: "NF-0002", Status: "CLOSED"})
	require.NoError(t, err)

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)

	assert.Equal(t, model.Invoice{ID: id, Number: "NF-0002", Status: "CLOSED"}, saved)
}

func TestUpdateInvoiceWhenMissingReturnsErrRecordNotFound(t *testing.T) {
	repository := newRepository(t)

	err := repository.UpdateInvoice(model.Invoice{ID: 9999, Number: "NF-0001", Status: "OPEN"})

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUpdateInvoiceLeavesTheOtherInvoicesAlone(t *testing.T) {
	repository := newRepository(t)

	first, err := repository.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "OPEN"})
	require.NoError(t, err)

	second, err := repository.CreateInvoice(model.Invoice{Number: "NF-0002", Status: "OPEN"})
	require.NoError(t, err)

	require.NoError(t, repository.UpdateInvoice(model.Invoice{ID: first, Number: "NF-0001", Status: "CLOSED"}))

	var untouched model.Invoice
	require.NoError(t, testConnection.First(&untouched, second).Error)

	assert.Equal(t, "NF-0002", untouched.Number)
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

func invoiceWithItems(items ...model.InvoiceItem) model.Invoice {
	return model.Invoice{Number: "NF-0001", Status: "OPEN", Items: items}
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

func TestCreateInvoiceRefreshesTheProductAlreadyRegistered(t *testing.T) {
	repository := newRepository(t)

	_, err := repository.CreateInvoice(invoiceWithItems(itemOf(11, "PRD-0001", "Camiseta", 2, 30.99)))
	require.NoError(t, err)

	_, err = repository.CreateInvoice(invoiceWithItems(itemOf(11, "PRD-0001", "Camiseta polo", 1, 45.5)))
	require.NoError(t, err)

	var product model.Product
	require.NoError(t, testConnection.Where("inventory_id = ?", 11).First(&product).Error)

	assert.Equal(t, "Camiseta polo", product.Name)
	assert.Equal(t, 45.5, product.Price)
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

	_, err := repository.CreateInvoice(invoiceWithItems(itemOf(11, "PRD-0001", "Camiseta", 2, 30.99)))
	require.NoError(t, err)

	_, err = repository.CreateInvoice(model.Invoice{Number: "NF-0002", Status: "OPEN"})
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
		Number: "NF-0001",
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
		ID: id, Number: "NF-0001", Status: "OPEN", Items: []model.InvoiceItem{},
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

	err = repository.UpdateInvoice(model.Invoice{ID: id, Number: "NF-0001", Status: "CLOSED"})
	require.NoError(t, err)

	invoice, err := repository.GetInvoiceByID(id)
	require.NoError(t, err)

	assert.Equal(t, "CLOSED", invoice.Status)
	require.Len(t, invoice.Items, 1, "an update that says nothing about the items must not erase them")
	assert.Equal(t, "Camiseta", invoice.Items[0].ProductName)
}

func TestUpdateInvoiceKeepsTheProductAlreadyRegistered(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(invoiceWithItems(itemOf(11, "PRD-0001", "Camiseta", 2, 30.99)))
	require.NoError(t, err)

	err = repository.UpdateInvoice(model.Invoice{
		ID:     id,
		Number: "NF-0001",
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
		Number: "NF-0001",
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
		ID: first, Number: "NF-0001", Status: "CLOSED", Items: []model.InvoiceItem{},
	})
	require.NoError(t, err)

	untouched, err := repository.GetInvoiceByID(second)
	require.NoError(t, err)

	require.Len(t, untouched.Items, 1)
	assert.Equal(t, "Caneca", untouched.Items[0].ProductName)
}
