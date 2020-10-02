package gatherers

import (
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func TestInsert_Multiple(t *testing.T) {
	original := []Equity{{
		Name: "1",
	}, {
		Name: "4",
	}}

	added := []Equity{{
		Name: "2",
	}, {
		Name: "3",
	}}

	expected := []Equity{{
		Name: "1",
	}, {
		Name: "2",
	}, {
		Name: "3",
	}, {
		Name: "4",
	}}

	assert.Equal(t, expected, insert(original, 1, added...))
}

func TestInsert_Single(t *testing.T) {
	original := []Equity{{
		Name: "1",
	}, {
		Name: "3",
	}}

	added := Equity{
		Name: "2",
	}

	expected := []Equity{{
		Name: "1",
	}, {
		Name: "2",
	}, {
		Name: "3",
	}}

	assert.Equal(t, expected, insert(original, 1, added))
}

func TestInsert_Nothing(t *testing.T) {
	original := []Equity{{
		Name: "1",
	}, {
		Name: "2",
	}}

	expected := []Equity{{
		Name: "1",
	}, {
		Name: "2",
	}}

	assert.Equal(t, expected, insert(original, 1))
}

func TestInsert_Low(t *testing.T) {
	original := []Equity{{
		Name: "2",
	}, {
		Name: "3",
	}}

	added := Equity{
		Name: "1",
	}

	expected := []Equity{{
		Name: "1",
	}, {
		Name: "2",
	}, {
		Name: "3",
	}}

	assert.Equal(t, expected, insert(original, -1, added))
}

func TestInsert_High(t *testing.T) {
	original := []Equity{{
		Name: "1",
	}, {
		Name: "2",
	}}

	added := Equity{
		Name: "3",
	}

	expected := []Equity{{
		Name: "1",
	}, {
		Name: "2",
	}, {
		Name: "3",
	}}

	assert.Equal(t, expected, insert(original, 100, added))
}

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

func TestGatherer_ChunkList_NoChunkingNeeded(t *testing.T) {
	list := []string{"a", "b", "c"}
	assert.Equal(t, [][]string{list}, chunkList(list, 5))
}

func TestGatherer_ChunkList_Chunk(t *testing.T) {
	list := []string{"a", "b", "c"}
	assert.Equal(t, [][]string{{"a", "b"}, {"c"}}, chunkList(list, 2))
}

func TestGatherer_ChunkList_Empty(t *testing.T) {
	var expected [][]string
	assert.Equal(t, expected, chunkList([]string{}, 2))
}

func TestGatherer_ChunkList_Nil(t *testing.T) {
	var expected [][]string
	assert.Equal(t, expected, chunkList(nil, 2))
}

func TestGatherer_FilterByMarketOpen_Filters(t *testing.T) {
	times := MarketTimes{
		marketTimesMap: make(map[time.Time]timeRange),
	}

	times.marketTimesMap[time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)] = timeRange{
		start: time.Now().Add(-5 * time.Hour),
		end:   time.Now().Add(10 * time.Hour),
	}

	now := time.Now().Round(time.Minute)
	equities := filterByMarketOpen("test", []alpaca.Bar{{
		Time: now.Unix(),
	}, {
		Time: time.Now().Add(20 * time.Hour).Unix(),
	}}, &times)

	expected := []Equity{
		{
			Name:      "test",
			Time:      now,
			Open:      0,
			Close:     0,
			Low:       0,
			High:      0,
			Volume:    0,
			generated: false,
		},
	}

	log.Printf("%v", equities[0].Time)
	log.Printf("%v", now)
	assert.Equal(t, expected, equities)

}

func TestGatherer_FillGaps_FillsCuringMarketOn(t *testing.T) {
	times := MarketTimes{
		marketTimesMap: make(map[time.Time]timeRange),
	}

	times.marketTimesMap[time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(),
		0, 0, 0, 0, time.Local)] = timeRange{
		start: time.Now().Add(-5 * time.Hour),
		end:   time.Now().Add(5 * time.Hour),
	}

	now := time.Now().Round(time.Minute)
	equities := []Equity{
		{
			Name:      "test",
			Time:      now,
			High:      1,
			generated: false,
		}, {
			Name:      "test",
			Time:      now.Add(2 * time.Minute),
			High:      2,
			generated: false,
		},
	}

	expected := []Equity{
		{
			Name:      "test",
			Time:      now,
			High:      1,
			generated: false,
		}, {
			Name:      "test",
			Time:      now.Add(time.Minute),
			High:      1,
			generated: true,
		}, {
			Name:      "test",
			Time:      now.Add(2 * time.Minute),
			High:      2,
			generated: false,
		},
	}

	assert.Equal(t, expected, fillGaps(equities, &times, time.Minute))
}

