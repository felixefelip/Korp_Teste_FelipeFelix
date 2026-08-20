package model

type Invoice struct {
	ID     int           `json:"id"     gorm:"primaryKey"`
	Number string        `json:"number" gorm:"type:varchar(30);not null"`
	Status string        `json:"status" gorm:"type:varchar(10);not null"`
	Items  []InvoiceItem `json:"items"  gorm:"foreignKey:InvoiceID"`
}

func (i Invoice) Total() float64 {
	total := 0.0

	for _, item := range i.Items {
		total += item.Total()
	}

	return total
}
