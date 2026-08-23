package db

import (
	"time"

	"inventory/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OutboxRepository struct {
	connection *gorm.DB
}

func NewOutboxRepository(connection *gorm.DB) *OutboxRepository {
	return &OutboxRepository{
		connection: connection,
	}
}

func (or *OutboxRepository) ClaimEvents(limit int, lease time.Duration) ([]model.OutboxEvent, error) {
	var events []model.OutboxEvent

	err := or.connection.Transaction(func(tx *gorm.DB) error {
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("published_at IS NULL AND next_attempt_at <= ?", time.Now()).
			Order("id").
			Limit(limit).
			Find(&events).Error
		if err != nil {
			return err
		}

		if len(events) == 0 {
			return nil
		}

		ids := make([]int, 0, len(events))
		for _, event := range events {
			ids = append(ids, event.ID)
		}

		return tx.
			Model(&model.OutboxEvent{}).
			Where("id IN ?", ids).
			Update("next_attempt_at", time.Now().Add(lease)).Error
	})
	if err != nil {
		return []model.OutboxEvent{}, err
	}

	return events, nil
}

func (or *OutboxRepository) MarkPublished(id int) error {
	result := or.connection.
		Model(&model.OutboxEvent{ID: id}).
		Updates(map[string]any{"published_at": time.Now(), "last_error": ""})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (or *OutboxRepository) MarkFailed(id int, cause string, nextAttemptAt time.Time) error {
	result := or.connection.
		Model(&model.OutboxEvent{ID: id}).
		Updates(map[string]any{
			"attempts":        gorm.Expr("attempts + 1"),
			"last_error":      cause,
			"next_attempt_at": nextAttemptAt,
		})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