func TestGatherer_FillGaps_DoesNotFillDuringAfterMarket(t *testing.T) {
	times := MarketTimes{
		marketTimesMap: make(map[time.Time]timeRange),
	}

	now := time.Now().Round(time.Minute)

	times.marketTimesMap[time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(),
		0, 0, 0, 0, time.Local)] = timeRange{
		start: now.Add(-5 * time.Hour),
		end:   now.Add(5 * time.Hour),
	}
	times.marketTimesMap[time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-1,
		0, 0, 0, 0, time.Local)] = timeRange{
		start: now.Add(-24 * time.Hour).Add(-5 * time.Hour),
		end:   now.Add(-24 * time.Hour).Add(5 * time.Hour),
	}

	equities := []Equity{
		{
			Name:      "right before MarketClose",
			Time:      now.Add((5 - 24) * time.Hour).Add(-1 * time.Minute),
			High:      1,
			generated: false,
		}, {
			Name:      "at marketOpen",
			Time:      now.Add(-5 * time.Hour),
			High:      2,
			generated: false,
		},
	}

	expected := []Equity{
		{
			Name:      "right before MarketClose",
			Time:      now.Add((5 - 24) * time.Hour).Add(-1 * time.Minute),
			High:      1,
			generated: false,
		}, {
			Name:      "at marketOpen",
			Time:      now.Add(-5 * time.Hour),
			High:      2,
			generated: false,
		},
	}

	assert.Equal(t, expected, fillGaps(equities, &times, time.Minute))
}

func TestGatherer_FillGaps_lastMinuteOfMarketMissingAndFills(t *testing.T) {
	times := MarketTimes{
		marketTimesMap: make(map[time.Time]timeRange),
	}

	now := time.Now().Round(time.Minute)

	times.marketTimesMap[time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(),
		0, 0, 0, 0, time.Local)] = timeRange{
		start: now.Add(-5 * time.Hour),
		end:   now,
	}
	times.marketTimesMap[time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-1,
		0, 0, 0, 0, time.Local)] = timeRange{
		start: now.Add(-24 * time.Hour).Add(-5 * time.Hour),
		end:   now.Add(-24 * time.Hour),
	}

	equities := []Equity{
		{
			Name:      "1 min before MarketClose",
			Time:      now.Add(-24 * time.Hour).Add(-2 * time.Minute),
			High:      1,
			generated: false,
		}, {
			Name:      "at marketOpen",
			Time:      now.Add(-5 * time.Hour),
			High:      2,
			generated: false,
		},
	}

	expected := []Equity{
		{
			Name:      "1 min before MarketClose",
			Time:      now.Add(-24 * time.Hour).Add(-2 * time.Minute),
			High:      1,
			generated: false,
		}, {
			Name:      "1 min before MarketClose",
			Time:      now.Add(-24 * time.Hour).Add(-1 * time.Minute),
			High:      1,
			generated: true,
		}, {
			Name:      "at marketOpen",
			Time:      now.Add(-5 * time.Hour),
			High:      2,
			generated: false,
		},
	}

	assert.Equal(t, expected, fillGaps(equities, &times, time.Minute))
}

func mockGetBars(_ GathererConfig, _ []string, _ int) (map[string][]alpaca.Bar, error) {
	bars := make(map[string][]alpaca.Bar)
	bars["TSLA"] = []alpaca.Bar{{
		Time: time.Now().Add(-3 * time.Minute).Round(time.Minute).Unix(),
		Open: 1,
	}, {
		Time: time.Now().Add(-1 * time.Minute).Round(time.Minute).Unix(),
		Open: 3,
	}}
	return bars, nil

	//todo
}

func TestGather_GatherPage_Success(t *testing.T) {
	equities := gatherPage([]string{"TSLA"}, GathererConfig{
		Client: *alpaca.NewClient(common.Credentials()),
		Period: time.Minute,
	}, mockGetBars)

	log.Printf("todo %v", equities)

}
