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

	order, _, err := b.GetOpenOrder(context.Background(), broker.Futures, "FARTCOINUSDT")
	if err != nil {
		fmt.Println("----", err.Error())
	}

	fmt.Printf("%+v\n", order)

	stream := b.CreatePrivateStream(nil)

	time.Sleep(time.Second * 7)

	sub, err := stream.SubscribeOrder()

	for d := range sub.C {
		fmt.Printf("%+v\n", d)
	}
}
