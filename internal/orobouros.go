package internal

// Ouroboros contains a ring buffer style stack that can have a Rasterized array of the current state printed out
type Ouroboros struct {
	Capacity  int
	beginning int
	stack     []float64
}

func (o Ouroboros) Push(ticker float64) Ouroboros {
	stack := o.stack
	stack[o.beginning] = ticker

	return Ouroboros{
		Capacity:  o.Capacity,
		beginning: (o.beginning + 1) % o.Capacity,
		stack:     stack,
	}
}

func (o Ouroboros) Peek() float64 {
	return o.stack[(o.beginning+len(o.stack)-1)%o.Capacity]
}

// Raster copies the current low level data-structure into
// an array where index 0 is the beginning of the queue
func (o Ouroboros) Raster() []float64 {
	if len(o.stack) <= 0 {
		return []float64{}
	}
	return append(o.stack[o.beginning:], o.stack[:o.beginning]...)
}

func NewOuroboros(start []float64) Ouroboros {
	return Ouroboros{
		Capacity:  len(start),
		beginning: 0,
		stack:     start,
	}
}
