package main

import (
	"librarium/internal/app"
)

func main() {
	application, err := app.NewLibrariumApplication()
	if err != nil {
		panic(err)
	}

	application.Run()
}
