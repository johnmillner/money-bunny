package monitor

import "errors"

type RasterizingStack struct {
	capacity  int
	beginning int
	stack     []Ticker
}

func (c *RasterizingStack) IsFull() bool {
	for _, ticker := range c.stack {
		if ticker == (Ticker{}) {
			return false
		}
	}

	return true
}

func (c *RasterizingStack) Push(ticker Ticker) error {
	if ticker == (Ticker{}) {
		return errors.New("cannot add an empty ticker")
	}

	c.stack[(c.beginning+len(c.stack))%c.capacity] = ticker

	if c.IsFull() {
		c.beginning = (c.beginning + 1) % c.capacity
	}

	return nil
}

func (c *RasterizingStack) Pop() {

}

func (c *RasterizingStack) Peek() Ticker {
	return c.stack[(c.beginning+len(c.stack))%c.capacity]
}

func NewRasterizingStack(capacity int) RasterizingStack {
	return RasterizingStack{
		capacity:  capacity,
		beginning: 0,
		stack:     make([]Ticker, capacity),
	}
}

// rasterize copies the current low level data-structure into
// an array where index 0 is the beginning of the queue
func (c *RasterizingStack) Rasterize() []Ticker {
	return c.stack
}
