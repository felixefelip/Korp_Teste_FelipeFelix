package model

type Invoice struct {
	ID     int    `json:"id"     gorm:"primaryKey"`
	Number string `json:"number" gorm:"type:varchar(30);not null"`
	Status string `json:"status" gorm:"type:varchar(10);not null"`
}
