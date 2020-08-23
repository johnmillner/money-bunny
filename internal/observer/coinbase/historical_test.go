package coinbase

import (
	"github.com/johnmillner/robo-macd/internal/observer"
	"github.com/johnmillner/robo-macd/internal/yaml"
	"strings"
	"testing"
	"time"
)

func TestPriceMonitor_PopulateHistorical(t *testing.T) {
	coinbase := Coinbase{}
	err := yaml.ParseYaml("../../../configs/coinbase.yaml", &coinbase)

	if err != nil {
		t.Fatal(err)
	}

	channel := make(chan []observer.Ticker, 1000)
	monitor := NewMonitor("BTC-USD", 60*time.Second, 200, &channel, coinbase)

	err = monitor.gatherFrameOfHistorical()
	if err != nil {
		t.Fatal(err)
	}

	priceSize := len(monitor.prices.Raster())
	if 200 != priceSize {
		t.Fatalf("expected 200 Tickers, was %d", priceSize)
	}

	tickers := monitor.prices.Raster()

	expectedMinStart := time.Now().Add(-2 * monitor.Granularity).UTC()
	expectedMaxStart := time.Now().Add(2 * monitor.Granularity).UTC()
	if tickers[len(tickers)-1].Time.After(expectedMaxStart) || tickers[len(tickers)-1].Time.Before(expectedMinStart) {
		t.Fatalf("start time of historical frame is outside of expected range %s, start %s, end %s",
			tickers[len(tickers)-1].Time,
			expectedMinStart,
			expectedMaxStart)
	}

	for i, ticker := range tickers {
		if ticker.ProductId != "BTC-USD" {
			t.Fatalf("tickerId was expected to be BTC-USD and was %s", ticker.ProductId)
		}
		expectedTime := tickers[0].Time.Add(monitor.Granularity * time.Duration(i))
		if !expectedTime.Equal(ticker.Time) {
			t.Fatalf("expected timestamp to be %s but was %s, %v", expectedTime, ticker.Time, tickers)
		}
	}
}

func TestCreateCandleQuery(t *testing.T) {
	coinbase := Coinbase{}
	err := yaml.ParseYaml("../../../configs/coinbase.yaml", &coinbase)
	if err != nil {
		t.Fatal(err)
	}

	channel := make(chan []observer.Ticker, 1000)
	monitor := NewMonitor("BTC-USD", 60*time.Second, 200, &channel, coinbase)

	url, err := CreateCandleQuery(monitor)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(url)

	start, err := time.Parse(TimeFormat, url.Query().Get("start"))
	if err != nil {
		t.Fatal(err)
	}
	if start.After(time.Now()) {
		t.Fatal("expected start to be after now")
	}

	end, err := time.Parse(TimeFormat, url.Query().Get("end"))
	if err != nil {
		t.Fatal(err)
	}
	if end.Before(time.Now()) {
		t.Fatal("expected end to be after now")
	}

	if url.Query().Get("granularity") != "60" {
		t.Fatalf("expected granularity to be 10s was %s", url.Query().Get("granularity"))
	}

	segmentsOfCall := strings.Split(coinbase.Price.HistoricalPrice.Https, "%s")
	for _, segment := range segmentsOfCall {
		if !strings.Contains(url.String(), segment) {
			t.Fatalf("expected %s but was not found in url", segment)
		}
	}
}
