package main

import (
	"context"
	"encoding/json"
	"fmt"
	"nlkli/raytrade/internal/broker/bybit"
	"nlkli/raytrade/internal/broker/bybit/models"
	"testing"
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
	ch, err := stream.Subscribe(topics, 8)
	fmt.Println("OK")
	if err != nil {
		t.Error(err)
	}
	for data := range ch {
		var klineData models.StreamKlineData
		json.Unmarshal(data.Data, &klineData)
		fmt.Printf("%+v\n", klineData)
	}
}
