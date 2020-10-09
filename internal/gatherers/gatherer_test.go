package gatherers

import (
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/google/uuid"
	"github.com/johnmillner/robo-macd/internal/alpaca_wrapper"
	"github.com/johnmillner/robo-macd/internal/coordinator"
	"github.com/johnmillner/robo-macd/internal/utils"
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

func TestGather_GatherPage_Success(t *testing.T) {
	equities := gatherPage([]string{"TSLA"}, GathererConfig{
		Period: time.Minute,
		Alpaca: alpaca_wrapper.MockedAlpaca{
			Bars:     alpaca_wrapper.MockGetBars,
			Calendar: alpaca_wrapper.MockCalander,
		},
	})

	log.Printf("%v", equities)

	assert.Equal(t, 1, len(equities))
	assert.Equal(t, 3, len(equities[0]))
	assert.True(t, equities[0][1].generated)
}

func TestGather_GatherPage_Failure(t *testing.T) {
	assert.Panics(t, func() {
		gatherPage([]string{"TSLA"}, GathererConfig{
			Period: time.Minute,
			Alpaca: alpaca_wrapper.MockedAlpaca{
				Bars:     alpaca_wrapper.MockGetBarsFail,
				Calendar: alpaca_wrapper.MockCalander,
			},
		})
	})
}

func TestGather_Gather_OverLimit(t *testing.T) {
	g := Gatherer{}
	output := make(chan []Equity)
	g.gather(GathererConfig{
		EquityData: output,
		Symbols:    []string{"TSLA"},
		Period:     time.Minute,
		Limit:      2,
		Alpaca: alpaca_wrapper.MockedAlpaca{
			Bars:     alpaca_wrapper.MockGetBars,
			Calendar: alpaca_wrapper.MockCalander,
		},
	})

	equities := <-output

	assert.Equal(t, 2, len(equities))
}

func TestGather_Gather_UnderLimit(t *testing.T) {
	g := Gatherer{}
	output := make(chan []Equity)
	g.gather(GathererConfig{
		EquityData: output,
		Symbols:    []string{"TSLA"},
		Period:     time.Minute,
		Limit:      100,
		Alpaca: alpaca_wrapper.MockedAlpaca{
			Bars:     alpaca_wrapper.MockGetBars,
			Calendar: alpaca_wrapper.MockCalander,
		},
	})

	equities := <-output

	assert.Equal(t, 3, len(equities))
}

func TestGather_InitGather_firstPage(t *testing.T) {
	output := make(chan []Equity)

	InitGatherer(utils.Configurator{
		Me:        uuid.UUID{},
		Module:    "gatherer",
		ConfigIn:  make(chan utils.Config),
		ConfigOut: nil,
		Config: GathererConfig{
			Active:     true,
			EquityData: output,
			Symbols:    []string{"TSLA"},
			Period:     time.Minute,
			Limit:      100,
			Alpaca: alpaca_wrapper.MockedAlpaca{
				Bars:     alpaca_wrapper.MockGetBars,
				Calendar: alpaca_wrapper.MockCalander,
			},
		},
	})

	equities := <-output

	assert.Equal(t, 3, len(equities))
}

func TestGather_InitGather_Shutdown(t *testing.T) {
	configIn := make(chan utils.Config)
	configOut := make(chan utils.Config)

	InitGatherer(utils.Configurator{
		Me:        uuid.UUID{},
		Module:    "gatherer",
		ConfigIn:  configIn,
		ConfigOut: configOut,
		Config:    GathererConfig{},
	})

	messageOut := <-configOut
	casted, ok := messageOut.(coordinator.Shutdown)
	assert.True(t, ok)
	assert.True(t, casted.Shutdown)
}

func TestGather_InitGather_firstPageThenShutdown(t *testing.T) {
	output := make(chan []Equity)
	configIn := make(chan utils.Config)
	configOut := make(chan utils.Config)

	InitGatherer(utils.Configurator{
		Me:        uuid.UUID{},
		Module:    "gatherer",
		ConfigIn:  configIn,
		ConfigOut: configOut,
		Config: GathererConfig{
			Active:     true,
			EquityData: output,
			Symbols:    []string{"TSLA"},
			Period:     time.Minute,
			Limit:      100,
			Alpaca: alpaca_wrapper.MockedAlpaca{
				Bars:     alpaca_wrapper.MockGetBars,
				Calendar: alpaca_wrapper.MockCalander,
			},
		},
	})

	equities := <-output

	assert.Equal(t, 3, len(equities))

	configIn <- GathererConfig{}

	messageOut := <-configOut
	casted, ok := messageOut.(coordinator.Shutdown)
	assert.True(t, ok)
	assert.True(t, casted.Shutdown)
}

func TestGathererConfig_GetToFrom(t *testing.T) {
	to := uuid.New()
	from := uuid.New()
	gc := GathererConfig{
		To:   to,
		From: from,
	}

	assert.Equal(t, to, gc.GetTo())
	assert.Equal(t, from, gc.GetFrom())
}
