package main

import (
	"context"
	"encoding/json"
	"fmt"
	"nlkli/raytrade/internal/broker"
	"nlkli/raytrade/internal/broker/bybit"
	"nlkli/raytrade/internal/broker/bybit/models"
	"nlkli/raytrade/internal/cdl"
	"testing"
	"time"
)

func TestBybitGetKline(t *testing.T) {
	client := bybit.NewClientFromEnv(context.Background())
	kline, err := client.GetKline(models.CategoryLinear, "BTCUSDT", models.Interval15Min, nil, nil, nil)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%+v\n", kline)
}

func TestWsConnV2(t *testing.T) {
	tx := make(chan []byte)

	client := bybit.NewClientFromEnv(context.Background())

	stream := client.CreatePublicStreamV2(models.CategoryLinear, tx)

	sub, err := stream.Subscribe("kline.5.BTCUSDT")
	if err != nil {
		t.Error(err)
	}

	go func() {
		time.Sleep(time.Second * 10)
		stream.Unsubscribe("kline.5.BTCUSDT")
		println("Send unsub")
	}()

	for d := range sub {
		println(string(d.Data))
	}

	println("Is unsub")
	time.Sleep(time.Second * 30)
}

func TestWsConnect(t *testing.T) {
}

func pp(v any) {
	s, _ := json.MarshalIndent(v, "", "    ")
	fmt.Println(string(s))
}

func TestGetCandles(t *testing.T) {
	client := bybit.NewClientFromEnv(context.Background())
	b := bybit.NewBroker(client)
	candles, err := b.GetCandles(broker.Futures, "BTCUSDT", cdl.M15, 5, nil, nil)
	if err != nil {
		t.Error(err)
	}
	pp(candles)
	r, err := b.GetCandles(broker.Futures, "BTCUSDT", cdl.M15, 5, nil, nil)
	if err != nil {
		t.Error(err)
	}
	pp(r)
	esc, err := b.ExtendEndCandles(candles[:len(candles)-2], broker.Futures, "BTCUSDT", cdl.M15, 2)
	if err != nil {
		t.Error(err)
	}
	pp(esc)
}

func TestCandlesStream(t *testing.T) {
	client := bybit.NewClientFromEnv(context.Background())
	b := bybit.NewBroker(client)

	stream := b.CreateStream(broker.Futures)
	sub, err := stream.SubscribeCandleStream("BTCUSDT", cdl.M1)

	if err != nil {
		t.Error(err)
	}

	for d := range sub.C {
		fmt.Printf("%+v\n", d)
	}
}

func TestGetOrderBook(t *testing.T) {
	client := bybit.NewClientFromEnv(context.Background())
	b := bybit.NewBroker(client)

	ob, err := b.GetOrderBook(broker.Futures, "BTCUSDT", 10)

	if err != nil {
		t.Error(err)
	}

	fmt.Println("ASKS:")
	for _, a := range ob[1] {
		fmt.Printf("price: %f, size: %f\n", a[0], a[1])
	}

	fmt.Println("BIDS:")
	for _, b := range ob[0] {
		fmt.Printf("price: %f, size: %f\n", b[0], b[1])
	}
}
