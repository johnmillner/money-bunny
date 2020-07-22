package monitor

import (
	"errors"
	"fmt"
)

// RasterStack contains a ring buffer style stack that can have a Rasterized array of the current state printed out
type RasterStack struct {
	capacity  int
	beginning int
	stack     []Ticker
}

func (c *RasterStack) IsFull() bool {
	return len(c.stack) >= c.capacity
}

func (c *RasterStack) Push(ticker Ticker) {
	if !c.IsFull() {
		c.stack = append(c.stack, ticker)
		return
	}

	c.stack[c.beginning] = ticker
	c.beginning = (c.beginning + 1) % c.capacity
}

func (c *RasterStack) Pop() {
	if len(c.stack) <= 0 {
		return
	}
	c.stack = c.Raster()
	c.stack = c.stack[:len(c.stack)-1]
}

func (c *RasterStack) Peek(back int) (Ticker, error) {
	if len(c.stack) < 1 {
		return Ticker{}, errors.New(fmt.Sprintf("cannot peek because stack the stack is smaller than %d", back))
	}

	return c.stack[(c.beginning+len(c.stack)-back)%c.capacity], nil
}

// Raster copies the current low level data-structure into
// an array where index 0 is the beginning of the queue
func (c *RasterStack) Raster() []Ticker {
	if len(c.stack) <= 0 {
		return []Ticker{}
	}
	return append(c.stack[c.beginning:], c.stack[:c.beginning]...)
}

func NewRasterStack(capacity int) RasterStack {
	return RasterStack{
		capacity:  capacity,
		beginning: 0,
		stack:     make([]Ticker, 0),
	}
}
