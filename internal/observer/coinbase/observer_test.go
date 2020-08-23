package coinbase

import (
	"github.com/johnmillner/robo-macd/internal/observer"
	"github.com/johnmillner/robo-macd/internal/yaml"
	"log"
	"reflect"
	"testing"
	"time"
)

func TestPriceMonitor_UpdatePrice(t *testing.T) {
	channel := make(chan []observer.Ticker, 100)

	monitor := NewMonitor("BTC-USD", 60*time.Second, 2, &channel, Coinbase{})

	t1 := observer.Ticker{
		ProductId: "BTC-USD",
		Price:     1,
		Time:      time.Now().Round(time.Minute).UTC(),
	}
	t2 := observer.Ticker{
		ProductId: "BTC-USD",
		Price:     2,
		Time:      time.Now().Add(time.Minute).Round(time.Minute).UTC(),
	}
	t3 := observer.Ticker{
		ProductId: "BTC-USD",
		Price:     3,
		Time:      time.Now().Add(time.Minute).Add(time.Second * 30).UTC(),
	}
	t4 := observer.Ticker{
		ProductId: "BTC-USD",
		Price:     4,
		Time:      time.Now().Add(time.Minute * 2).Round(time.Minute).UTC(),
	}
	t45 := observer.Ticker{
		ProductId: "BTC-USD",
		Price:     4,
		Time:      time.Now().Add(time.Minute * 4).Round(time.Minute).UTC(),
	}
	t5 := observer.Ticker{
		ProductId: "BTC-USD",
		Price:     6,
		Time:      time.Now().Add(time.Minute * 5).Round(time.Minute).UTC(),
	}
	t56 := observer.Ticker{
		ProductId: "BTC-USD",
		Price:     6,
		Time:      time.Now().Add(time.Minute * 6).Round(time.Minute).UTC(),
	}
	t6 := observer.Ticker{
		ProductId: "BTC-USD",
		Price:     7,
		Time:      time.Now().Add(time.Minute * 7).Round(time.Minute).UTC(),
	}

	monitor.UpdatePrice(t1)
	monitor.UpdatePrice(t2)
	if !reflect.DeepEqual(monitor.prices.Raster(), []observer.Ticker{t1, t2}) {
		t.Fatalf("expected %v, got %v", []observer.Ticker{t1, t2}, monitor.prices.Raster())
	}
	monitor.UpdatePrice(t3)
	if !reflect.DeepEqual(monitor.prices.Raster(), []observer.Ticker{t1, t2}) {
		t.Fatalf("expected %v, got %v", []observer.Ticker{t1, t2}, monitor.prices.Raster())
	}
	monitor.UpdatePrice(t4)
	if !reflect.DeepEqual(monitor.prices.Raster(), []observer.Ticker{t2, t4}) {
		t.Fatalf("expected %v, got %v", []observer.Ticker{t2, t4}, monitor.prices.Raster())
	}
	monitor.UpdatePrice(t5)
	if !reflect.DeepEqual(monitor.prices.Raster(), []observer.Ticker{t45, t5}) {
		t.Fatalf("expected %v, got %v", []observer.Ticker{t45, t5}, monitor.prices.Raster())
	}
	monitor.UpdatePrice(t6)
	if !reflect.DeepEqual(monitor.prices.Raster(), []observer.Ticker{t56, t6}) {
		t.Fatalf("expected %v, got %v", []observer.Ticker{t56, t6}, monitor.prices.Raster())
	}
}

func TestPriceMonitor_Initialize(t *testing.T) {
	coinbase := Coinbase{}
	err := yaml.ParseYaml("../../../configs\\coinbase.yaml", &coinbase)
	if err != nil {
		log.Fatal(err)
	}

	channel := make(chan []observer.Ticker, 1000)
	NewMonitor("BTC-USD", 60*time.Second, 2, &channel, coinbase).Initialize()
}
