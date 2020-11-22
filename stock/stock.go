package stock

import (
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/markcheno/go-talib"
	"github.com/spf13/viper"
	"log"
)

type Stock struct {
	Symbol                             string
	Price, High, Low                   Ouroboros
	Macd, Signal, Trend, Vel, Acc, Atr []float64
	updates                            chan Stock
}

func NewStock(symbol string, bar []alpaca.Bar, updates chan Stock) *Stock {
	closingPrices := make([]float64, len(bar))
	lowPrices := make([]float64, len(bar))
	highPrices := make([]float64, len(bar))
	for i, b := range bar {
		closingPrices[i] = float64(b.Close)
		lowPrices[i] = float64(b.Low)
		highPrices[i] = float64(b.High)
	}

	macd, signal, _ := talib.Macd(
		closingPrices,
		viper.GetInt("macd.fast"),
		viper.GetInt("macd.slow"),
		viper.GetInt("macd.signal"))

	trend, vel, acc := getTrends(closingPrices)

	atr := talib.Atr(
		highPrices,
		lowPrices,
		closingPrices,
		viper.GetInt("atr"))

	return &Stock{
		Symbol:  symbol,
		Price:   NewOuroboros(closingPrices),
		Low:     NewOuroboros(lowPrices),
		High:    NewOuroboros(highPrices),
		Macd:    macd,
		Signal:  signal,
		Trend:   trend,
		Vel:     vel,
		Acc:     acc,
		Atr:     atr,
		updates: updates,
	}
}

// significant room for optimization here
func (s *Stock) Update(close, low, high float64) {
	s.Price = s.Price.Push(close)
	s.Low = s.Low.Push(low)
	s.High = s.High.Push(high)

	prices := s.Price.Raster()
	s.Macd, s.Signal, _ = talib.Macd(
		prices,
		viper.GetInt("macd.fast"),
		viper.GetInt("macd.slow"),
		viper.GetInt("macd.signal"))

	s.Trend, s.Vel, s.Acc = getTrends(prices)

	s.Atr = talib.Atr(
		s.High.Raster(),
		s.Low.Raster(),
		prices,
		viper.GetInt("atr"))

	s.updates <- *s
}

func (s *Stock) Snapshot() ([]float64, []float64, []float64, []float64, []float64, []float64, []float64) {
	priceRaster := s.Price.Raster()
	return priceRaster[len(priceRaster)-viper.GetInt("snapshot-lookback-min")-1:],
		s.Macd[len(s.Macd)-viper.GetInt("snapshot-lookback-min")-1:],
		s.Signal[len(s.Signal)-viper.GetInt("snapshot-lookback-min")-1:],
		s.Trend[len(s.Trend)-viper.GetInt("snapshot-lookback-min")-1:],
		s.Vel[len(s.Vel)-viper.GetInt("snapshot-lookback-min")-1:],
		s.Acc[len(s.Acc)-viper.GetInt("snapshot-lookback-min")-1:],
		s.Atr[len(s.Atr)-viper.GetInt("snapshot-lookback-min")-1:]
}

func getTrends(price []float64) ([]float64, []float64, []float64) {
	trend := talib.Ema(price, viper.GetInt("trend"))

	trendVelocity := make([]float64, len(trend))
	for i := range trend {
		if i == 0 || trend[i-1] == 0 {
			continue
		}
		trendVelocity[i] = trend[i] - trend[i-1]
	}
	trendVelocity = trendVelocity[1:]

	trendAcceleration := make([]float64, len(trendVelocity))
	for i := range trendVelocity {
		if i == 0 || trendVelocity[i-1] == 0 {
			continue
		}
		trendAcceleration[i] = trendVelocity[i] - trendVelocity[i-1]
	}
	trendAcceleration = trendAcceleration[1:]

	return trend, trendVelocity, trendAcceleration
}

func (s *Stock) IsPositiveMacdCrossOver() bool {
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

func (s *Stock) IsNegativeMacdCrossUnder() bool {
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

func (s *Stock) IsBelowTrend() bool {
	return s.Price.Peek() < s.Trend[len(s.Trend)-1]
}

func (s *Stock) IsUpwardsTrend() bool {
	return s.Vel[len(s.Vel)-1] > 0 || s.Acc[len(s.Acc)-1] > 0
}

func (s Stock) LogSnapshot(action string, price, qty, takeProfit, stopLoss float64) {
	p, m, i, t, v, a, r := s.Snapshot()
	log.Printf("%s %s:\n\t"+
		"price %v\n\t"+
		"macd %v\n\t"+
		"signal %v\n\t"+
		"trend %v\n\t"+
		"vel %v\n\t"+
		"acc %v\n\t"+
		"atr %v\n\t"+
		"maxProfit: %f\n\t"+
		"maxLoss: %v\n\t"+
		"price: %f\n\t"+
		"takeProfit: %f\n\t"+
		"stopLoss: %f\n\t"+
		"qty: %f",
		action,
		s.Symbol,
		p[len(p)-viper.GetInt("snapshot-lookback-min")-1:],
		m[len(m)-viper.GetInt("snapshot-lookback-min")-1:],
		i[len(i)-viper.GetInt("snapshot-lookback-min")-1:],
		t[len(t)-viper.GetInt("snapshot-lookback-min")-1:],
		v[len(v)-viper.GetInt("snapshot-lookback-min")-1:],
		a[len(a)-viper.GetInt("snapshot-lookback-min")-1:],
		r[len(r)-viper.GetInt("snapshot-lookback-min")-1:],
		(takeProfit-price)*qty,
		(price-stopLoss)*qty,
		price,
		takeProfit,
		stopLoss,
		qty)
}
