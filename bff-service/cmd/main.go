package main

import (
	"bff-service/internal/app"
	"context"
	"log"
)

func main() {
	ctx := context.Background()

	application := app.NewApp(ctx)

	if err := application.Run(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
