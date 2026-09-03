package model_test

import (
	"testing"

	"billing/internal/model"

	"github.com/stretchr/testify/assert"
)

func itemWithICMS(quantity int, unitPrice, rate float64) model.InvoiceItem {
	item := model.InvoiceItem{Quantity: quantity, UnitPrice: unitPrice}

	return item.WithICMS(rate)
}

func itemWithIPI(quantity int, unitPrice, rate float64) model.InvoiceItem {
	item := model.InvoiceItem{Quantity: quantity, UnitPrice: unitPrice}

	return item.WithIPI(rate)
}

func TestWithICMSTakesTheBaseFromTheTotalOfTheItem(t *testing.T) {
	item := itemWithICMS(2, 30.99, 18)

	assert.Equal(t, 61.98, item.ICMSBase)
	assert.Equal(t, item.Total(), item.ICMSBase)
}

func TestWithICMSAppliesTheRateOverTheBase(t *testing.T) {
	item := itemWithICMS(2, 30.99, 18)

	assert.Equal(t, 18.0, item.ICMSRate)
	assert.Equal(t, 11.16, item.ICMSValue, "18% of 61.98 is 11.1564, rounded to the cent")
}

func TestWithICMSRoundsTheValueItStores(t *testing.T) {
	item := itemWithICMS(1, 19.9, 12)

	assert.Equal(t, 2.39, item.ICMSValue, "12% of 19.90 is 2.388, rounded up")
}

func TestWithICMSChargesNothingWithoutARate(t *testing.T) {
	item := itemWithICMS(3, 12.5, 0)

	assert.Zero(t, item.ICMSValue)
	assert.Equal(t, 37.5, item.ICMSBase, "the base exists even when the item is not taxed")
}

func TestInvoiceSumsTheICMSOfItsItems(t *testing.T) {
	invoice := model.Invoice{
		Items: []model.InvoiceItem{
			itemWithICMS(2, 30.99, 18),
			itemWithICMS(1, 19.9, 12),
		},
	}

	assert.Equal(t, 81.88, model.RoundMoney(invoice.ICMSBase()))
	assert.Equal(t, 13.55, model.RoundMoney(invoice.ICMSValue()))
}

func TestWithIPITakesTheBaseFromTheTotalOfTheItem(t *testing.T) {
	item := itemWithIPI(2, 30.99, 10)

	assert.Equal(t, 61.98, item.IPIBase)
	assert.Equal(t, 10.0, item.IPIRate)
	assert.Equal(t, 6.2, item.IPIValue, "10% of 61.98 is 6.198, rounded to the cent")
}

func TestWithIPIChargesNothingWithoutARate(t *testing.T) {
	item := itemWithIPI(3, 12.5, 0)

	assert.Zero(t, item.IPIValue)
	assert.Equal(t, 37.5, item.IPIBase)
}

func TestTheTwoTaxesLiveTogetherOnTheSameItem(t *testing.T) {
	item := model.InvoiceItem{Quantity: 2, UnitPrice: 30.99}.WithICMS(18).WithIPI(10)

	assert.Equal(t, 11.16, item.ICMSValue)
	assert.Equal(t, 6.2, item.IPIValue)
	assert.Equal(t, 61.98, item.Total(), "the line keeps showing the value of the products")
}

func TestInvoiceAddsTheIPIOnTopOfTheProducts(t *testing.T) {
	invoice := model.Invoice{
		Items: []model.InvoiceItem{
			model.InvoiceItem{Quantity: 2, UnitPrice: 30.99}.WithICMS(18).WithIPI(10),
			model.InvoiceItem{Quantity: 1, UnitPrice: 19.9}.WithICMS(12).WithIPI(5),
		},
	}

	assert.Equal(t, 81.88, model.RoundMoney(invoice.ProductsTotal()))
	assert.Equal(t, 7.2, model.RoundMoney(invoice.IPIValue()), "6.20 plus 1.00")
	assert.Equal(t, 89.08, model.RoundMoney(invoice.Total()),
		"the ipi is charged on top, so it belongs to the total of the invoice")
}

func TestInvoiceWithoutIPIKeepsTheTotalOfTheProducts(t *testing.T) {
	invoice := model.Invoice{
		Items: []model.InvoiceItem{itemWithICMS(2, 30.99, 18)},
	}

	assert.Equal(t, invoice.ProductsTotal(), invoice.Total())
}

func TestInvoiceTotalDoesNotAddTheICMSOnTop(t *testing.T) {
	invoice := model.Invoice{
		Items: []model.InvoiceItem{
			itemWithICMS(2, 30.99, 18),
			itemWithICMS(1, 19.9, 12),
		},
	}

	assert.Equal(t, 81.88, model.RoundMoney(invoice.Total()),
		"the icms is inside the price, so it is not summed into the total")
	assert.Equal(t, invoice.ProductsTotal(), invoice.Total())
}

func TestInvoiceWithoutItemsHasNoICMS(t *testing.T) {
	invoice := model.Invoice{}

	assert.Zero(t, invoice.ICMSBase())
	assert.Zero(t, invoice.ICMSValue())
	assert.Zero(t, invoice.IPIBase())
	assert.Zero(t, invoice.IPIValue())
	assert.Zero(t, invoice.Total())
}
