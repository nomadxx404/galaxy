package main

import (
	"context"
	"files-service/internal/app"
	"log"
)

func main() {
	ctx := context.Background()

	application, cleanup := app.NewApp(ctx)
	defer cleanup()

	if err := application.Run(ctx); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
