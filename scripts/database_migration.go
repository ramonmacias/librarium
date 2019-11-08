package main

import (
	"github.com/ramonmacias/librarium/internal/app/interface/persistance/postgres"
)

func main() {
	db := postgres.NewClient("localhost", "5432", "ramon", "librarium_database", "ramon_postgres_pass").Connect().DB()
	db.AutoMigrate(&postgres.User{})
	db.AutoMigrate(&postgres.Book{})
}
