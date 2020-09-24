![Go](https://github.com/johnmillner/robo-macd/workflows/Go/badge.svg)


# Code In Progress - Don't trust your money with this yet :) 

# Stock Conveyor 
Robo-Advisor that will trade stocks based off of MACD crossover


Strategy inspired by the video here by [Trading Rush](https://www.youtube.com/watch?v=nmffSjdZbWQ):

[![MACD explanation video from Trading Rush](https://img.youtube.com/vi/nmffSjdZbWQ/0.jpg)](https://www.youtube.com/watch?v=nmffSjdZbWQ)

# Concept
This program will scan the stock market - determining which stocks are having a MACD Crossover event. 
When it determines that a stock is crossing over positively, it will initiate a buy order - and equally when it determines 
a held stock is crossing over negatively, it will initiate a sell order. 
Multiple time scans can be run in parallel such that there are essentially conveyor belts of opportunities available. 

For Pattern Day Trading reasons, the example is configured to not buy/sell an equity on the same day. 

There could be a conveyor belt fo 1hr period, 4hr period, 1 day, etc periods all running in parallel playing short 
term swing trades as well as longer term plays - with different stocks bought (entering the conveyor belt) and sold (exiting the conveyor belt) asynchronously

# To Use
... todo

# Structure
[Diagrams found in draw.io](https://app.diagrams.net/?lightbox=1&highlight=0000ff&edit=_blank&layers=1&nav=1&title=RoboAdvisor#Uhttps%3A%2F%2Fdrive.google.com%2Fuc%3Fid%3D1fZWEaOWSyaqYmPYYk0OZuidXkcBH2hcp%26export%3Ddownload)

## Alpaca
[Alpaca](https://alpaca.markets/) is an API first, 0 commission broker that is used by this robo-advisor to interact with equity markets

## Gatherers
### Price Streamer
Streams in live prices from targeted equities (from configuration/UI) using Alpaca API
### Price Fetcher
fetches historical prices from targeted equities (from configuration/UI) using Alpaca API

## Transformer
Takes in the gatherer's data and transforms to wanted structure (in this base case - MACD)
### MACD Transformer
groups and transforms the raw price data from the gathers into a MACD chart

## Manager
Takes in the Transformed Data from the Transformers and determines for each targeted equity whether to buy/sell/hold that equity
### MACD Manager
takes in MACD Transformers chart and determines whether to buy, sell, or hold the stock at this moment

## Executor
Takes in the Managers high level decisions and translates them into actions to take place in the broker, managing the low level portfolio

## Coordinator
acts as the "main" of the program, instantiating the service and acting as a coordination hub for configurations

## UI
Provides ways to view the account, gains, trades, win/loss rate, and allows for the configuration of:
 * targeted equities and frequencies for the gathers
 * specific configurations for the gatherer(s)
 * specific configurations for the transformer(s)
 * specific configurations for the managers(s)
 * specific configurations for the executor

