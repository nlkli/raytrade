package main

import (
	"context"
	"nlkli/raytrade/internal/app"
)

func main() {
	app.Run(context.Background(), "config.json")
}
