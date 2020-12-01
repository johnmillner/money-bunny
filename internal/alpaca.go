package internal

import (
	"fmt"
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"log"
	"strconv"
	"sync"
	"time"
)

type Alpaca struct {
	Client *alpaca.Client
}

func NewAlpaca() *Alpaca {
	return &Alpaca{
		Client: alpaca.NewClient(&common.APIKey{
			ID:     viper.GetString("alpaca.key"),
			Secret: viper.GetString("alpaca.secret"),
		})}
}

func (a Alpaca) GetStocks(symbols ...string) []*Stock {
	stocks := make([]*Stock, 0)

	limit := viper.GetInt("trend") + viper.GetInt("snapshot-lookback-min") + 2
	chunks := SplitList(symbols, viper.GetInt("chunk-size"))

	m := sync.RWMutex{}
	wg := sync.WaitGroup{}

	for _, chunk := range chunks {
		wg.Add(1)

		go func(chunk []string) {
			defer wg.Done()

			bars, err := a.Client.ListBars(chunk, alpaca.ListBarParams{
				Timeframe: "1Min",
				Limit:     &limit,
			})

			if err != nil {
				log.Panicf("could not gather historical prices due to %s", err)
			}

			for symbol, bar := range bars {
				if len(bar) < limit {
					continue
				}

				if time.Now().Sub(bar[len(bar)-1].GetTime()) > 2*time.Minute {
					continue
				}

				m.Lock()
				stocks = append(stocks, NewStock(symbol, bar))
				m.Unlock()
			}
		}(chunk)
	}

	wg.Wait()

	return stocks
}

func SplitList(symbols []string, chunkSize int) [][]string {
	chunks := make([][]string, 0)

	for i := 0; i < len(symbols); i += chunkSize {
		stop := i + chunkSize
		if len(symbols) < stop {
			stop = len(symbols)
		}
		chunks = append(chunks, symbols[i:stop])
	}

	return chunks
}

func (a Alpaca) GetMarketTime() (bool, time.Time, time.Time) {
	today := time.Now().Format("2006-01-02")
	times, err := a.Client.GetCalendar(&today, &today)

	if err != nil {
		log.Panicf("could not gather todays time due to %s", err)
	}

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Panicf("could not parse timezone due to %s", err)
	}
	marketOpen, err := time.ParseInLocation("2006-01-02T15:04", fmt.Sprintf("%sT%s", times[0].Date, times[0].Open), loc)
	if err != nil {
		log.Panicf("could not parse time due to %s", err)
	}
	marketClose, err := time.ParseInLocation("2006-01-02T15:04", fmt.Sprintf("%sT%s", times[0].Date, times[0].Close), loc)
	if err != nil {
		log.Panicf("could not parse time due to %s", err)
	}

	return today == times[0].Date, marketOpen, marketClose
}

func (a Alpaca) OrderBracket(symbol string, qty, takeProfit, stopLoss, stopLimit float64) {
	tp := decimal.NewFromFloat(takeProfit)
	sl := decimal.NewFromFloat(stopLoss)
	st := decimal.NewFromFloat(stopLimit)

	_, err := a.Client.PlaceOrder(
		alpaca.PlaceOrderRequest{
			AssetKey:    &symbol,
			Qty:         decimal.NewFromFloat(qty),
			Side:        "buy",
			Type:        "market",
			TimeInForce: "day",
			OrderClass:  "bracket",
			TakeProfit: &alpaca.TakeProfit{
				LimitPrice: &tp,
			},
			StopLoss: &alpaca.StopLoss{
				StopPrice:  &sl,
				LimitPrice: &st,
			},
		})

	if err != nil {
		log.Printf("could not complete order for %s from alpaca_wrapper due to %s", symbol, err)
	}
}

func (a Alpaca) ListOpenOrders() []alpaca.Order {
	open := "open"
	roll := false
	orders, err := a.Client.ListOrders(&open, nil, nil, &roll)
	if err != nil {
		log.Panicf("could not list open orders in account due to %s", err)
		// todo recover
	}

	return orders
}

func (a Alpaca) GetSpendableAmount() float64 {
	account := a.GetAccount()

	// equity * multiplier = 100k
	// buying power = 76k
	// 100k-76k = amount spent = 24k
	// if m=1
	// (equity * m - amountSpent) = 25k * 1 - 24k = amount left to spend = 1k
	// if m=2
	// (equity * m - amountSpent) = 25k * 2 - 24k = amount left to spend = 26k
	// if m=3
	// (equity * m - amountSpent) = 25k * 3 - 24k = amount left to spend = 51k
	// if m=4
	// (equity * m - amountSpent) = 25k * 4 - 24k = amount left to spend = 76k

	equity, _ := account.Equity.Float64()
	buyingPower, _ := account.BuyingPower.Float64()
	marginMultiplier := viper.GetFloat64("margin-multiplier")
	multiplier, err := strconv.ParseFloat(account.Multiplier, 8)
	if err != nil {
		log.Panicf("could not parse the multiplier into a float due to %s", err)
	}

	idealMargin := equity * multiplier
	spent := idealMargin - buyingPower
	return equity*marginMultiplier - spent
}

func (a Alpaca) LiquidatePosition(order alpaca.Order) {
	err := a.Client.CancelOrder(order.ID)

	if err != nil {
		log.Panicf("could not cancel old order for %s due to %s", order.Symbol, err)
	}

	err = a.Client.ClosePosition(order.Symbol)

	if err != nil {
		log.Printf("could not liqudate old position for %s due to %s", order.Symbol, err)
	}
}

func (a Alpaca) GetQuote(symbol string) *alpaca.LastQuoteResponse {
	quote, err := a.Client.GetLastQuote(symbol)

	if err != nil {
		log.Panicf("could not get the last quote for %s due to %s", symbol, err)
	}

	return quote
}

func (a *Alpaca) ListPositions() []alpaca.Position {
	quote, err := a.Client.ListPositions()

	if err != nil {
		log.Panicf("could not get the current positions due to %s", err)
	}

	return quote
}

func (a *Alpaca) GetAccount() alpaca.Account {
	account, err := a.Client.GetAccount()

	if err != nil {
		log.Panicf("could not complete portfollio gather from alpaca_wrapper due to %s", err)
	}

	return *account
}
