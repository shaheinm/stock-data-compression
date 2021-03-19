# Polygon.io API Data Compression

## What Is It
A tiny CLI that does naive, lossless compression for AAPL trade data from the [Polygon.io Trades API](https://polygon.io/docs/get_v2_ticks_stocks_trades__ticker___date__anchor). Compression ratio is only ~65%, but the LZW-inspired dictionary idea should be prevalent. 

The compression ratio could be improved with more analysis of the results from the API (i.e. better understanding of prevalent patterns, or even better - a good separator that I can split the string on). I originally created the dictionary dynamically by splitting on commas, creating about 128 keys, but the compression ratio was worse. I also could have stolen one of the various implementations of `zstandard`, et al out in the wild (`zstd` gave me a ~19% compression ratio).

There are no external dependencies. 

## How To Use It
Build the binary:
```
$ go build trades.go
```

Check out the help text:
```
$ ./trades compress --help
Usage: ./trades <compress> [options]
Options:
  -apiKey string
        Your Polygon.io API Key with access to Trades REST interface (Required)
  -day string
        YYYY-MM-DD to get trades for (Default is my birthday) (default "2020-11-13")
  -o string
        Filename for compressed file. (default "aapl_full_day.json.shahein")

$ ./trades decompress --help
Usage: ./trades decompress [options]
Options:
  -f string
        File to decompress (Required) (default "aapl_full_day.json.shahein")
  -o string
        Filename for decompressed file. (default "aapl_full_day.json")
```

Compress will display a file size comparison and compression ratio:
```
$ ./trades compress --apiKey=superSecretApiKey
Original file size: 62746687  ;;  Compressed file size: 40295068  ;;  Compression ratio: 64%
```

Decompress will display the filename of the decompressed file:
```
$ ./trades decompress
aapl_full_day.json
```

There are also some basic tests for the compress and decompress functions, if you want to run them (i.e. `go test`).

## Releases

Just for fun, I also cut a release (v0.0.1), so you can do the following:
```
$ go get -u github.com/shaheinm/stock-data-compression
$ stock-data-compression compress -day="2021-03-17" -o="aapl_20210317_trades.json.shahein" -apiKey=supersecretapikey
$ stock-data-compression decompress -f="aapl_20210317_trades.json.shahein" -o="aapl_20210317_trades.json"
```
