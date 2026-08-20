package main

import (
	"inventory/internal/infra/db"
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

	if err := dbConnection.AutoMigrate(&model.Product{}); err != nil {
		panic(err)
	}

	server := web.New(dbConnection)

	server.Run(":8000")
}
