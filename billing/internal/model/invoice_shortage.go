package model

type InvoiceShortage struct {
	ID          int    `gorm:"primaryKey"`
	InvoiceID   int    `gorm:"not null;index"`
	InventoryID int    `gorm:"not null"`
	ProductCode string `gorm:"type:varchar(30);not null"`
	ProductName string `gorm:"type:varchar(255);not null"`
	Required    int    `gorm:"not null"`
	Available   int    `gorm:"not null"`
}
