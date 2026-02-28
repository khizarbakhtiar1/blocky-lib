package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/khizar/bc-lib/pkg/chains"
	"github.com/khizar/bc-lib/pkg/client"
	"github.com/khizar/bc-lib/pkg/transaction"
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
				},
				Timeout: 30,
			},
			"polygon": {
				Name:    "Polygon Mumbai",
				ChainID: chains.PolygonMumbai,
				RPCURLs: []string{
					"https://rpc-mumbai.maticvigil.com",
				},
				Timeout: 30,
			},
		},
	}

	bcClient, err := client.New(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer bcClient.Close()

	ctx := context.Background()

	// Connect to all chains
	if err := bcClient.ConnectAll(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	fmt.Println("Connected to multiple chains!")

	// Example: Check balance on multiple chains
	address := types.MustAddressFromHex("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")

	for _, chainName := range bcClient.ChainNames() {
		balance, err := bcClient.GetBalance(ctx, chainName, address)
		if err != nil {
			log.Printf("Failed to get balance on %s: %v", chainName, err)
			continue
		}

		etherBalance := types.WeiToEther(balance)
		fmt.Printf("%s balance: %s\n", chainName, etherBalance)
	}

	// Example: Build a transaction (not signed, just constructed)
	fmt.Println("\nBuilding a transaction...")

	toAddress := types.MustAddressFromHex("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	value := types.EtherToWei(big.NewInt(1)) // 1 ETH

	tx, err := transaction.Transfer(toAddress, value, chains.EthereumSepolia).
		Nonce(0).
		MaxFeePerGas(big.NewInt(30000000000)).        // 30 gwei
		MaxPriorityFeePerGas(big.NewInt(2000000000)). // 2 gwei
		Build()

	if err != nil {
		log.Fatalf("Failed to build transaction: %v", err)
	}

	fmt.Printf("Transaction built successfully!\n")
	fmt.Printf("  To: %s\n", tx.To)
	fmt.Printf("  Value: %s wei\n", tx.Value)
	fmt.Printf("  Gas Limit: %d\n", tx.GasLimit)
	fmt.Printf("  Type: %d (EIP-1559)\n", tx.Type)

	// Estimate gas for the transaction
	estimatedGas, err := bcClient.EstimateGas(ctx, "ethereum", tx)
	if err != nil {
		log.Printf("Failed to estimate gas: %v", err)
	} else {
		fmt.Printf("  Estimated Gas: %d\n", estimatedGas)
	}

	fmt.Println("\nNote: To actually send this transaction, you would need to:")
	fmt.Println("  1. Sign it with a wallet's private key")
	fmt.Println("  2. Use SendRawTransaction to broadcast it")
}
