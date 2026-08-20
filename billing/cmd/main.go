package main

import (
	"billing/internal/infra/db"
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

	server := web.New(dbConnection)

	server.Run(":8001")
}
