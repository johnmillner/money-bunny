package monitor

import (
	"reflect"
	"testing"
	"time"
)

func TestPriceMonitor_UpdatePrice(t *testing.T) {
	channel := make(chan []Ticker, 5)

	monitor := NewMonitor("BTC-USD", 60*time.Second, 2, &channel, Coinbase{})

	t1 := Ticker{
		ProductId: "BTC-USD",
		Price:     1,
		Time:      time.Now().Round(time.Minute).UTC(),
	}
	t2 := Ticker{
		ProductId: "BTC-USD",
		Price:     2,
		Time:      time.Now().Add(time.Minute).Round(time.Minute).UTC(),
	}
	t3 := Ticker{
		ProductId: "BTC-USD",
		Price:     3,
		Time:      time.Now().Add(time.Minute).Add(time.Second * 30).UTC(),
	}
	t4 := Ticker{
		ProductId: "BTC-USD",
		Price:     4,
		Time:      time.Now().Add(time.Minute * 2).Round(time.Minute).UTC(),
	}

	monitor.UpdatePrice(t1)
	monitor.UpdatePrice(t2)
	if !reflect.DeepEqual(monitor.prices.Raster(), []Ticker{t1, t2}) {
		t.Fatalf("expected %v, got %v", []Ticker{t1, t2}, monitor.prices.Raster())
	}
	monitor.UpdatePrice(t3)
	if !reflect.DeepEqual(monitor.prices.Raster(), []Ticker{t1, t2}) {
		t.Fatalf("expected %v, got %v", []Ticker{t1, t2}, monitor.prices.Raster())
	}
	monitor.UpdatePrice(t4)
	if !reflect.DeepEqual(monitor.prices.Raster(), []Ticker{t2, t4}) {
		t.Fatalf("expected %v, got %v", []Ticker{t2, t4}, monitor.prices.Raster())
	}
}
