package models

type Kline struct {
	Category string      `json:"category"`
	Symbol   string      `json:"symbol"`
	List     [][7]string `json:"list"`
}
