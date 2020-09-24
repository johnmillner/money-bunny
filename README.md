![Go](https://github.com/johnmillner/robo-macd/workflows/Go/badge.svg)


# Code In Progress - Don't trust your money with this yet :) 

# robo-macd
Simple MACD based Robo-Advisor


Inspired by the video here by [Trading Rush](https://www.youtube.com/watch?v=nmffSjdZbWQ):

[![MACD explanation video from Trading Rush](https://img.youtube.com/vi/nmffSjdZbWQ/0.jpg)](https://www.youtube.com/watch?v=nmffSjdZbWQ)

This is a MACD based robo-advisor that will buy and sell crypto from coinbase based solely on that crypto's MACD.

# To Use
... todo

# Structure

![Diagrams](/readme-static/diagrams.svg)

## Alpaca
[Alpaca](https://alpaca.markets/) is an API first, 0 commission broker that is used by this robo-advisor to interact with equity markets

## Gatherers
### Streamer
Streams in live prices from targeted equities (from configuration/UI) using Alpaca API
### Fetcher
fetches historical prices from targeted equities (from configuration/UI) using Alpaca API

## Transformer
Takes in the gatheres data and transforms to wanted structure (in this base case - MACD)

## Manager
Takes in the Transformed Data from the Transformers and determines for each targeted equity whether to buy/sell/hold that equity

## Executor
Takes in the Managers high level decisions and translates them into actions to take place in the broker, managing the low level portfolio

## UI
Provides ways to view the account, gains, trades, win/loss rate, and allows configuration for:
 * targeted equities and frequencies for the gathers
 * specific configurations for the transformer(s)
 * specific configurations for the executor(s)

