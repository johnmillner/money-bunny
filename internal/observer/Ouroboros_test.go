package observer

import (
	"reflect"
	"testing"
	"time"
)

func TestRasterizingStack_IsFull_edgeFull(t *testing.T) {
	r := NewOuroboros(0)
	if !r.IsFull() {
		t.Fail()
	}
}

func TestRasterizingStack_IsFull_full(t *testing.T) {
	r := NewOuroboros(1)
	r.Push(Ticker{
		ProductId: "1",
	})

	if !r.IsFull() {
		t.Fail()
	}
}

func TestRasterizingStack_IsFull_empty(t *testing.T) {
	r := NewOuroboros(1)
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
	t5 := Ticker{
		ProductId: "5",
		Price:     0,
		Time:      time.Time{},
	}

	r := NewOuroboros(2)

	r.Push(t1)
	expected := make([]Ticker, 1)
	expected[0] = t1
	if !reflect.DeepEqual(expected, r.Raster()) {
		t.Fatalf("expected %v, got %v", expected, r.Raster())
	}

	r.Push(t2)
	expected = make([]Ticker, 2)
	expected[0] = t1
	expected[1] = t2
	if !reflect.DeepEqual(expected, r.Raster()) {
		t.Fatalf("expected %v, got %v", expected, r.Raster())
	}

	r.Push(t3)
	expected[0] = t2
	expected[1] = t3
	if !reflect.DeepEqual(expected, r.Raster()) {
		t.Fatalf("expected %v, got %v", expected, r.Raster())
	}

	r.Push(t4)
	expected[0] = t3
	expected[1] = t4
	if !reflect.DeepEqual(expected, r.Raster()) {
		t.Fatalf("expected %v, got %v", expected, r.Raster())
	}

	r.Push(t5)
	expected[0] = t4
	expected[1] = t5
	if !reflect.DeepEqual(expected, r.Raster()) {
		t.Fatalf("expected %v, got %v", expected, r.Raster())
	}
}

func TestRasterizingStack_Peek(t *testing.T) {
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

	r := NewOuroboros(2)

	result, err := r.Peek()
	if err == nil {
		t.Fatal("should receive an error when peeking at empty stack")
	}
	r.Push(t1)
	result, err = r.Peek()
	if err != nil {
		t.Fatal(err)
	}
	if result != t1 {
		t.Fatalf("expected %v, got %v", t1, result)
	}

	r.Push(t2)
	result, err = r.Peek()
	if err != nil {
		t.Fatal(err)
	}
	if result != t2 {
		t.Fatalf("expected %v, got %v", t2, result)
	}

	r.Push(t3)
	r.Push(t4)
	result, err = r.Peek()
	if err != nil {
		t.Fatal(err)
	}
	if result != t4 {
		t.Fatalf("expected %v, got %v", t4, result)
	}
}
