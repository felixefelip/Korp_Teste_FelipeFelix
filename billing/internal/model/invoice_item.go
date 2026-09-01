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
	ICMSRate    float64 `gorm:"not null;default:0"`
	ICMSBase    float64 `gorm:"not null;default:0"`
	ICMSValue   float64 `gorm:"not null;default:0"`
}

func (i InvoiceItem) Total() float64 {
	return float64(i.Quantity) * i.UnitPrice
}

func (i InvoiceItem) WithICMS(rate float64) InvoiceItem {
	i.ICMSRate = rate
	i.ICMSBase = RoundMoney(i.Total())
	i.ICMSValue = RoundMoney(i.ICMSBase * rate / 100)

	return i
}
