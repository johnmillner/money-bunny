package main

import (
	"github.com/johnmillner/robo-macd/config"
	"github.com/johnmillner/robo-macd/io"
	"github.com/johnmillner/robo-macd/stock"
	"github.com/spf13/viper"
	"log"
	"math"
	"sync"
	"time"
)

func main() {
	log.Print("starting robo-scalper")

	// read in configs
	config.Config()

	symbols := viper.GetStringSlice("stocks")

	a := io.NewAlpaca()
	stocks := a.GetHistoricalStocks(symbols)
	io.LiveUpdates(stocks)

	wg := sync.WaitGroup{}
	defer wg.Wait()

	for _, symbol := range symbols {
		wg.Add(1)
		go monitorStock(stocks[symbol], a)
	}
}

func monitorStock(stock *stock.Stock, a io.Alpaca) {
	marketOpen, marketClose := a.GetMarketTime()
	marketGetIn := marketOpen.Add(time.Duration(viper.GetInt("trade-after-open-min")) * time.Minute)
	marketGetOut := marketClose.Add(time.Duration(viper.GetInt("close-before-close-min")) * time.Minute)

	for update := range stock.Updates {
		orders := a.GetOrders(stock.Symbol)
		for _, order := range orders {
			// sell all orders if close to marketClose
			if marketGetOut.After(marketClose) {
				log.Printf("liqudating %s since it's close to market close %v current time %v",
					order.Symbol,
					marketClose,
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

		// check if this stock is good to buy
		if marketGetIn.Before(time.Now()) && marketGetOut.After(time.Now()) && update.IsReadyToBuy() {
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

func getOrderParameters(s stock.Stock, a io.Alpaca) (bool, float64, float64, float64, float64, float64) {
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
	for qty*price > portfolio {
		qty = qty - 1
	}

	// ensure that risk is not too boring or risky to be worth trading
	withinRisk :=
		stopLoss < exposure*(1-viper.GetFloat64("exposure-tolerance")) ||
			stopLoss > exposure*(1+viper.GetFloat64("exposure-tolerance"))

	return withinRisk, price, qty, takeProfit, stopLoss, stopLimit
}
