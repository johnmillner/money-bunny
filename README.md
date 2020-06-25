# robo-macd
Simple MACD based Robo-Advisor

Inspired by the video here: https://www.youtube.com/watch?v=nmffSjdZbWQ

This is a MACD based robo-advisor that will buy and sell crypto from coinbase based solely on that crypto's MACD.

# configuration
## configs/coinbaseAuth.yml
this is a secure file that contains your own coinbase api authentication details. There are two values - it is on the gitignore list for safety.
```yaml
key: test
secret: test
```
you can retrieve these details by ...todo

# configs/config.yml
configuration file that details stoploss percentage, profit target, macd and trend settings, etc.
