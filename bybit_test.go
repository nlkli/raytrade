package main

import (
	"context"
	"encoding/json"
	"fmt"
	"nlkli/raytrade/internal/broker/bybit"
	"nlkli/raytrade/internal/broker/bybit/models"
	"sort"
	"strconv"
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

func printOrderBook(ob *[2][][2]float64) {
	fmt.Println("\n=================== ORDER BOOK ===================")
	fmt.Printf("LEN: bids=%d asks=%d\n", len(ob[0]), len(ob[1]))
	fmt.Println("---------------------------------------------------")
	fmt.Println("  ASKS (продажа)    |    BIDS (покупка)  ")
	fmt.Println("  цена     | объем   |    цена     | объем   ")
	fmt.Println("---------------------------------------------------")

	maxLen := len(ob[0])
	if len(ob[1]) > maxLen {
		maxLen = len(ob[1])
	}

	for i := 0; i < maxLen; i++ {
		// ASKS - идем с конца (от лучшей цены к худшей)
		askIdx := len(ob[1]) - 1 - i
		if askIdx >= 0 {
			fmt.Printf("%8.2f | %8.4f", ob[1][askIdx][0], ob[1][askIdx][1])
		} else {
			fmt.Print("                ")
		}

		fmt.Print("  |  ")

		// BIDS - идем с начала (от лучшей цены к худшей)
		if i < len(ob[0]) {
			fmt.Printf("%8.2f | %8.4f", ob[0][i][0], ob[0][i][1])
		}
		fmt.Println()
	}
	fmt.Println("===================================================")
	fmt.Println()
}

func TestOrderBookStream(t *testing.T) {
	client := bybit.NewClientFromEnv(context.Background())

	tx := make(chan []byte, 2)
	depth := 50

	stream := client.CreatePublicStreamV2(models.CategoryLinear, tx, nil)
	ch, err := stream.Subscribe("orderbook.50.BTCUSDT")

	if err != nil {
		t.Error(err)
	}

	// [bids, asks][][price, size]
	ob := [2][][2]float64{
		make([][2]float64, 0, 100),
		make([][2]float64, 0, 100),
	}

	for d := range ch {
		printOrderBook(&ob)

		var obData models.StreamOrderBookData
		if err := json.Unmarshal(d.Data, &obData); err != nil {
			continue
		}

		switch d.Type {
		case "snapshot":
			ob[0] = ob[0][:0]
			ob[1] = ob[1][:0]

			for _, bid := range obData.Bids {
				price, _ := strconv.ParseFloat(bid[0], 64)
				size, _ := strconv.ParseFloat(bid[1], 64)
				ob[0] = append(ob[0], [2]float64{price, size})
			}

			for _, ask := range obData.Asks {
				price, _ := strconv.ParseFloat(ask[0], 64)
				size, _ := strconv.ParseFloat(ask[1], 64)
				ob[1] = append(ob[1], [2]float64{price, size})
			}

		case "delta":
			if bids := obData.Bids; len(bids) > 0 {
				n := len(ob[0])
				for _, bid := range bids {
					price, _ := strconv.ParseFloat(bid[0], 64)
					size, _ := strconv.ParseFloat(bid[1], 64)

					i := sort.Search(n, func(j int) bool {
						return ob[0][j][0] <= price
					})

					if i < n && ob[0][i][0] == price {
						if size == 0 {
							copy(ob[0][i:], ob[0][i+1:n])
							ob[0] = ob[0][:n-1]
							n--
						} else {
							ob[0][i][1] = size
						}
					} else if size != 0 {
						ob[0] = append(ob[0], [2]float64{})
						copy(ob[0][i+1:], ob[0][i:n])
						ob[0][i] = [2]float64{price, size}
						n++
					}
				}

				if n > depth {
					ob[0] = ob[0][:depth]
				}
			}

			if asks := obData.Asks; len(asks) > 0 {
				n := len(ob[1])
				for _, ask := range asks {
					price, _ := strconv.ParseFloat(ask[0], 64)
					size, _ := strconv.ParseFloat(ask[1], 64)

					i := sort.Search(n, func(j int) bool {
						return ob[1][j][0] >= price
					})

					if i < n && ob[1][i][0] == price {
						if size == 0 {
							copy(ob[1][i:], ob[1][i+1:n])
							ob[1] = ob[1][:n-1]
							n--
						} else {
							ob[1][i][1] = size
						}
					} else if size != 0 {
						ob[1] = append(ob[1], [2]float64{})
						copy(ob[1][i+1:], ob[1][i:n])
						ob[1][i] = [2]float64{price, size}
						n++
					}
				}

				if n > depth {
					ob[1] = ob[1][:depth]
				}
			}
		}
	}
}

func TestCandlesStream(t *testing.T) {
	// client := bybit.NewClientFromEnv(context.Background())
	// b := bybit.NewBroker(client)

	// stream := b.CreateStream(broker.Futures, nil)
	// sub, err := stream.SubscribeCandleStream("BTCUSDT", cdl.M1)

	// if err != nil {
	// 	t.Error(err)
	// }

	// for d := range sub.C {
	// 	fmt.Printf("%+v\n", d)
	// }
}

func TestPosStream(t *testing.T) {
	client := bybit.NewClientFromEnv(context.Background())

	tx := make(chan []byte)

	stream := client.CreatePrivateStreamV2(tx, nil)

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
