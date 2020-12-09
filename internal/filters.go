package internal

import (
	"github.com/johnmillner/money-bunny/io"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"sync"
	"time"
)

func FilterByTradable(a *io.Alpaca) []string {
	active := "active"
	assets, err := a.Client.ListAssets(&active)

	if err != nil {
		logrus.
			WithError(err).
			Panic("could not list assets from Alpaca")
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

	start := time.Now()

	for _, symbol := range symbols {
		wg.Add(1)
		go func(symbol string) {
			defer wg.Done()
			marketCap := io.GetMarketCap(symbol)

			if marketCap < minCap {
				return
			}

			lock.Lock()
			caps = append(caps, symbol)
			lock.Unlock()
		}(symbol)
	}

	wg.Wait()

	logrus.Debugf("it took %s to filter by market caps for %d symbols", time.Now().Sub(start).String(), len(symbols))

	lock.RLock()
	defer lock.RUnlock()
	return caps

}

func FilterByRiskGoal(budget, price, stopLoss, qty float64) (bool, float64, float64) {
	minRisk := budget * viper.GetFloat64("risk") * (1 - viper.GetFloat64("exposure-tolerance"))
	risk := (price - stopLoss) * qty
	return risk > minRisk, minRisk, risk
}

func FilterByVolume(stock *Stock, qty float64) bool {
	totalVol := float64(0)
	for _, snapshot := range stock.Snapshots.Get() {
		totalVol += snapshot.Vol
	}

	avgVol := totalVol / float64(stock.Snapshots.capacity)

	return avgVol*viper.GetFloat64("min-average-vol-multiple") > qty
}

func FilterByMacdEntry(s *Stock) bool {
	return IsBelowTrend(s) && IsUpwardsTrend(s) && IsBuyingMacdCrossOver(s)
}

func FilterByMacdExit(s *Stock) bool {
	return !IsBelowTrend(s) && IsDownwardsTrend(s) && IsSellingMacdCrossUnder(s)
}

func IsBuyingMacdCrossOver(s *Stock) bool {
	macdStart := s.Macd[len(s.Macd)-2]
	macdEnd := s.Macd[len(s.Macd)-1]
	signalStart := s.Signal[len(s.Signal)-2]
	signalEnd := s.Signal[len(s.Signal)-1]

	ok, intersection := findIntersection(
		point{
			x: 1,
			y: macdEnd,
		},
		point{
			x: 0,
			y: macdStart,
		},
		point{
			x: 1,
			y: signalEnd,
		},
		point{
			x: 0,
			y: signalStart,
		})

	return ok &&
		intersection.x >= 0 && // ensure cross over happened in the last sample
		intersection.x <= 1 && // ^
		macdEnd > macdStart && // ensure it is a positive cross over event
		intersection.y < 0 // ensure that the crossover happened in negative space
}

func IsSellingMacdCrossUnder(s *Stock) bool {
	macdStart := s.Macd[len(s.Macd)-2]
	macdEnd := s.Macd[len(s.Macd)-1]
	signalStart := s.Signal[len(s.Signal)-2]
	signalEnd := s.Signal[len(s.Signal)-1]

	ok, intersection := findIntersection(
		point{
			x: 1,
			y: macdEnd,
		},
		point{
			x: 0,
			y: macdStart,
		},
		point{
			x: 1,
			y: signalEnd,
		},
		point{
			x: 0,
			y: signalStart,
		})

	return ok &&
		intersection.x >= 0 && // ensure cross over happened in the last sample
		intersection.x <= 1 && // ^
		macdEnd < macdStart && // ensure it is a negative cross over event
		intersection.y > 0 // ensure that the crossover happened in positive space
}

type point struct {
	x, y float64
}

func findIntersection(a, b, c, d point) (bool, point) {
	a1 := b.y - a.y
	b1 := a.x - b.x
	c1 := a1*(a.x) + b1*(a.y)

	a2 := d.y - c.y
	b2 := c.x - d.x
	c2 := a2*(c.x) + b2*(c.y)

	determinant := a1*b2 - a2*b1

	if determinant == 0 {
		return false, point{}
	}

	return true, point{
		x: (b2*c1 - b1*c2) / determinant,
		y: (a1*c2 - a2*c1) / determinant,
	}
}

func IsBelowTrend(s *Stock) bool {
	return s.Snapshots.Get()[s.Snapshots.capacity-1].Price < s.Trend[len(s.Trend)-1]
}

func IsUpwardsTrend(s *Stock) bool {
	return s.Vel[len(s.Vel)-1] > 0 || s.Acc[len(s.Acc)-1] > 0
}

func IsDownwardsTrend(s *Stock) bool {
	return s.Vel[len(s.Vel)-1] < 0 || s.Acc[len(s.Acc)-1] < 0
}
