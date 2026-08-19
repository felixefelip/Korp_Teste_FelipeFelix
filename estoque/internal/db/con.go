package db

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func Env() string {
	return env("GO_ENV", "development")
}

func DatabaseName() string {
	return "go_api_" + Env()
}

func dsn(dbname string) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		env("DB_HOST", "go_db"),
		env("DB_PORT", "5432"),
		env("DB_USER", "postgres"),
		env("DB_PASSWORD", "1234"),
		dbname,
	)
}

func open(dbname string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn(dbname)), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
}

func ConnectDB() (*gorm.DB, error) {
	dbname := DatabaseName()

	db, err := open(dbname)
	if err != nil {
		return nil, err
	}

	fmt.Println("Successfully connected to " + dbname)

	return db, nil
}

func EnsureDatabase() error {
	dbname := DatabaseName()

	admin, err := open(env("DB_ADMIN_NAME", "postgres"))
	if err != nil {
		return err
	}

	sqlDB, err := admin.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	var exists bool
	err = admin.Raw("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = ?)", dbname).
		Scan(&exists).Error
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	if err := admin.Exec(fmt.Sprintf("CREATE DATABASE %q", dbname)).Error; err != nil {
		return err
	}

	fmt.Println("Created database " + dbname)

	return nil
}
