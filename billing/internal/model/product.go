package model

type Product struct {
	ID          int     `gorm:"primaryKey"`
	InventoryID int     `gorm:"not null;uniqueIndex"`
	Code        string  `gorm:"type:varchar(30);not null"`
	Name        string  `gorm:"type:varchar(255);not null"`
	Unit        string  `gorm:"type:varchar(10);not null"`
	Price       float64 `gorm:"not null"`
}
