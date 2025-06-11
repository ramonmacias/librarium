package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"librarium/internal/app"
	"librarium/internal/postgres"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file")
	}

	application, err := app.NewLibrariumApplication(
		app.WithDatabaseSource(
			&postgres.DataSource{
				UserName: os.Getenv("DB_USER"),
				Password: os.Getenv("DB_PASSWORD"),
				Port:     os.Getenv("DB_PORT"),
				DBName:   os.Getenv("DB_NAME"),
				SSLMode:  os.Getenv("DB_SSL_MODE"),
			},
		),
		app.WithServerAddress(os.Getenv("ADDRESS")),
	)
	if err != nil {
		panic(err)
	}

	application.Run()
}
