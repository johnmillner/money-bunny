package monitor

import (
	"reflect"
	"testing"
	"time"
)

func TestRasterizingStack_IsFull_edgeFull(t *testing.T) {
	r := NewRasterizingStack(0)
	if !r.IsFull() {
		t.Fail()
	}
}

func TestRasterizingStack_IsFull_full(t *testing.T) {
	r := NewRasterizingStack(1)
	err := r.Push(Ticker{
		ProductId: "1",
	})

	if err != nil {
		t.Error(err)
	}

	if !r.IsFull() {
		t.Fail()
	}
}

func TestRasterizingStack_IsFull_empty(t *testing.T) {
	r := NewRasterizingStack(1)
	if r.IsFull() {
		t.Fail()
	}
}

func TestRasterizingStack_Push(t *testing.T) {
	t1 := Ticker{
		ProductId: "1",
		Price:     0,
		Time:      time.Time{},
	}
	t2 := Ticker{
		ProductId: "2",
		Price:     0,
		Time:      time.Time{},
	}
	t3 := Ticker{
		ProductId: "3",
		Price:     0,
		Time:      time.Time{},
	}
	t4 := Ticker{
		ProductId: "4",
		Price:     0,
		Time:      time.Time{},
	}

	r := NewRasterizingStack(2)

	_ = r.Push(t1)
	expected := [2]Ticker{t1, nil}
	if !reflect.DeepEqual(expected, r.rasterize()) {
		t.Fail()
	}

	_ = r.Push(t1)
	expected = [2]Ticker{t1, t2}
	if !reflect.DeepEqual(expected, r.rasterize()) {
		t.Fail()
	}

	_ = r.Push(t1)
	expected = [2]Ticker{t2, t3}
	if !reflect.DeepEqual(expected, r.rasterize()) {
		t.Fail()
	}

	_ = r.Push(t1)
	expected = [2]Ticker{t3, t4}
	if !reflect.DeepEqual(expected, r.rasterize()) {
		t.Fail()
	}
}
