package macd

import "github.com/johnmillner/robo-macd/internal/observer"

type RoboConfig struct {
	MacdCalculator struct {
		Period float64 `yaml:"period"`
		Macd   struct {
			Twelve    int `yaml:"12-period-ema"`
			TwentySix int `yaml:"26-period-ema"`
			Nine      int `yaml:"9-period-ema"`
		} `yaml:"macd"`
		Trend struct {
			TrendEmaPeriod int `yaml:"trend-ema-period"`
		} `yaml:"trend"`
	} `yaml:"macd-calculator"`
}

type Macd struct {
	Product string
	Signal  []float64
	Macd    []float64
	Trend   []float64
}

func NewMacd(tickers []observer.Ticker, roboConfig RoboConfig) Macd {
	values := make([]float64, len(tickers))
	for i, value := range tickers {
		values[i] = value.Price
	}

	return Macd{
		Product: tickers[0].ProductId,
		Macd:    calculateMacd(values, roboConfig.MacdCalculator.Macd.Twelve, roboConfig.MacdCalculator.Macd.TwentySix),
		Signal:  calculateSignal(values, roboConfig.MacdCalculator.Macd.Nine),
		Trend:   calculateTrend(values, roboConfig.MacdCalculator.Trend.TrendEmaPeriod),
	}
}

func calculateTrend(values []float64, trendPeriod int) []float64 {
	return calculateTrend(values, trendPeriod)
}

func calculateMacd(values []float64, twelve int, twentySix int) []float64 {
	twelveEma := calculateEma(values, twelve)
	twentySixEma := calculateEma(values, twentySix)

	macd := make([]float64, len(twentySixEma))
	for i := range twentySixEma {
		macd[i] = twelveEma[i+(twentySix-twelve)] - twentySixEma[i]
	}

	return macd
}

func calculateSignal(values []float64, nine int) []float64 {
	return calculateEma(values, nine)
}

func calculateEma(values []float64, period int) []float64 {
	return nil
}

func calculateSma(values []float64, period int) []float64 {
	return nil
}
