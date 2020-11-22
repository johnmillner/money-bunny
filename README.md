![Go](https://github.com/johnmillner/robo-macd/workflows/Go/badge.svg)


# Code In Progress - Don't trust your money with this yet :) 

# Robo Scalper
Robo-Advisor that will trade stocks based off of MACD crossover

Strategy inspired by the video here by [Trading Rush](https://www.youtube.com/watch?v=nmffSjdZbWQ):

# Configuration
in the config.yml add your alpaca key, secret, and your (live) alpaca key again to polygon key (or if you have a polygon subscription directly, that)
Alternatively, these values can be stored as environment variables at 
- RS_ALPACA.KEY
- RS_ALPACA.SECRET
- RS_POLYGON.KEY

additional configurations can be found inside of config/config.yml to control risk, reward, and stock selection
equally these can be overriden by environment variables with the prefix `RS_`

## Alpaca
[Alpaca](https://alpaca.markets/) is an API first, 0 commission broker that is used by this robo-trader to interact with equity markets

## Polygon
[Polygon](https://polygon.io) is a Low latency market data API where live ticker data is streamed in - access to this is included with a live-trading alpaca account for free

# stocks
recomend to choose about 10 stocks with high volatility, volume, and market cap