package internal

import (
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

func FilterByCap(symbols ...string) []string {
	caps := make([]string, 0)
	minCap := viper.GetFloat64("min-market-cap")
	lock := sync.RWMutex{}
	wg := sync.WaitGroup{}

	for _, symbol := range symbols {
		wg.Add(1)
		go func(symbol string) {
			defer wg.Done()
			marketCap := GetMarketCap(symbol)

			if marketCap < minCap {
				return
			}

			lock.Lock()
			caps = append(caps, symbol)
			lock.Unlock()
		}(symbol)
	}

	wg.Wait()
	lock.RLock()
	defer lock.RUnlock()

	return caps

}

func FilterByRiskGoal(stock *Stock) bool {
	tradeRisk := viper.GetFloat64("stop-loss-atr-ratio") * stock.Atr[len(stock.Atr)-1] / stock.Price.Peek()
	upperRisk := viper.GetFloat64("risk") * (1 + viper.GetFloat64("exposure-tolerance"))
	lowerRisk := viper.GetFloat64("risk") * (1 - viper.GetFloat64("exposure-tolerance"))

	return tradeRisk > lowerRisk && tradeRisk < upperRisk
}

func FilterByVolume(stock *Stock, qty float64) bool {
	totalVol := float64(0)
	for _, vol := range stock.Vol.Raster() {
		totalVol += vol
	}
	avgVol := totalVol / float64(stock.Vol.Capacity)

	return avgVol*viper.GetFloat64("min-average-vol-multiple") > qty
}
