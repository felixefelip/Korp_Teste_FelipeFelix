package model

type InvoiceItem struct {
	ID          int     `gorm:"primaryKey"`
	InvoiceID   int     `gorm:"not null;index"`
	ProductID   int     `gorm:"not null;index"`
	Product     Product `gorm:"foreignKey:ProductID"`
	ProductCode string  `gorm:"type:varchar(30);not null"`
	ProductName string  `gorm:"type:varchar(255);not null"`
	Unit        string  `gorm:"type:varchar(10);not null"`
	Quantity    int     `gorm:"not null"`
	UnitPrice   float64 `gorm:"not null"`
}

func (i InvoiceItem) Total() float64 {
	return float64(i.Quantity) * i.UnitPrice
}
