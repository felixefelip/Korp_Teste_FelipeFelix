package model

const (
	FirstInvoiceSeries = 1
	FirstInvoiceNumber = 1
	MaxInvoiceNumber   = 999999
)

type InvoiceDocument struct {
	Series int
	Number int
}

func (d InvoiceDocument) Suggested() bool {
	return d.Number > 0
}

func ResolveInvoiceSeries(requested, last int) int {
	if requested > 0 {
		return requested
	}

	if last > 0 {
		return last
	}

	return FirstInvoiceSeries
}

func NextInvoiceDocument(series, lastNumber int) InvoiceDocument {
	if lastNumber <= 0 {
		return InvoiceDocument{Series: series, Number: FirstInvoiceNumber}
	}

	if lastNumber >= MaxInvoiceNumber {
		return InvoiceDocument{Series: series}
	}

	return InvoiceDocument{Series: series, Number: lastNumber + 1}
}
