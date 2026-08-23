package model

import (
	"errors"
	"fmt"
)

const (
	InvoiceTypeIn  = "IN"
	InvoiceTypeOut = "OUT"

	InvoiceStatusOpen      = "OPEN"
	InvoiceStatusClosing   = "CLOSING"
	InvoiceStatusClosed    = "CLOSED"
	InvoiceStatusReopening = "REOPENING"
)

var (
	ErrInvoiceDuplicated    = errors.New("invoice number already used in this series")
	ErrInvoiceClosed        = errors.New("closed invoice")
	ErrInvoiceOpen          = errors.New("open invoice")
	ErrInvoiceProcessing    = errors.New("invoice being processed")
	ErrInvoiceNotProcessing = errors.New("invoice is not being processed")
)

type Invoice struct {
	ID     int           `json:"id"     gorm:"primaryKey"`
	Series int           `json:"series" gorm:"not null;uniqueIndex:idx_invoice_document"`
	Number int           `json:"number" gorm:"not null;uniqueIndex:idx_invoice_document"`
	Type   string        `json:"type"   gorm:"type:varchar(3);not null;default:'OUT'"`
	Status string        `json:"status" gorm:"type:varchar(10);not null"`
	Items  []InvoiceItem `json:"items"  gorm:"foreignKey:InvoiceID"`

	FailureReason string            `json:"failureReason" gorm:"type:varchar(30);not null;default:''"`
	Shortages     []InvoiceShortage `json:"shortages"     gorm:"foreignKey:InvoiceID"`
}

func (i Invoice) FormattedNumber() string {
	return fmt.Sprintf("%03d/%06d", i.Series, i.Number)
}

func (i Invoice) Closed() bool {
	return i.Status == InvoiceStatusClosed
}

func (i Invoice) Processing() bool {
	return i.Status == InvoiceStatusClosing || i.Status == InvoiceStatusReopening
}

func (i Invoice) Editable() bool {
	return i.Status == InvoiceStatusOpen
}

func (i Invoice) MovesStockOut() bool {
	return i.Type == InvoiceTypeOut
}

func (i Invoice) Total() float64 {
	total := 0.0

	for _, item := range i.Items {
		total += item.Total()
	}

	return total
}
