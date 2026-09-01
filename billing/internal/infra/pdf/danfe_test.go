package pdf

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"billing/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func invoiceWith(items ...model.InvoiceItem) model.Invoice {
	return model.Invoice{
		ID:     1,
		Series: 1,
		Number: 42,
		Type:   model.InvoiceTypeOut,
		Status: model.InvoiceStatusClosed,
		Items:  items,
	}
}

func item(name string) model.InvoiceItem {
	return model.InvoiceItem{
		ProductCode: "ABC-1",
		ProductName: name,
		Unit:        "UN",
		Quantity:    3,
		UnitPrice:   12.5,
	}
}

func TestDanfeReturnsAPdfDocument(t *testing.T) {
	document, err := Danfe(invoiceWith(item("Cadeira")))

	require.NoError(t, err)
	assert.True(t, bytes.HasPrefix(document, []byte("%PDF-")), "the response has to be a pdf")
	assert.Greater(t, len(document), 500, "an empty pdf would mean nothing was drawn")
}

func TestDanfeRendersAnInvoiceWithoutItems(t *testing.T) {
	document, err := Danfe(invoiceWith())

	require.NoError(t, err)
	assert.True(t, bytes.HasPrefix(document, []byte("%PDF-")))
}

func TestDanfeFitsAFewItemsInASinglePage(t *testing.T) {
	assert.Equal(t, 1, render(invoiceWith(item("Cadeira"), item("Mesa"))).pdf.PageNo())
}

func TestDanfeBreaksALongInvoiceIntoMorePages(t *testing.T) {
	items := make([]model.InvoiceItem, 0, 120)

	for index := range 120 {
		items = append(items, item(fmt.Sprintf("Produto %d", index)))
	}

	assert.Greater(t, render(invoiceWith(items...)).pdf.PageNo(), 1,
		"120 items do not fit in one page")
}

func TestDanfeTruncatesADescriptionWiderThanItsColumn(t *testing.T) {
	document := newDocument()
	document.pdf.SetFont("Helvetica", "", 8)

	name := strings.Repeat("Cadeira giratória de escritório ", 5)
	fitted := document.fit(name, itemColumns[1].width)

	assert.Less(t, len(fitted), len(name))
	assert.True(t, strings.HasSuffix(fitted, "..."))
	assert.LessOrEqual(t,
		document.pdf.GetStringWidth(document.translate(fitted)),
		itemColumns[1].width-2,
	)
}

func TestDanfeKeepsAShortDescriptionUntouched(t *testing.T) {
	document := newDocument()
	document.pdf.SetFont("Helvetica", "", 8)

	assert.Equal(t, "Cadeira", document.fit("Cadeira", itemColumns[1].width))
}

func TestMoneyUsesTheBrazilianFormat(t *testing.T) {
	cases := map[float64]string{
		0:          "0,00",
		12.5:       "12,50",
		999.99:     "999,99",
		1234.5:     "1.234,50",
		1234567.89: "1.234.567,89",
		-1234.5:    "-1.234,50",
	}

	for value, expected := range cases {
		assert.Equal(t, expected, money(value), "%f", value)
	}
}

func TestOperationLabelFollowsTheInvoiceType(t *testing.T) {
	assert.Equal(t, "SAÍDA", operationLabel(model.Invoice{Type: model.InvoiceTypeOut}))
	assert.Equal(t, "ENTRADA", operationLabel(model.Invoice{Type: model.InvoiceTypeIn}))
}

func TestDanfeItemColumnsFillTheContentWidth(t *testing.T) {
	width := 0.0

	for _, column := range itemColumns {
		width += column.width
	}

	assert.Equal(t, contentWidth, width, "the row has to fill the page, no more and no less")
}

func TestPercentFormatsTheRateOfTheItem(t *testing.T) {
	assert.Equal(t, "18,00%", percent(18))
	assert.Equal(t, "0,00%", percent(0))
	assert.Equal(t, "7,50%", percent(7.5))
}
