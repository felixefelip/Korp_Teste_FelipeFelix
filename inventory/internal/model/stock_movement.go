package model

import "errors"

const (
	MovementIn  = "in"
	MovementOut = "out"

	MovementOriginAdjustment = "adjustment"
	MovementOriginInvoice    = "invoice"
)

var ErrMovementFromInvoice = errors.New("stock movement originated from an invoice")

type StockMovement struct {
	ID            int    `gorm:"primaryKey"`
	ProductID     int    `gorm:"not null;index"`
	Type          string `gorm:"type:varchar(3);not null"`
	Origin        string `gorm:"type:varchar(20);not null"`
	Quantity      int    `gorm:"not null"`
	Confirmed     bool   `gorm:"not null;default:false"`
	InvoiceItemID *int   `gorm:"uniqueIndex"`
}

func (m StockMovement) FromInvoice() bool {
	return m.InvoiceItemID != nil
}
