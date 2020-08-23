package observer

import (
	"errors"
	"time"
)

type Ticker struct {
	ProductId string    `json:"product_id"`
	Price     float64   `json:"price,string"`
	Time      time.Time `json:"time"`
}

// Ouroboros contains a ring buffer style stack that can have a Rasterized array of the current state printed out
type Ouroboros struct {
	Capacity  int
	beginning int
	stack     []Ticker
}

func (o Ouroboros) IsFull() bool {
	return len(o.stack) >= o.Capacity
}

func (o Ouroboros) Push(ticker Ticker) Ouroboros {
	stack := o.stack

	if !(len(o.stack) >= o.Capacity) {
		return Ouroboros{
			Capacity:  o.Capacity,
			beginning: 0,
			stack:     append(o.stack, ticker),
		}
	}

	stack[o.beginning] = ticker

	return Ouroboros{
		Capacity:  o.Capacity,
		beginning: (o.beginning + 1) % o.Capacity,
		stack:     stack,
	}
}

func (o Ouroboros) Peek() (Ticker, error) {
	if len(o.stack) < 1 {
		return Ticker{}, errors.New("cannot peek because stack the stack is smaller than 1")
	}

	return o.stack[(o.beginning+len(o.stack)-1)%o.Capacity], nil
}

// Raster copies the current low level data-structure into
// an array where index 0 is the beginning of the queue
func (o Ouroboros) Raster() []Ticker {
	if len(o.stack) <= 0 {
		return []Ticker{}
	}
	return append(o.stack[o.beginning:], o.stack[:o.beginning]...)
}

func NewOuroboros(capacity int) Ouroboros {
	return Ouroboros{
		Capacity:  capacity,
		beginning: 0,
		stack:     make([]Ticker, 0),
	}
}
