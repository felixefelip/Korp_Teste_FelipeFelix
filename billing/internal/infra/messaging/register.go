package messaging

import (
	"billing/internal/infra/db"

	"gorm.io/gorm"
)

func Register(connection *gorm.DB) {
	NewRelay(db.NewOutboxRepository(connection), BillingExchange).Start()
}
