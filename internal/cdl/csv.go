package cdl

import (
	"encoding/csv"
	"errors"
	"os"
	"strconv"
)

var (
	csvColumns = [6]string{
		"time",
		"open",
		"high",
		"low",
		"close",
		"volume",
	}

	errParseCandle = errors.New("candle parsing error: invalid data format")
	errEmptyData   = errors.New("the file contains no data")
	errInvalidData = errors.New("invalid file format or headers")
)

func ParseCandleFromRawData(data [6]string) (candle Candle, err error) {
	candle.Time, err = strconv.ParseInt(data[0], 10, 64)
	if err != nil {
		return candle, errParseCandle
	}
	candle.O, err = strconv.ParseFloat(data[1], 64)
	if err != nil {
		return candle, errParseCandle
	}
	candle.H, err = strconv.ParseFloat(data[2], 64)
	if err != nil {
		return candle, errParseCandle
	}
	candle.L, err = strconv.ParseFloat(data[3], 64)
	if err != nil {
		return candle, errParseCandle
	}
	candle.C, err = strconv.ParseFloat(data[4], 64)
	if err != nil {
		return candle, errParseCandle
	}
	candle.Volume, err = strconv.ParseFloat(data[5], 64)
	if err != nil {
		return candle, errParseCandle
	}

	return candle, nil
}

func CandlesFromCsv(path string) ([]Candle, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) <= 1 {
		return nil, errEmptyData
	}
	if len(records[0]) != len(csvColumns) {
		return nil, errInvalidData
	}

	for i, col := range records[0] {
		if col != csvColumns[i] {
			return nil, errInvalidData
		}
	}

	candles := make([]Candle, 0, len(records)-1)
	for _, record := range records[1:] {
		candle, err := ParseCandleFromRawData([6]string(record))
		if err != nil {
			return candles, err
		}
		candles = append(candles, candle)
	}

	return candles, nil
}
