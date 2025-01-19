package config

import "github.com/ethereum/go-ethereum/common"

// Config holds all application configuration
type Config struct {
	EthereumNodes      []string         // List of RPC node URLs for redundancy
	HighValueThreshold float64          // Minimum value to trigger high value alert
	WatchedTokens      []common.Address // Token contracts to monitor
	WatchedContracts   []common.Address // Other contracts to monitor
	MongoURI           string           // MongoDB connection URI
	MongoUser          string           // MongoDB username
	MongoPassword      string           // MongoDB password
	DatabaseName       string           // MongoDB database name
}

// NewDefaultConfig returns a config with default values
func NewDefaultConfig() *Config {
	return &Config{
		EthereumNodes: []string{
			"wss://ethereum.publicnode.com",
			"wss://mainnet.gateway.tenderly.co",
			"wss://rpc.ankr.com/eth/ws",
			// you can add more RPCs here, access chainlist.org
		},
		HighValueThreshold: 100,
		WatchedTokens: []common.Address{
			common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"), // DAI
			common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606EB48"), // USDC
		},
		WatchedContracts: []common.Address{
			common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), // WETH
			common.HexToAddress("0x5d3a536E4D6DbD6114cc1Ead35777bAB948E3643"), // Compound cDAI
		},
		MongoURI:      "mongodb://localhost:27017",
		MongoUser:     "root",
		MongoPassword: "example",
		DatabaseName:  "ethereum_monitor",
	}
}
