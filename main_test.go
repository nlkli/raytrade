package main

import (
	"context"
	"fmt"
	"nlkli/raytrade/internal/broker"
	"nlkli/raytrade/internal/broker/bybit"
	"testing"
	"time"
)

func TestPos(t *testing.T) {
	client := bybit.NewClientFromEnv()
	b := bybit.NewBroker(client)

	pos, err := b.GetPosition(context.Background(), broker.Futures, "FARTCOINUSDT")
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%+v\n", pos)

	stream := b.CreatePrivateStream(nil)

	time.Sleep(time.Second * 7)

	sub, err := stream.SubscribePosition()

	for d := range sub.C {
		fmt.Printf("%+v", d)
	}
}
