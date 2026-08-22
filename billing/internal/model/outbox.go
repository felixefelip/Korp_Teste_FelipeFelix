package model

import "time"

const OutboxAggregateInvoice = "invoice"

type OutboxEvent struct {
	ID            int       `gorm:"primaryKey"`
	EventID       string    `gorm:"type:uuid;not null;uniqueIndex"`
	AggregateType string    `gorm:"type:varchar(30);not null"`
	AggregateID   int       `gorm:"not null;index"`
	RoutingKey    string    `gorm:"type:varchar(60);not null"`
	Payload       []byte    `gorm:"type:jsonb;not null"`
	CreatedAt     time.Time `gorm:"not null"`
	PublishedAt   *time.Time
	Attempts      int       `gorm:"not null;default:0"`
	NextAttemptAt time.Time `gorm:"not null;index"`
	LastError     string    `gorm:"type:text"`
}

func (e OutboxEvent) Published() bool {
	return e.PublishedAt != nil
}

type OutboxRepository interface {
	ClaimEvents(limit int, lease time.Duration) ([]OutboxEvent, error)
	MarkPublished(id int) error
	MarkFailed(id int, cause string, nextAttemptAt time.Time) error
}
