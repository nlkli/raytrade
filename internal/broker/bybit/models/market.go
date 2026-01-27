package models

type KlineResult struct {
	Category string      `json:"category"`
	Symbol   string      `json:"symbol"`
	List     [][7]string `json:"list"`
}
