package main

import (
	"context"
	"fmt"
	"nlkli/raytrade/internal/broker/bybit"
	"nlkli/raytrade/internal/broker/bybit/models"
	"testing"

	"github.com/gorilla/websocket"
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
	conn, _, err := websocket.DefaultDialer.Dial("wss://stream.bybit.com/v5/public/linear", nil)
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	msg := `{
		"req_id": "test",
		"op": "subscribe",
		"args": [
			"kline.1.BTCUSDT"
		]
	}`

	err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		t.Error(err)
	}

	n := 0
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			t.Error(err)
		}
		fmt.Println(string(data))
		if n > 6 {
			break
		}
		n++
	}
}
