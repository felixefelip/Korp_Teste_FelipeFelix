package model

import "errors"

const (
	InvoiceTypeIn  = "IN"
	InvoiceTypeOut = "OUT"

	InvoiceStatusOpen   = "OPEN"
	InvoiceStatusClosed = "CLOSED"
)

var ErrInvoiceClosed = errors.New("closed invoice")

type Invoice struct {
	ID     int           `json:"id"     gorm:"primaryKey"`
	Number string        `json:"number" gorm:"type:varchar(30);not null"`
	Type   string        `json:"type"   gorm:"type:varchar(3);not null;default:'OUT'"`
	Status string        `json:"status" gorm:"type:varchar(10);not null"`
	Items  []InvoiceItem `json:"items"  gorm:"foreignKey:InvoiceID"`
}

func (i Invoice) Closed() bool {
	return i.Status == InvoiceStatusClosed
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
