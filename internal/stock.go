package internal

import (
	"fmt"
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/johnmillner/money-bunny/io"
	"github.com/markcheno/go-talib"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"time"
)

type Stock struct {
	Symbol                             string
	Snapshots                          *ouroboros
	Macd, Signal, Trend, Vel, Acc, Atr []float64
}

type Snapshot struct {
	Price, High, Low, Vol float64
	timestamp             time.Time
}

func NewStock(symbol string, bar []alpaca.Bar) *Stock {
	snapshots := NewOuroboros(make([]Snapshot, len(bar)))

	for _, b := range bar {
		snapshots.Insert(Snapshot{
			Price:     float64(b.Close),
			High:      float64(b.High),
			Low:       float64(b.Low),
			Vol:       float64(b.Volume),
			timestamp: b.GetTime(),
		})
	}

	macd, signal, trend, vel, acc, atr := getIndicators(snapshots.Get())

	return &Stock{
		Symbol:    symbol,
		Snapshots: snapshots,
		Macd:      macd,
		Signal:    signal,
		Trend:     trend,
		Vel:       vel,
		Acc:       acc,
		Atr:       atr,
	}
}

func (s *Stock) Update(aggregate io.Aggregate) *Stock {
	s.Snapshots.Insert(Snapshot{
		Price:     aggregate.C,
		High:      aggregate.H,
		Low:       aggregate.L,
		Vol:       aggregate.V,
		timestamp: aggregate.E,
	})

	s.Macd, s.Signal, s.Trend, s.Vel, s.Acc, s.Atr = getIndicators(s.Snapshots.Get())

	s.LogSnapshot("test", 0, 0, 0, 0)
	return s
}

func getPriceAndTime(snapshots []Snapshot) ([]float64, []time.Time) {
	closingPrices := make([]float64, len(snapshots))
	times := make([]time.Time, len(snapshots))

	for i, snapshot := range snapshots {
		closingPrices[i] = snapshot.Price
		times[i] = snapshot.timestamp
	}

	return closingPrices, times
}

func getIndicators(snapshots []Snapshot) ([]float64, []float64, []float64, []float64, []float64, []float64) {
	closingPrices := make([]float64, len(snapshots))
	lowPrices := make([]float64, len(snapshots))
	highPrices := make([]float64, len(snapshots))
	volume := make([]float64, len(snapshots))

	for i, snapshot := range snapshots {
		closingPrices[i] = snapshot.Price
		lowPrices[i] = snapshot.Low
		highPrices[i] = snapshot.High
		volume[i] = snapshot.Vol
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

	return macd, signal, trend, vel, acc, atr
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

func (s Stock) LogSnapshot(action string, price, qty, takeProfit, stopLoss float64) {
	s.CreateGraph()

	logrus.
		WithField("stock", s.Symbol).
		WithField("action", action).
		WithField("total", price*qty).
		WithField("qty", qty).
		WithField("maxProfit", (takeProfit-price)*qty).
		WithField("maxLoss", (price-stopLoss)*qty).
		WithField("price", price).
		WithField("takeProfit", takeProfit).
		WithField("stopLoss", stopLoss).
		Info()
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

	price, _ := getPriceAndTime(s.Snapshots.Get())

	// Put data into instance
	prices.AddSeries("Price", convertToItems(price[len(price)-1-lookback:]))
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
