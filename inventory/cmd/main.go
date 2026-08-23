package main

import (
	"inventory/internal/infra/db"
	"inventory/internal/infra/messaging"
	"inventory/internal/infra/web"
	"inventory/internal/model"
)

func main() {
	if err := db.EnsureDatabase(); err != nil {
		panic(err)
	}

	dbConnection, err := db.ConnectDB()
	if err != nil {
		panic(err)
	}

	if err := dbConnection.AutoMigrate(&model.Product{}, &model.StockMovement{}, &model.OutboxEvent{}); err != nil {
		panic(err)
	}

	messaging.Register(dbConnection)

	server := web.New(dbConnection)

	server.Run(":8000")
}
