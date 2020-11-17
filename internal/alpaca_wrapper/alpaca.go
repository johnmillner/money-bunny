package alpaca_wrapper

import (
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/shopspring/decimal"
	"log"
	"time"
)

type AlpacaInterface interface {
	GetBars(period time.Duration, symbols []string, limit int) (map[string][]alpaca.Bar, error)
	GetCalendar(start, end string) ([]alpaca.CalendarDay, error)
	BuyBracket(symbol string, qty int, takeProfit, stopLoss, stopLimit float64) (*alpaca.Order, error)
	GetAccount() (*alpaca.Account, error)
	ListAsserts() ([]alpaca.Asset, error)
	GetQuote(symbol string) (*alpaca.LastQuoteResponse, error)
}

type Alpaca struct {
	client *alpaca.Client
}

func (a Alpaca) getClient() *alpaca.Client {
	if a.client == nil {
		a.client = alpaca.NewClient(common.Credentials())
	}

	return a.client
}

func (a Alpaca) GetBars(period time.Duration, symbols []string, limit int) (map[string][]alpaca.Bar, error) {
	return a.getClient().ListBars(symbols, alpaca.ListBarParams{
		Timeframe: durationToTimeframe(period),
		Limit:     &limit,
	})
}

func (a Alpaca) GetCalendar(start, end string) ([]alpaca.CalendarDay, error) {
	return alpaca.GetCalendar(&start, &end)
}

func (a Alpaca) BuyBracket(symbol string, qty int, takeProfit, stopLoss, stopLimit float64) (*alpaca.Order, error) {
	profit := decimal.NewFromFloat(takeProfit)
	loss := decimal.NewFromFloat(stopLoss)
	limit := decimal.NewFromFloat(stopLimit)

	return alpaca.PlaceOrder(alpaca.PlaceOrderRequest{
		AssetKey:    &symbol,
		Qty:         decimal.New(int64(qty), 1),
		Side:        "buy",
		Type:        "market",
		TimeInForce: "day",
		OrderClass:  "bracket",
		TakeProfit: &alpaca.TakeProfit{
			LimitPrice: &profit,
		},
		StopLoss: &alpaca.StopLoss{
			LimitPrice: &limit,
			StopPrice:  &loss,
		},
	})
}

func (a Alpaca) GetAccount() (*alpaca.Account, error) {
	return alpaca.GetAccount()
}

func (a Alpaca) ListAsserts() ([]alpaca.Asset, error) {
	status := "active"
	return alpaca.ListAssets(&status)
}

func (a Alpaca) GetQuote(symbol string) (*alpaca.LastQuoteResponse, error) {
	return alpaca.GetLastQuote(symbol)
}

func durationToTimeframe(dur time.Duration) string {
	switch dur {
	case time.Minute:
		return string(alpaca.Min1)
	case time.Minute * 5:
		return string(alpaca.Min5)
	case time.Minute * 15:
		return string(alpaca.Min15)
	case time.Hour:
		return string(alpaca.Hour1)
	case time.Hour * 24:
		return string(alpaca.Day1)
	default:
		log.Panicf("cannot translate duration given to alpaca_wrapper timeframe, given: %f (in seconds) "+
			"- only acceptable durations are 1min, 5min, 15min, 1day", dur.Seconds())
		return ""
	}
}
