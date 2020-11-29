package stock

import (
	"fmt"
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/markcheno/go-talib"
	"github.com/spf13/viper"
	"log"
	"os"
	"time"
)

type Stock struct {
	Symbol                                  string
	Price, High, Low                        Ouroboros
	Macd, Signal, Trend, Vel, Acc, Atr, Vol []float64
	Updates                                 chan Stock
}

func NewStockAtr(symbol string, bar []alpaca.Bar) *Stock {
	closingPrices := make([]float64, len(bar))
	lowPrices := make([]float64, len(bar))
	highPrices := make([]float64, len(bar))
	volume := make([]float64, len(bar))

	for i, b := range bar {
		closingPrices[i] = float64(b.Close)
		lowPrices[i] = float64(b.Low)
		highPrices[i] = float64(b.High)
		volume[i] = float64(b.Volume)
	}

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
		Atr:     atr,
		Vol:     volume,
		Updates: make(chan Stock, 100),
	}
}

func NewStock(symbol string, bar []alpaca.Bar) Stock {
	closingPrices := make([]float64, len(bar))
	lowPrices := make([]float64, len(bar))
	highPrices := make([]float64, len(bar))
	volume := make([]float64, len(bar))
	for i, b := range bar {
		closingPrices[i] = float64(b.Close)
		lowPrices[i] = float64(b.Low)
		highPrices[i] = float64(b.High)
		volume[i] = float64(b.Volume)
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

	return Stock{
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
		Vol:     volume,
		Updates: make(chan Stock, 100),
	}
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

func (s *Stock) IsBuyingMacdCrossOver() bool {
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

func (s *Stock) IsSellingMacdCrossUnder() bool {
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

func (s *Stock) IsReadyToBuy() bool {
	return s.IsBelowTrend() && s.IsUpwardsTrend() && s.IsBuyingMacdCrossOver()
}

func (s *Stock) IsReadyToSell() bool {
	return !s.IsBelowTrend() && s.IsDownwardsTrend() && s.IsSellingMacdCrossUnder()
}

func (s *Stock) IsBelowTrend() bool {
	return s.Price.Peek() < s.Trend[len(s.Trend)-1]
}

func (s *Stock) IsUpwardsTrend() bool {
	return s.Vel[len(s.Vel)-1] > 0 || s.Acc[len(s.Acc)-1] > 0
}

func (s *Stock) IsDownwardsTrend() bool {
	return s.Vel[len(s.Vel)-1] < 0 || s.Acc[len(s.Acc)-1] < 0
}

func (s Stock) LogSnapshot(action string, price, qty, takeProfit, stopLoss float64) {
	s.CreateGraph()

	log.Printf("%s %s:\n\t"+
		"total: %f\n\t"+
		"qty: %f\n\t"+
		"maxProfit: %f\n\t"+
		"maxLoss: %v\n\t"+
		"price: %f\n\t"+
		"takeProfit: %f\n\t"+
		"stopLoss: %f",
		action,
		s.Symbol,
		price*qty,
		qty,
		(takeProfit-price)*qty,
		(price-stopLoss)*qty,
		price,
		takeProfit,
		stopLoss)
}

func (s *Stock) CreateGraph() {
	lookback := viper.GetInt("snapshot-lookback-min")
	scaleStart := float32(70)

	// create a new line instance
	prices := charts.NewLine()
	macds := charts.NewLine()
	trends := charts.NewLine()
	atrs := charts.NewLine()

	yAxisOpts := charts.WithYAxisOpts(opts.YAxis{
		Scale: true,
	})
	toolTipOpts := charts.WithTooltipOpts(opts.Tooltip{
		Show:      true,
		TriggerOn: "mousemove|click",
	})
	zoomOpts := charts.WithDataZoomOpts(opts.DataZoom{
		Start:      scaleStart,
		End:        100,
		XAxisIndex: []int{0},
	})
	initOpts := charts.WithInitializationOpts(opts.Initialization{
		Width: "100%",
	})

	prices.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Price and Trend",
		Subtitle: "minutes",
	}), yAxisOpts, toolTipOpts, zoomOpts, initOpts)

	macds.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "MACD and Signal",
		Subtitle: "minutes",
	}), yAxisOpts, toolTipOpts, zoomOpts, initOpts)

	trends.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Trend Velocity and Acceleration",
		Subtitle: "minutes",
	}), yAxisOpts, toolTipOpts, zoomOpts, initOpts)

	atrs.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Average True Range",
		Subtitle: "minutes",
	}), yAxisOpts, toolTipOpts, zoomOpts, initOpts)

	xAxis := make([]int, 0)

	for i := lookback; i >= 0; i-- {
		xAxis = append(xAxis, i)
	}

	prices.SetXAxis(xAxis)
	macds.SetXAxis(xAxis)
	trends.SetXAxis(xAxis)
	atrs.SetXAxis(xAxis)

	// Put data into instance
	prices.AddSeries("Price", convertToItems(s.Price.Raster()[len(s.Price.Raster())-1-lookback:]))
	prices.AddSeries("Trend", convertToItems(s.Trend[len(s.Trend)-1-lookback:]))

	macds.AddSeries("Macd", convertToItems(s.Macd[len(s.Macd)-1-lookback:]))
	macds.AddSeries("Signal", convertToItems(s.Signal[len(s.Signal)-1-lookback:]))

	trends.AddSeries("Trend Velocity", convertToItems(s.Vel[len(s.Vel)-1-lookback:]))
	trends.AddSeries("Trend Acceleration", convertToItems(s.Acc[len(s.Acc)-1-lookback:]))

	atrs.AddSeries("ATR", convertToItems(s.Atr[len(s.Atr)-1-lookback:]))

	page := components.NewPage()
	page.AddCharts(prices, trends, macds, atrs)
	page.AddCustomizedCSSAssets("graph.css")
	page.PageTitle = s.Symbol

	f, _ := os.Create(fmt.Sprintf("snapshots/%s-%s.html", time.Now().Format("2006-01-02T15-04-05"), s.Symbol))
	page.SetLayout(components.PageFlexLayout)
	_ = page.Render(f)
}

func convertToItems(array []float64) []opts.LineData {
	items := make([]opts.LineData, len(array))
	for i := range array {
		items[i] = opts.LineData{Value: array[i]}
	}

	return items
}
