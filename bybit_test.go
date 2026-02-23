package main

import (
	"context"
	"fmt"
	"nlkli/raytrade/internal/broker"
	"nlkli/raytrade/internal/broker/bybit"
	"testing"
)

func TestPos(t *testing.T) {
	client := bybit.NewClientFromEnv()
	b := bybit.NewBroker(client)

	pos, err := b.GetPosition(context.Background(), broker.Futures, "FARTCOINUSDT")
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%+v\n", pos)
}
