package io

import (
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/johnmillner/money-bunny/stock"
	"github.com/spf13/viper"
	"log"
	"sync"
)

func FilterByTradability(a *Alpaca) []string {
	active := "active"
	assets, err := a.Client.ListAssets(&active)

	if err != nil {
		log.Panicf("could not list assets due to %s", err)
	}

	symbols := make([]string, 0)
	for _, asset := range assets {
		if asset.Tradable && asset.Marginable && asset.EasyToBorrow {
			symbols = append(symbols, asset.Symbol)
		}
	}

	return symbols
}

func splitList(symbols []string, chunkSize int) [][]string {
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

func FilterByRisk(a *Alpaca, symbols []string) []string {
	chunks := splitList(symbols, viper.GetInt("chunk-size"))
	period := viper.GetInt("atr") + 1

	stocks := make([]string, 0)
	m := sync.RWMutex{}
	wg := sync.WaitGroup{}

	for _, chunk := range chunks {
		wg.Add(1)

		go func(chunk []string) {
			defer wg.Done()

			chunkBars, err := a.Client.ListBars(chunk, alpaca.ListBarParams{
				Timeframe: "1Min",
				Limit:     &period,
			})

			if err != nil {
				log.Panicf("could not gather atrs due to %s", err)
			}

			for symbol, bars := range chunkBars {
				if len(bars) < period {
					continue
				}

				potential := stock.NewStockAtr(symbol, bars)
				meetsRisk := meetsRiskGoal(potential)
				if !meetsRisk {
					continue
				}

				m.Lock()
				stocks = append(stocks, potential.Symbol)
				m.Unlock()
			}

		}(chunk)
	}

	wg.Wait()
	return stocks

}

func meetsRiskGoal(stock *stock.Stock) bool {
	tradeRisk := viper.GetFloat64("stop-loss-atr-ratio") * stock.Atr[len(stock.Atr)-1] / stock.Price.Peek()
	upperRisk := viper.GetFloat64("risk") * (1 + viper.GetFloat64("exposure-tolerance"))
	lowerRisk := viper.GetFloat64("risk") * (1 - viper.GetFloat64("exposure-tolerance"))

	return tradeRisk > lowerRisk && tradeRisk < upperRisk
}

func FilterByBuyable(stocks ...stock.Stock) []stock.Stock {
	buyableStocks := make([]stock.Stock, 0)
	for _, potential := range stocks {
		if potential.IsReadyToBuy() {
			buyableStocks = append(buyableStocks, potential)
		}

	}
	return buyableStocks
}

func FilterByMarketCap(stock stock.Stock) bool {

	marketCap, err := GetMarketCap(stock.Symbol)
	if err != nil {
		log.Printf("%v", err)
	}

	return marketCap >= viper.GetFloat64("min-market-cap")
}

func FilterByVolume(stock stock.Stock, qty float64) bool {
	totalVol := float64(0)
	for _, vol := range stock.Vol {
		totalVol += vol
	}
	avgVol := totalVol / float64(len(stock.Vol))

	return avgVol*viper.GetFloat64("min-average-vol-multiple") > qty
}
