package main

import (
	"context"
	"log"
	"os"

	"github.com/syols/keeper/config"
	"github.com/syols/keeper/internal/app"
)

func main() {
	log.SetOutput(os.Stdout)
	settings, err := config.NewConfig("client.yml")
	if err != nil {
		log.Fatal(err.Error())
	}

	client, err := app.NewClient(settings)
	if err == nil {
		if err := client.Run(context.Background()); err != nil {
			log.Fatal(err.Error())
		}
	}
}
