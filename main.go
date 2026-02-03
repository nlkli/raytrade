package main

import (
	"context"
	"nlkli/raytrade/internal/app2"
)

func main() {
	app2.Run(context.Background(), "config.json")
}
