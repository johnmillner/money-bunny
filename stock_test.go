package main

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/johnmillner/robo-macd/config"
	"github.com/johnmillner/robo-macd/io"
	"os"
	"testing"
)

func TestNewStock(t *testing.T) {
	config.Config()

	a := io.NewAlpaca()

	stocks := a.GetHistoricalStocks([]string{"TSLA"})

	s := stocks["TSLA"]

	// create a new line instance
	prices := charts.NewLine()
	macds := charts.NewLine()
	trends := charts.NewLine()
	atrs := charts.NewLine()

	prices.SetGlobalOptions(charts.WithYAxisOpts(opts.YAxis{
		Scale: true,
	}), charts.WithTooltipOpts(opts.Tooltip{
		Show:      true,
		TriggerOn: "mousemove|click",
	}), charts.WithDataZoomOpts(opts.DataZoom{
		Start:      50,
		End:        100,
		XAxisIndex: []int{0},
	}), charts.WithTooltipOpts(opts.Tooltip{
		Trigger: "axis",
	}))

	macds.SetGlobalOptions(charts.WithYAxisOpts(opts.YAxis{
		Scale: true,
	}), charts.WithTooltipOpts(opts.Tooltip{
		Show:      true,
		TriggerOn: "mousemove|click",
	}), charts.WithDataZoomOpts(opts.DataZoom{
		Start:      50,
		End:        100,
		XAxisIndex: []int{0},
	}), charts.WithTooltipOpts(opts.Tooltip{
		Trigger: "axis",
	}))

	trends.SetGlobalOptions(charts.WithYAxisOpts(opts.YAxis{
		Scale: true,
	}), charts.WithTooltipOpts(opts.Tooltip{
		Show:      true,
		TriggerOn: "mousemove|click",
	}), charts.WithDataZoomOpts(opts.DataZoom{
		Start:      50,
		End:        100,
		XAxisIndex: []int{0},
	}), charts.WithTooltipOpts(opts.Tooltip{
		Trigger: "axis",
	}))

	atrs.SetGlobalOptions(charts.WithYAxisOpts(opts.YAxis{
		Scale: true,
	}), charts.WithTooltipOpts(opts.Tooltip{
		Show:      true,
		TriggerOn: "mousemove|click",
	}), charts.WithDataZoomOpts(opts.DataZoom{
		Start:      50,
		End:        100,
		XAxisIndex: []int{0},
	}), charts.WithTooltipOpts(opts.Tooltip{
		Trigger: "axis",
	}))

	xAxis := make([]int, 0)

	for i := range s.Trend[200:] {
		xAxis = append(xAxis, i)
	}
	prices.SetXAxis(xAxis)
	macds.SetXAxis(xAxis)
	trends.SetXAxis(xAxis)
	atrs.SetXAxis(xAxis)

	// Put data into instance
	prices.AddSeries("Price", convertToItems(s.Trend[200:]))
	prices.AddSeries("Trend", convertToItems(s.Price.Raster()[200:]))

	macds.AddSeries("Macd", convertToItems(s.Macd[200:]))
	macds.AddSeries("Signal", convertToItems(s.Signal[200:]))

	trends.AddSeries("Trend Velocity", convertToItems(s.Vel[200:]))
	trends.AddSeries("Trend Acceleration", convertToItems(s.Acc[200:]))

	atrs.AddSeries("ATR", convertToItems(s.Atr[200:]))

	page := components.NewPage()
	page.AddCharts(prices, macds, trends, atrs)

	// Where the magic happens
	f, _ := os.Create("line.html")
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
