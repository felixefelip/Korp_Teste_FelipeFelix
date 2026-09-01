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

func TestInvoiceTotalDoesNotAddTheICMSOnTop(t *testing.T) {
	invoice := model.Invoice{
		Items: []model.InvoiceItem{
			itemWithICMS(2, 30.99, 18),
			itemWithICMS(1, 19.9, 12),
		},
	}

	assert.Equal(t, 81.88, model.RoundMoney(invoice.Total()),
		"the icms is inside the price, so it is not summed into the total")
}

func TestInvoiceWithoutItemsHasNoICMS(t *testing.T) {
	invoice := model.Invoice{}

	assert.Zero(t, invoice.ICMSBase())
	assert.Zero(t, invoice.ICMSValue())
}
