package config

import (
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
)

type Config struct {
	EthereumNode       string
	HighValueThreshold float64
	WatchedTokens      []common.Address
	WatchedContracts   []common.Address
	MongoURI           string
	MongoUser          string
	MongoPassword      string
	DatabaseName       string
	ServerPort         string
	RateLimitPerSecond float64
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	threshold, _ := strconv.ParseFloat(os.Getenv("HIGH_VALUE_THRESHOLD"), 64)
	rateLimit, _ := strconv.ParseFloat(os.Getenv("RATE_LIMIT_PER_SECOND"), 64)

	return &Config{
		EthereumNode:       os.Getenv("ETHEREUM_NODE"),
		HighValueThreshold: threshold,
		MongoURI:           os.Getenv("MONGO_URI"),
		MongoUser:          os.Getenv("MONGO_USER"),
		MongoPassword:      os.Getenv("MONGO_PASSWORD"),
		DatabaseName:       os.Getenv("MONGO_DATABASE"),
		ServerPort:         os.Getenv("SERVER_PORT"),
		RateLimitPerSecond: rateLimit,
		WatchedTokens: []common.Address{
			common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"), // DAI
			common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606EB48"), // USDC
		},
		WatchedContracts: []common.Address{
			common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), // WETH
			common.HexToAddress("0x5d3a536E4D6DbD6114cc1Ead35777bAB948E3643"), // Compound cDAI
		},
	}, nil
}
