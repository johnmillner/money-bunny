package gatherers

import (
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"log"
	"time"
)

func durationToTimeframe(dur time.Duration) string {
	switch dur {
	case time.Minute:
		return "1Min"
	case time.Minute * 5:
		return "5Min"
	case time.Minute * 15:
		return "15Min"
	case time.Hour * 24:
		return "1D"
	default:
		log.Fatalf("cannot translate duration given to alpaca timeframe, given: %f (in seconds) "+
			"- only acceptable durations are 1min, 5min, 15min, 1day", dur.Seconds())
		return dur.String()
	}
}

func Request(timeRanges []timeRange, config GathererConfig) map[string][]alpaca.Bar {
	log.Printf("attempting request")
	// determine request range and break out symbols
	var symbols []string
	start := time.Unix(1<<63-62135596801, 999999999)
	end := time.Time{}
	for _, timeRange := range timeRanges {
		symbols = append(symbols, timeRange.equity)

		//find minimum startDate
		if timeRange.start.Before(start) {
			start = timeRange.start
		}
		//find maximum endDate
		if timeRange.end.After(end) {
			end = timeRange.end
		}
	}

	log.Printf("start %s, end %s", start.String(), end.String())

	// determine the number of dataPoints between start and end
	dataPoints := end.Sub(start).Seconds() / config.Period.Seconds()
	maxDataPoints := 1000
	log.Printf("attempting to gather %v data points", dataPoints)

	totalValues := make(map[string][]alpaca.Bar)
	for i := int64(0); i < int64(dataPoints); i = i + int64(maxDataPoints) {
		// determine this requests limit
		thisStart := start.Add(time.Duration(i) * config.Period)
		thisEnd := end.Add(time.Duration(i+int64(maxDataPoints)-1) * config.Period)

		// generate request
		log.Printf("grabbing page of data")
		values, err := config.Client.ListBars(symbols, alpaca.ListBarParams{
			Timeframe: durationToTimeframe(config.Period),
			StartDt:   &thisStart,
			EndDt:     &thisEnd,
			Limit:     &maxDataPoints,
		})

		if err != nil {
			log.Printf("could not gather bars from alpaca due to %s", err)
			return nil
		}

		for symbol, bars := range values {
			totalValues[symbol] = append(totalValues[symbol], bars...)
		}
	}

	return totalValues
}
