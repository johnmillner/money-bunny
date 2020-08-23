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
	err := yaml.ParseYaml("../../../configs\\coinbase.yaml", &coinbase)

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

	for i, ticker := range monitor.prices.Raster() {
		if ticker.ProductId != "BTC-USD" {
			t.Fatalf("tickerId was expected to be BTC-USD and was %s", ticker.ProductId)
		}
		if !time.Now().Round(time.Minute).UTC().Add(time.Minute * time.Duration(-1*(monitor.prices.Capacity-i))).Equal(ticker.Time.UTC()) {
			t.Fatalf("expected timestamp to be %s but was %s for %v", time.Now().UTC().Round(time.Minute).Add(time.Minute*time.Duration(-1*(monitor.prices.Capacity-i))), ticker.Time.UTC(), monitor.prices.Raster())
		}
	}
}

func TestCreateCandleQuery(t *testing.T) {
	coinbase := Coinbase{}
	err := yaml.ParseYaml("../../../configs\\coinbase.yaml", &coinbase)
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
