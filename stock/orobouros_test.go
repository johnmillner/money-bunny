package stock

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestNewOuroboros(t *testing.T) {
	expected := []float64{1, 2, 3}
	result := NewOuroboros(expected)
	assert.Equal(t, result.stack, expected)
	assert.Equal(t, result.Capacity, len(expected))
	assert.Equal(t, result.beginning, 0)
}

func TestOuroboros_Peek(t *testing.T) {
	expected := []float64{1, 2, 3}
	result := NewOuroboros(expected)

	assert.Equal(t, result.Peek(), float64(3))
}

func TestOuroboros_Push(t *testing.T) {
	expected := []float64{1, 2, 3}
	result := NewOuroboros(expected).Push(4)

	assert.Equal(t, result.stack, []float64{4, 2, 3})
	assert.Equal(t, result.Capacity, len(expected))
	assert.Equal(t, result.beginning, 1)

	assert.Equal(t, result.Peek(), float64(4))
}

func TestOuroboros_Raster(t *testing.T) {
	expected := []float64{1, 2, 3}
	result := NewOuroboros(expected).Push(4)

	assert.Equal(t, result.Raster(), []float64{2, 3, 4})
}

func TestOuroboros_RasterEmpty(t *testing.T) {
	assert.Equal(t, NewOuroboros([]float64{}).Raster(), []float64{})
}
