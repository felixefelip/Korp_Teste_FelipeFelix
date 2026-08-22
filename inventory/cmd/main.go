package main

import (
	"inventory/internal/infra/db"
	"inventory/internal/infra/messaging"
	invoicemq "inventory/internal/infra/messaging/invoice"
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

	if err := dbConnection.AutoMigrate(&model.Product{}, &model.StockMovement{}); err != nil {
		panic(err)
	}

	handler := invoicemq.NewHandler()

	messaging.NewConsumer(messaging.InvoiceRequestsQueue, handler.HandleCloseRequested).Start()

	server := web.New(dbConnection)

	server.Run(":8000")
}
