package alpaca_wrapper

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGatherer_DurationToTimeframe(t *testing.T) {
	assert.Equal(t, "1Min", durationToTimeframe(time.Minute))
	assert.Equal(t, "5Min", durationToTimeframe(5*time.Minute))
	assert.Equal(t, "15Min", durationToTimeframe(15*time.Minute))
	assert.Equal(t, "1H", durationToTimeframe(time.Hour))
	assert.Equal(t, "1D", durationToTimeframe(24*time.Hour))

	assert.Panics(t, func() {
		_ = durationToTimeframe(2 * time.Minute)
	})
}
