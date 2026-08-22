package main

import (
	"billing/internal/infra/db"
	"billing/internal/infra/messaging"
	invoicemq "billing/internal/infra/messaging/invoice"
	"billing/internal/infra/web"
	"billing/internal/model"
)

func main() {
	if err := db.EnsureDatabase(); err != nil {
		panic(err)
	}

	dbConnection, err := db.ConnectDB()
	if err != nil {
		panic(err)
	}

	if err := dbConnection.AutoMigrate(&model.Product{}, &model.Invoice{}, &model.InvoiceItem{}); err != nil {
		panic(err)
	}

	broker, err := messaging.Connect()
	if err != nil {
		panic(err)
	}
	defer broker.Close()

	server := web.New(dbConnection, invoicemq.NewPublisher(broker.Channel()))

	server.Run(":8001")
}
