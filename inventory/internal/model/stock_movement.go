package model

import "errors"

const (
	MovementIn  = "in"
	MovementOut = "out"

	MovementOriginAdjustment = "adjustment"
	MovementOriginInvoice    = "invoice"

	InvoiceTypeIn  = "IN"
	InvoiceTypeOut = "OUT"
)

var ErrMovementFromInvoice = errors.New("stock movement originated from an invoice")

type StockMovement struct {
	ID        int    `gorm:"primaryKey"`
	ProductID int    `gorm:"not null;index"`
	Type      string `gorm:"type:varchar(3);not null"`
	Origin    string `gorm:"type:varchar(20);not null"`
	Quantity  int    `gorm:"not null"`
	Confirmed bool   `gorm:"not null;default:false"`

	BillingInvoiceItemID *int   `gorm:"uniqueIndex"`
	BillingInvoiceID     *int   `gorm:"index"`
	InvoiceNumber        string `gorm:"type:varchar(30)"`
	CloseEventID         string `gorm:"type:varchar(36);index"`
}

func (m StockMovement) FromInvoice() bool {
	return m.BillingInvoiceItemID != nil
}

func (m StockMovement) WithoutInvoice() StockMovement {
	m.BillingInvoiceItemID = nil
	m.BillingInvoiceID = nil
	m.InvoiceNumber = ""
	m.CloseEventID = ""

	return m
}
