package main

import (
	"context"
	"flag"
	"nlkli/raytrade/internal/app"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	configPath := flag.String("c", "config.json", "config path")

	flag.Parse()

	app.Run(ctx, *configPath)
}
