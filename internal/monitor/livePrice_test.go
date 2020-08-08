package monitor

import (
	"github.com/johnmillner/robo-macd/internal/yaml"
	"testing"
	"time"
)

func TestPriceMonitor_PopulateLive(t *testing.T) {
	coinbase := Coinbase{}
	err := yaml.ParseYaml("../../configs\\coinbase.yaml", &coinbase)
	if err != nil {
		t.Fatal(err)
	}

	channel := make(chan []Ticker, 1000)
	monitor := NewMonitor("BTC-USD", 1*time.Second, 10, &channel, coinbase)

	go monitor.PopulateLive()

	counter := 0
	for tickers := range channel {

		counter++
		for i, ticker := range tickers {
			if ticker.ProductId != "BTC-USD" {
				t.Fatalf("tickerId was expected to be BTC-USD and was %s", ticker.ProductId)
			}
			expectedTime := tickers[0].Time.UTC().Add(monitor.Granularity * time.Duration(i))
			if !expectedTime.Equal(ticker.Time) {
				t.Fatalf("expected timestamp to be %s but was %s, %v", expectedTime, ticker.Time, tickers)
			}
		}

		if counter >= 21 {
			break
		}
	}

}
