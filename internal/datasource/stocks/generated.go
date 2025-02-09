package stocks

import (
	"encoding/json"
	"time"
)

type StockData struct {
	Timestamps []time.Time
	Symbols    []Symbol
}

type Symbol struct {
	Name   string
	Link   string
	Values []float64
}

type CurrentTradingPeriod struct {
	Pre     TradingPeriod `json:"pre"`
	Regular TradingPeriod `json:"regular"`
	Post    TradingPeriod `json:"post"`
}

type TradingPeriod struct {
	Timezone  string `json:"timezone"`
	End       int    `json:"end"`
	Start     int    `json:"start"`
	Gmtoffset int    `json:"gmtoffset"`
}

type Indicators struct {
	Quote    []QuoteData `json:"quote"`
	AdjClose []AdjClose  `json:"adjclose"`
}

type QuoteData struct {
	Volume []int     `json:"volume"`
	Open   []float64 `json:"open"`
	High   []float64 `json:"high"`
	Close  []float64 `json:"close"`
	Low    []float64 `json:"low"`
}

type AdjClose struct {
	AdjClose []float64 `json:"adjclose"`
}

type Meta struct {
	Currency             string               `json:"currency"`
	Symbol               string               `json:"symbol"`
	ExchangeName         string               `json:"exchangeName"`
	InstrumentType       string               `json:"instrumentType"`
	FirstTradeDate       int                  `json:"firstTradeDate"`
	RegularMarketTime    int                  `json:"regularMarketTime"`
	Gmtoffset            int                  `json:"gmtoffset"`
	Timezone             string               `json:"timezone"`
	ExchangeTimezoneName string               `json:"exchangeTimezoneName"`
	RegularMarketPrice   float64              `json:"regularMarketPrice"`
	ChartPreviousClose   float64              `json:"chartPreviousClose"`
	PriceHint            int                  `json:"priceHint"`
	CurrentTradingPeriod CurrentTradingPeriod `json:"currentTradingPeriod"`
	DataGranularity      string               `json:"dataGranularity"`
	Range                string               `json:"range"`
	ValidRanges          []string             `json:"validRanges"`
}

type Result struct {
	Meta          Meta  `json:"meta"`
	TimestampUnix []int `json:"timestamp"`
	Timestamp     []time.Time
	Indicators    Indicators `json:"indicators"`
}

func (w *Result) UnmarshalJSON(data []byte) error {
	type Alias Result

	tmp := Alias{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	tmp.Timestamp = make([]time.Time, len(tmp.TimestampUnix))
	for index, unix := range tmp.TimestampUnix {
		tmp.Timestamp[index] = time.Unix(int64(unix), 0)
	}

	*w = Result(tmp)
	return nil
}

type Chart struct {
	Result []Result    `json:"result"`
	Error  interface{} `json:"error"`
}

type Response struct {
	Chart Chart `json:"chart"`
}
