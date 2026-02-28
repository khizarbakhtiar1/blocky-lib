package main

import (
	"context"
	"fmt"
	"log"

	"github.com/khizar/bc-lib/pkg/chains"
	"github.com/khizar/bc-lib/pkg/client"
	"github.com/khizar/bc-lib/pkg/types"
)

func main() {
	// Create a new multi-chain client
	config := client.Config{
		Chains: map[string]chains.ChainConfig{
			"ethereum": {
				Name:    "Ethereum Sepolia",
				ChainID: chains.EthereumSepolia,
				RPCURLs: []string{
					"https://ethereum-sepolia-rpc.publicnode.com",
					"https://sepolia.drpc.org",
				},
				Timeout: 30,
			},
		},
	}

	// Initialize the client
	bcClient, err := client.New(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer bcClient.Close()

	// Connect to Ethereum Sepolia
	ctx := context.Background()
	if err := bcClient.Connect(ctx, "ethereum"); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	fmt.Println("✅ Connected to Ethereum Sepolia!")

	// Example address (Vitalik's address)
	address := types.MustAddressFromHex("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")
	
	// Get balance
	balance, err := bcClient.GetBalance(ctx, "ethereum", address)
	if err != nil {
		log.Fatalf("Failed to get balance: %v", err)
	}

	// Convert to ether and display
	etherBalance := types.WeiToEther(balance)
	fmt.Printf("Address: %s\n", address)
	fmt.Printf("Balance: %s ETH\n", etherBalance)
	fmt.Printf("Balance: %s wei\n", balance)

	// Get latest block number
	blockNumber, err := bcClient.GetBlockNumber(ctx, "ethereum")
	if err != nil {
		log.Fatalf("Failed to get block number: %v", err)
	}
	fmt.Printf("Latest block: %s\n", blockNumber)

	// Get nonce for the address
	nonce, err := bcClient.GetNonce(ctx, "ethereum", address)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}
	fmt.Printf("Nonce: %d\n", nonce)

	// Get suggested gas price
	gasPrice, err := bcClient.SuggestGasPrice(ctx, "ethereum")
	if err != nil {
		log.Fatalf("Failed to get gas price: %v", err)
	}
	gasPriceGwei := types.WeiToGwei(gasPrice)
	fmt.Printf("Gas price: %s gwei\n", gasPriceGwei)
}

