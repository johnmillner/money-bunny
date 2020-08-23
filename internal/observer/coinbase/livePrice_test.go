package coinbase

import (
	"github.com/johnmillner/robo-macd/internal/observer"
	"github.com/johnmillner/robo-macd/internal/yaml"
	"testing"
	"time"
)

func TestPriceMonitor_PopulateLive(t *testing.T) {
	coinbase := Coinbase{}
	err := yaml.ParseYaml("../../../configs/coinbase.yaml", &coinbase)
	if err != nil {
		t.Fatal(err)
	}

	channel := make(chan []observer.Ticker, 1000)
	monitor := NewMonitor("BTC-USD", 5*time.Second, 5, &channel, coinbase)

	go monitor.PopulateLive()

	counter := 0
	for tickers := range channel {
		counter++
		t.Logf("current status \r\n %v", tickers)
		for i, ticker := range tickers {
			if ticker.ProductId != "BTC-USD" {
				t.Fatalf("tickerId was expected to be BTC-USD and was %s", ticker.ProductId)
			}
			expectedTime := tickers[0].Time.Add(monitor.Granularity * time.Duration(i))
			if !expectedTime.Equal(ticker.Time) {
				t.Fatalf("expected timestamp to be %s but was %s, %v", expectedTime, ticker.Time, tickers)
			}
		}

		if counter >= 11 {
			break
		}
	}

}
