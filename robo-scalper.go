package main

import (
	"github.com/johnmillner/robo-macd/config"
	"github.com/johnmillner/robo-macd/io"
	"github.com/johnmillner/robo-macd/stock"
	"github.com/spf13/viper"
	"log"
	"math"
	"runtime/debug"
	"sync"
	"time"
)

func main() {
	log.Print("starting robo-scalper")

	// read in configs
	config.Config()

	wg := sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()

	go recovery(time.Now(), func() {
		symbols := viper.GetStringSlice("stocks")

		a := io.NewAlpaca()
		stocks := a.GetHistoricalStocks(symbols)
		io.LiveUpdates(stocks)

		for _, symbol := range symbols {
			go monitorStock(stocks[symbol], a)
		}
	})
}

func recovery(start time.Time, f func()) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("recovering from panic %v", err)
			debug.PrintStack()
			if start.Add(time.Duration(viper.GetInt("recover-frequency-min")) * time.Minute).Before(time.Now()) {
				log.Panicf("too many panics - will not recover due to %v", err)
			}

			go recovery(time.Now(), f)
		}
	}()

	f()
}

func monitorStock(stock *stock.Stock, a *io.Alpaca) {
	_, marketOpen, marketClose := a.GetMarketTime()

	marketGetIn := marketOpen.Add(time.Duration(viper.GetInt("trade-after-open-min")) * time.Minute)
	marketGetOut := marketClose.Add(-1 * time.Duration(viper.GetInt("close-before-close-min")) * time.Minute)

	log.Printf("market times today are: open %v close %v", marketGetIn, marketGetOut)

	for update := range stock.Updates {
		orders := a.GetOrders(stock.Symbol)
		for _, order := range orders {
			// sell all orders if close to marketClose
			if marketGetOut.Before(time.Now()) {
				log.Printf("liqudating %s since it's close to market close %v current time %v",
					order.Symbol,
					marketGetOut,
					time.Now().UTC())
				a.LiquidatePosition(order)
				continue
			}

			// remove old order/positions
			if order.SubmittedAt.Add(time.Duration(viper.GetInt("liquidate-after-min")) * time.Minute).
				Before(time.Now()) {
				log.Printf("liqudating %s since it was too old submitted at %v current time %v",
					order.Symbol,
					order.SubmittedAt,
					time.Now().UTC())
				a.LiquidatePosition(order)
				continue
			}

			// check if stock should be sold because of macd signal
			if update.IsReadyToSell() {
				qty, _ := order.Qty.Float64()
				update.LogSnapshot("selling", update.Price.Peek(), qty, 0, 0)
				a.LiquidatePosition(order)
				continue
			}
		}

		if marketGetIn.After(time.Now()) || marketGetOut.Before(time.Now()) {
			continue
		}

		// check if this stock is good to buy
		if update.IsReadyToBuy() {
			withinRisk, price, qty, takeProfit, stopLoss, stopLimit := getOrderParameters(update, a)

			// check if trade is too risky or too boring
			if !withinRisk {
				update.LogSnapshot("skipping", price, qty, takeProfit, stopLoss)
				continue
			}

			a.OrderBracket(update.Symbol, qty, takeProfit, stopLoss, stopLimit)
			update.LogSnapshot("buying", price, qty, takeProfit, stopLoss)
		}
	}
}

func getOrderParameters(s stock.Stock, a *io.Alpaca) (bool, float64, float64, float64, float64, float64) {
	quote := io.GetQuote(s.Symbol)
	portfolio := a.GetPortfolioValue()
	portfolioRisk := viper.GetFloat64("risk")
	numberOfStocks := float64(len(viper.GetStringSlice("stocks")))
	marginPercent := viper.GetFloat64("margin")

	exposure := portfolio / numberOfStocks * portfolioRisk * (1 + marginPercent)

	atr := s.Atr[len(s.Atr)-1]
	price := quote.Last.Askprice - (quote.Last.Askprice-quote.Last.Bidprice)/2

	tradeRisk := 2 * atr
	rewardToRisk := viper.GetFloat64("riskReward")
	stopLossMax := viper.GetFloat64("stopLossMax")

	takeProfit := price + rewardToRisk*tradeRisk
	stopLoss := price - tradeRisk
	stopLimit := price - (1+stopLossMax)*tradeRisk

	qty := math.Round(math.Min(exposure/tradeRisk, portfolio/quote.Last.Askprice))

	//ensure we dont go over
	for qty*price > portfolio/numberOfStocks {
		qty = qty - 1
	}

	return true, price, qty, takeProfit, stopLoss, stopLimit
}
