package main

import (
	"companies-service/internal/app"
	"context"
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
