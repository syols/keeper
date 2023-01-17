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
	settings, err := config.NewConfig("server.yml")
	if err != nil {
		log.Fatal(err)
	}

	server, err := app.NewServer(context.Background(), settings)
	if err != nil {
		log.Fatal(err)
	}
	server.Run()
}
