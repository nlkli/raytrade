package main

import (
	"context"
	"fmt"
	"nlkli/raytrade/internal/broker/bybit"
	"nlkli/raytrade/internal/broker/bybit/models"
	"nlkli/raytrade/internal/ws"
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
	conn := ws.NewConn(context.Background(), "wss://stream.bybit.com/v5/public/linear")
	defer conn.Close()

	msg := `{
		"req_id": "test",
		"op": "subscribe",
		"args": [
			"kline.1.BTCUSDT"
		]
	}`
	conn.Send([]byte(msg))
	for {
		b, err := conn.Recv()
		if err != nil {
			t.Error(err)
		}
		fmt.Println(string(b))
	}
}
