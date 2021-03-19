package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
)

const basePolygonAPI = "https://api.polygon.io/v2/ticks/stocks/trades"

// Response object
type Response struct {
	Ticker      string        `json:"ticker"`
	ResultCount int           `json:"results_count"`
	Results     []Trades      `json:"results"`
	KeyMap      interface{}   `json:"map"`
}

// Trades holds all trade data for ticker on date
type Trades struct {
	OriginalID    *int    `json:"I,omitempty"`
	ExchangeID    int     `json:"x"`
	Price         float64 `json:"p"`
	TradeID       string  `json:"i"`
	Correction    *int    `json:"e,omitempty"`
	ReportingID   *int     `json:"r,omitempty"`
	Timestamp     int     `json:"t"`
	ExchangeTime  int     `json:"y"`
	ReportingTime *int     `json:"f,omitempty"`
	Sequence      int     `json:"q"`
	Conditions    *[]int   `json:"c,omitempty"`
	Size          int     `json:"s"`
	Tape          int     `json:"z"`
}

// Complete list of the day's trades
var allDayTrades Response

func getTrades(ticker string, day string, timestamp int, apiKey string) ([]byte, error) {
	queryURL := fmt.Sprintf("%s/%s/%s?&limit=50000&apiKey=%s", basePolygonAPI, ticker, day, apiKey)

	if timestamp != 0 {
		queryURL = fmt.Sprintf("%s/%s/%s?&limit=50000&timestamp=%d&apiKey=%s", basePolygonAPI, ticker, day, timestamp, apiKey)
	}

	resp, err := http.Get(queryURL)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var responseObject Response
	json.Unmarshal(responseBody, &responseObject)

	allDayTrades.Results = append(allDayTrades.Results, responseObject.Results...)
	allDayTrades.ResultCount += responseObject.ResultCount
	
	for responseObject.ResultCount == 50000 {
		// Need to add 1 nanosecond to timestamp to avoid duplicates
		return getTrades(ticker, day, responseObject.Results[49999].Timestamp + 1, apiKey)
	}

	// Copy static data into the allDayTrades object
	allDayTrades.Ticker = responseObject.Ticker
	allDayTrades.KeyMap = responseObject.KeyMap

	dat, err := json.Marshal(allDayTrades)
	if err != nil {
		log.Fatal(err)
	}

	return dat, nil
}

// Compress the result and write the file to disk
// Should result in roughly 35% file size reduction
func compress(in []byte, compressionMap map[string][]byte) []byte {
	firstTest := make([]byte, len(in))
	copy(firstTest, in)
	for k, v := range compressionMap {
		commonSequence := []byte(k)
		firstTest = bytes.ReplaceAll(firstTest, commonSequence, v)
	}

	origFileSize := len(in)
	compressedFileSize := len(firstTest)
	compressionRatio := float64(compressedFileSize) / float64(origFileSize) * 100
	
	fmt.Printf("Original file size: %d  ;;  Compressed file size: %d  ;;  Compression ratio: %d%%\n", origFileSize, compressedFileSize, int(compressionRatio))

	return firstTest
}

// Decompress reverses the byte replacement in Compress
func decompress(filename string, compressionMap map[string][]byte) []byte {
	in, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	firstTest := make([]byte, len(in))
	copy(firstTest, in)

	for k, v := range compressionMap {
		commonSequence := []byte(k)
		firstTest = bytes.ReplaceAll(firstTest, v, commonSequence)
	}

	return firstTest
}

func main() {
	// Extremely naive map to compress common strings in polygon API response for AAPL ticker
	// Based on some code analysis to determine common sequences in AAPL trade files
	compressionMap := map[string][]byte{
		`,"z":3}`: []byte{128}, // AAPL is NASDAQ listed
		`{"x":`: []byte{129},
		`,"c":[`: []byte{130},
		`,"t":1`: []byte{131}, // All epochs since 2001 begin with 1
		`,"y":1`: []byte{132},
		`,"f":1`: []byte{133},
		`,"p":`: []byte{144},
		`,"i":`: []byte{145},
		`,"r":12`: []byte{146},
		`,"r":10`: []byte{147},
		`],"s":`: []byte{148},
		`,"s":`: []byte{149},
		`,"q":`: []byte{150},
		`12,37`: []byte{151},
	}

	compressCommand := flag.NewFlagSet("compress", flag.ExitOnError)
	decompressCommand := flag.NewFlagSet("decompress", flag.ExitOnError)

	// Compress subcommand flags
	dayPtr := compressCommand.String("day", "2020-11-13", "YYYY-MM-DD to get trades for.")
	outputPtr := compressCommand.String("o", "aapl_full_day.json.shahein", "Filename for compressed file.")
	apiKeyPtr := compressCommand.String("apiKey", "", "Your Polygon.io API Key with access to Trades REST interface (Required)")

	// Decompress subcommand flags
	fileLocationPtr := decompressCommand.String("f", "aapl_full_day.json.shahein", "File to decompress (Required)")
	outputLocationPtr := decompressCommand.String("o", "aapl_full_day.json", "Filename for decompressed file.")

	compressCommand.Usage = func() {
		fmt.Printf("Usage: %s compress [options]\nOptions:\n", os.Args[0])
		compressCommand.PrintDefaults()
	}

	decompressCommand.Usage = func() {
		fmt.Printf("Usage: %s decompress [options]\nOptions:\n", os.Args[0])
		decompressCommand.PrintDefaults()
	}

	if len(os.Args) < 2 {
		fmt.Println("compress or decompress subcommand is required")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "compress":
		compressCommand.Parse(os.Args[2:])
	case "decompress":
		decompressCommand.Parse(os.Args[2:])
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	if compressCommand.Parsed() {
		// Required flags without defaults
		if *apiKeyPtr == "" {
			fmt.Println("Polygon API key required")
			compressCommand.PrintDefaults()
			os.Exit(1)
		}

		// Date format validation
		re := regexp.MustCompile("((19|20)\\d\\d)-(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])")
		if (*dayPtr == "" || re.MatchString(*dayPtr) == false) {
			fmt.Println("Date is required and must be in YYYY-MM-DD format.")
			compressCommand.PrintDefaults()
			os.Exit(1)
		}

		dat, _ := getTrades("AAPL", *dayPtr, 0, *apiKeyPtr)

		c := compress(dat, compressionMap)
		if len(c) > 0 {
			_ = ioutil.WriteFile(*outputPtr, c, 0644)
		}
	}

	if decompressCommand.Parsed() {
		d := decompress(*fileLocationPtr, compressionMap)
		_ = ioutil.WriteFile(*outputLocationPtr, d, 0644)
		fmt.Println(*outputLocationPtr)
	}
}
