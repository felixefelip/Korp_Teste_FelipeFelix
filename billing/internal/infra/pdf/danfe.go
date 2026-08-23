package pdf

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"billing/internal/model"

	"github.com/go-pdf/fpdf"
)

const (
	pageWidth    = 210.0
	pageHeight   = 297.0
	margin       = 10.0
	contentWidth = pageWidth - 2*margin
	bottomLimit  = pageHeight - 20.0
	rowHeight    = 6.0
)

var itemColumns = []struct {
	title string
	width float64
	align string
}{
	{"CÓDIGO", 25, "L"},
	{"DESCRIÇÃO", 79, "L"},
	{"UN", 14, "C"},
	{"QTD", 18, "R"},
	{"VL. UNIT.", 27, "R"},
	{"VL. TOTAL", 27, "R"},
}

func totalWidth() float64 {
	return itemColumns[len(itemColumns)-1].width
}

func Danfe(invoice model.Invoice) ([]byte, error) {
	var buffer bytes.Buffer

	if err := render(invoice).pdf.Output(&buffer); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func render(invoice model.Invoice) *document {
	document := newDocument()

	document.header()
	document.identification(invoice)
	document.items(invoice.Items)
	document.total(invoice.Total())

	return document
}

type document struct {
	pdf       *fpdf.Fpdf
	translate func(string) string
}

func newDocument() *document {
	pdf := fpdf.New("P", "mm", "A4", "")
	translate := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	pdf.SetMargins(margin, margin, margin)
	pdf.SetAutoPageBreak(false, margin)
	pdf.AliasNbPages("{nb}")

	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Helvetica", "", 7)
		pdf.CellFormat(
			contentWidth, 5,
			translate(fmt.Sprintf("Página %d de {nb}", pdf.PageNo())),
			"", 0, "C", false, 0, "",
		)
	})

	pdf.AddPage()

	return &document{pdf: pdf, translate: translate}
}

func (d *document) cell(width, height float64, content, border, align string, fill bool) {
	d.pdf.CellFormat(width, height, d.translate(content), border, 0, align, fill, 0, "")
}

func (d *document) header() {
	d.pdf.SetFont("Helvetica", "B", 14)
	d.cell(contentWidth, 9, "DANFE SIMPLIFICADA", "LTR", "C", false)
	d.pdf.Ln(-1)

	d.pdf.SetFont("Helvetica", "", 8)
	d.cell(
		contentWidth, 5,
		"Documento auxiliar da nota fiscal - representação simplificada, sem valor fiscal",
		"LBR", "C", false,
	)
	d.pdf.Ln(-1)
	d.pdf.Ln(3)
}

func (d *document) identification(invoice model.Invoice) {
	widths := []float64{70, 60, 60}
	labels := []string{"NÚMERO", "OPERAÇÃO", "QUANTIDADE DE ITENS"}
	values := []string{
		invoice.FormattedNumber(),
		operationLabel(invoice),
		strconv.Itoa(len(invoice.Items)),
	}

	d.pdf.SetFont("Helvetica", "", 6)

	for index, label := range labels {
		d.cell(widths[index], 4, label, "LTR", "L", false)
	}

	d.pdf.Ln(-1)
	d.pdf.SetFont("Helvetica", "B", 10)

	for index, value := range values {
		d.cell(widths[index], 6, value, "LBR", "L", false)
	}

	d.pdf.Ln(-1)
	d.pdf.Ln(3)
}

func (d *document) items(items []model.InvoiceItem) {
	d.itemsHeader()

	if len(items) == 0 {
		d.cell(contentWidth, rowHeight, "Nenhum item.", "1", "C", false)
		d.pdf.Ln(-1)

		return
	}

	for _, item := range items {
		if d.pdf.GetY()+rowHeight > bottomLimit {
			d.pdf.AddPage()
			d.itemsHeader()
		}

		values := []string{
			item.ProductCode,
			d.fit(item.ProductName, itemColumns[1].width),
			item.Unit,
			strconv.Itoa(item.Quantity),
			money(item.UnitPrice),
			money(item.Total()),
		}

		for index, column := range itemColumns {
			d.cell(column.width, rowHeight, values[index], "1", column.align, false)
		}

		d.pdf.Ln(-1)
	}
}

func (d *document) itemsHeader() {
	d.pdf.SetFont("Helvetica", "B", 7)
	d.pdf.SetFillColor(230, 230, 230)

	for _, column := range itemColumns {
		d.cell(column.width, rowHeight, column.title, "1", "C", true)
	}

	d.pdf.Ln(-1)
	d.pdf.SetFont("Helvetica", "", 8)
}

func (d *document) total(total float64) {
	if d.pdf.GetY()+8 > bottomLimit {
		d.pdf.AddPage()
	}

	d.pdf.SetFont("Helvetica", "B", 9)
	d.cell(contentWidth-totalWidth(), 8, "VALOR TOTAL DA NOTA", "1", "R", false)
	d.cell(totalWidth(), 8, money(total), "1", "R", false)
	d.pdf.Ln(-1)
}

func (d *document) fit(content string, width float64) string {
	limit := width - 2

	if d.pdf.GetStringWidth(d.translate(content)) <= limit {
		return content
	}

	runes := []rune(content)

	for len(runes) > 1 {
		runes = runes[:len(runes)-1]

		candidate := string(runes) + "..."
		if d.pdf.GetStringWidth(d.translate(candidate)) <= limit {
			return candidate
		}
	}

	return content
}

func operationLabel(invoice model.Invoice) string {
	if invoice.MovesStockOut() {
		return "SAÍDA"
	}

	return "ENTRADA"
}

func money(value float64) string {
	formatted := strconv.FormatFloat(value, 'f', 2, 64)
	parts := strings.SplitN(formatted, ".", 2)

	integer, sign := parts[0], ""
	if strings.HasPrefix(integer, "-") {
		sign, integer = "-", integer[1:]
	}

	groups := []string{}

	for len(integer) > 3 {
		groups = append([]string{integer[len(integer)-3:]}, groups...)
		integer = integer[:len(integer)-3]
	}

	groups = append([]string{integer}, groups...)

	return sign + strings.Join(groups, ".") + "," + parts[1]
}
