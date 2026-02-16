package main

import (
	"context"
	"encoding/json"
	"fmt"
	"nlkli/raytrade/internal/broker"
	"nlkli/raytrade/internal/broker/bybit"
	"nlkli/raytrade/internal/cdl"
	"testing"
	"time"
)

func TestBybitGetKline(t *testing.T) {
}

func TestWsConnV2(t *testing.T) {
	// tx := make(chan []byte)
	//
	// client := bybit.NewClientFromEnv(context.Background())
	//
	// stream := client.CreatePublicStreamV2(models.CategoryLinear, tx)
	//
	// sub, err := stream.Subscribe("kline.5.BTCUSDT")
	//
	//	if err != nil {
	//		t.Error(err)
	//	}
	//
	//	go func() {
	//		time.Sleep(time.Second * 10)
	//		stream.Unsubscribe("kline.5.BTCUSDT")
	//		println("Send unsub")
	//	}()
	//
	//	for d := range sub {
	//		println(string(d.Data))
	//	}
	//
	// println("Is unsub")
	// time.Sleep(time.Second * 30)
}

func TestWsConnect(t *testing.T) {
}

func pp(v any) {
	s, _ := json.MarshalIndent(v, "", "    ")
	fmt.Println(string(s))
}

func TestGetCandles(t *testing.T) {
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

func TestPosStream(t *testing.T) {
	client := bybit.NewClientFromEnv(context.Background())

	tx := make(chan []byte)

	stream := client.CreatePrivateStreamV2(tx)

	time.Sleep(2 * time.Second)

	ch, err := stream.Subscribe("position.linear")
	if err != nil {
		t.Error(err)
		return
	}

	for d := range ch {
		fmt.Println(string(d.Data))
	}

}
