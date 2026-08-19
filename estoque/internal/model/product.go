package model

type Product struct {
	ID    int     `json:"id"    gorm:"primaryKey"`
	Name  string  `json:"name"  gorm:"type:varchar(255);not null"`
	Price float64 `json:"price" gorm:"type:numeric(10,2);not null"`
	Stock int     `json:"stock" gorm:"type:integer;not null;default:0"`
}
