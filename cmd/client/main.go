package main

import (
	"context"
	"log"
	"os"

	"keeper/config"
	"keeper/internal/app"
)

func main() {
	log.SetOutput(os.Stdout)
	settings, err := config.NewConfig("client.yml")
	if err != nil {
		log.Fatal(err)
	}

	server, err := app.NewClient(context.Background(), settings)
	if err != nil {
		log.Fatal(err)
	}
	server.Run()
}
