package model

// Code nao tem indice unico de proposito: o negocio permite codigos repetidos.
type Product struct {
	ID    int     `json:"id"    gorm:"primaryKey"`
	Code  string  `json:"code"  gorm:"type:varchar(30);not null"`
	Name  string  `json:"name"  gorm:"type:varchar(255);not null"`
	Unit  string  `json:"unit"  gorm:"type:varchar(10);not null"`
	Price float64 `json:"price" gorm:"type:numeric(10,2);not null"`
	Stock int     `json:"stock" gorm:"type:integer;not null;default:0"`
}
