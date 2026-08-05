package main

import (
	"context"
	"log"

	"mission-note/internal/bootstrap"
)

func main() {
	ctx := context.Background()

	app, err := bootstrap.Init(ctx)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}
	defer app.Close()

	if err := app.Router.Run(":" + app.Port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
