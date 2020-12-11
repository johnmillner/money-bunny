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
	Symbol                   string
	Snapshots                *ouroboros
	Macd, Signal, Trend, Atr []float64
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

	macd, signal, trend, atr := getIndicators(snapshots.Get())

	return &Stock{
		Symbol:    symbol,
		Snapshots: snapshots,
		Macd:      macd,
		Signal:    signal,
		Trend:     trend,
		Atr:       atr,
	}
}

func (s *Stock) Update(aggregate io.Aggregate) *Stock {
	s.Snapshots.Insert(Snapshot{
		Price:     aggregate.C,
		High:      aggregate.H,
		Low:       aggregate.L,
		Vol:       aggregate.V,
		timestamp: time.Unix(aggregate.E, 0),
	})

	s.Macd, s.Signal, s.Trend, s.Atr = getIndicators(s.Snapshots.Get())

	return s
}

func GetRawData(snapshots []Snapshot) ([]float64, []float64, []float64, []float64, []time.Time) {
	closingPrices := make([]float64, len(snapshots))
	lowPrices := make([]float64, len(snapshots))
	highPrices := make([]float64, len(snapshots))
	volume := make([]float64, len(snapshots))
	times := make([]time.Time, len(snapshots))

	for i, snapshot := range snapshots {
		closingPrices[i] = snapshot.Price
		lowPrices[i] = snapshot.Low
		highPrices[i] = snapshot.High
		volume[i] = snapshot.Vol
		times[i] = snapshot.timestamp
	}

	return closingPrices, lowPrices, highPrices, volume, times
}

func getIndicators(snapshots []Snapshot) ([]float64, []float64, []float64, []float64) {
	closingPrices, lowPrices, highPrices, _, _ := GetRawData(snapshots)

	macd, signal, _ := talib.Macd(
		closingPrices,
		viper.GetInt("macd.fast"),
		viper.GetInt("macd.slow"),
		viper.GetInt("macd.signal"))

	trend := talib.Ema(closingPrices, viper.GetInt("trend"))

	atr := talib.Atr(
		highPrices,
		lowPrices,
		closingPrices,
		viper.GetInt("atr"))

	return macd, signal, trend, atr
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
	atrs.SetXAxis(xAxis)

	price, _, _, _, _ := GetRawData(s.Snapshots.Get())

	// Put data into instance
	prices.AddSeries("Price", convertToItems(price[len(price)-1-lookback:]))
	prices.AddSeries("Trend", convertToItems(s.Trend[len(s.Trend)-1-lookback:]))

	macds.AddSeries("Macd", convertToItems(s.Macd[len(s.Macd)-1-lookback:]))
	macds.AddSeries("Signal", convertToItems(s.Signal[len(s.Signal)-1-lookback:]))

	atrs.AddSeries("ATR", convertToItems(s.Atr[len(s.Atr)-1-lookback:]))

	page := components.NewPage()
	page.AddCharts(prices, macds, atrs)
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
