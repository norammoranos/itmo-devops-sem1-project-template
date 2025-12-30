package main

import (
	"fmt"
	"log"
	"os"

	"project_sem/internal/handlers/download"
	"project_sem/internal/handlers/upload"
	db "project_sem/internal/infrastructure/database"
	"project_sem/internal/infrastructure/server"
	"project_sem/usecases/prices"
)

const (
	databaseHostEnv     = "APP_DB_HOST"
	databaseNameEnv     = "APP_DB_NAME"
	databasePasswordEnv = "APP_DB_PASSWORD"
	databasePortEnv     = "APP_DB_PORT"
	databaseUserEnv     = "APP_DB_USER"
	portEnv             = "APP_PORT"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal(r)
		}
	}()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func dataSourceName() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv(databaseHostEnv),
		os.Getenv(databaseUserEnv),
		os.Getenv(databasePasswordEnv),
		os.Getenv(databaseNameEnv),
		os.Getenv(databasePortEnv),
	)
}

func run() error {
	database, err := db.New(dataSourceName())
	if err != nil {
		return fmt.Errorf("database.New: %w", err)
	}

	pricesUC := prices.New(database)
	srv := server.New(
		server.WithPort(os.Getenv(portEnv)),
		server.WithHandler("POST /api/v0/prices", upload.New(pricesUC)),
		server.WithHandler("GET /api/v0/prices", download.New(pricesUC)),
	)

	return srv.Run()
}
