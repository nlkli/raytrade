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

func TestWsConnect(t *testing.T) {
	client := bybit.NewClientFromEnv(context.Background())
	stream := client.CreatePublicStream(models.CategoryLinear)

	topics := []string{"kline.5.BTCUSDT", "kline.5.ADAUSDT", "kline.1.FARTCOINUSDT"}

	s, err := stream.Subscribe(topics, 8)
	if err != nil {
		t.Error(err)
	}

	go func() {
		<-time.After(5 * time.Second)
		stream.Close()
	}()

	for data := range s.C() {
		var klineData models.StreamKlineData
		json.Unmarshal(data.Data, &klineData)
		fmt.Printf("%+v\n", klineData)
	}
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
	done := make(chan struct{}, 1)
	ch, err := b.CandlesStream(done, broker.Futures, "BTCUSDT", cdl.M1)
	if err != nil {
		t.Error(err)
	}
	for c := range ch {
		fmt.Printf("%+v\n", c)
	}
}
