package main

import (
	"context"
	"fmt"
	"log"

	"github.com/khizar/bc-lib/pkg/chains"
	"github.com/khizar/bc-lib/pkg/client"
	"github.com/khizar/bc-lib/pkg/tokens"
	"github.com/khizar/bc-lib/pkg/types"
)

func main() {
	// Create a client connected to Ethereum mainnet
	config := client.Config{
		Chains: map[string]chains.ChainConfig{
			"ethereum": {
				Name:    "Ethereum Mainnet",
				ChainID: chains.EthereumMainnet,
				RPCURLs: []string{
					"https://eth.llamarpc.com",
					"https://ethereum.publicnode.com",
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

	// Connect to Ethereum
	if err := bcClient.Connect(ctx, "ethereum"); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	fmt.Println("✅ Connected to Ethereum Mainnet!")

	// Get the Ethereum chain
	ethChain, err := bcClient.GetChain("ethereum")
	if err != nil {
		log.Fatalf("Failed to get chain: %v", err)
	}

	// Example 1: Get USDT token info
	fmt.Println("\n📊 USDT Token Info:")
	usdtAddr, _ := tokens.GetTokenAddress("USDT", 1)
	usdt := tokens.NewERC20(usdtAddr, ethChain)

	name, err := usdt.Name(ctx)
	if err != nil {
		log.Printf("Failed to get name: %v", err)
	} else {
		fmt.Printf("  Name: %s\n", name)
	}

	symbol, err := usdt.Symbol(ctx)
	if err != nil {
		log.Printf("Failed to get symbol: %v", err)
	} else {
		fmt.Printf("  Symbol: %s\n", symbol)
	}

	decimals, err := usdt.Decimals(ctx)
	if err != nil {
		log.Printf("Failed to get decimals: %v", err)
	} else {
		fmt.Printf("  Decimals: %d\n", decimals)
	}

	totalSupply, err := usdt.TotalSupply(ctx)
	if err != nil {
		log.Printf("Failed to get total supply: %v", err)
	} else {
		formatted := tokens.FormatUnits(totalSupply, decimals)
		fmt.Printf("  Total Supply: %s USDT\n", formatted)
	}

	// Example 2: Check USDT balance of a whale address
	fmt.Println("\n💰 USDT Balance Check:")
	whaleAddr := types.MustAddressFromHex("0x28C6c06298d514Db089934071355E5743bf21d60") // Binance 14

	balance, err := usdt.BalanceOf(ctx, whaleAddr)
	if err != nil {
		log.Printf("Failed to get balance: %v", err)
	} else {
		formatted := tokens.FormatUnits(balance, decimals)
		fmt.Printf("  Address: %s\n", whaleAddr)
		fmt.Printf("  Balance: %s USDT\n", formatted)
	}

	// Example 3: Get USDC token info
	fmt.Println("\n📊 USDC Token Info:")
	usdcAddr, _ := tokens.GetTokenAddress("USDC", 1)
	usdc := tokens.NewERC20(usdcAddr, ethChain)

	usdcSymbol, _ := usdc.Symbol(ctx)
	usdcDecimals, _ := usdc.Decimals(ctx)
	fmt.Printf("  Symbol: %s\n", usdcSymbol)
	fmt.Printf("  Decimals: %d\n", usdcDecimals)

	// Example 4: Build a transfer calldata (without executing)
	fmt.Println("\n📝 Building Transfer Calldata:")
	toAddress := types.MustAddressFromHex("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	amount, _ := tokens.ParseUnits("100", 6) // 100 USDT (6 decimals)

	transferData := usdt.TransferData(toAddress, amount)
	fmt.Printf("  Transfer 100 USDT to %s\n", toAddress)
	fmt.Printf("  Calldata: 0x%x\n", transferData[:20])
	fmt.Printf("  (truncated, full length: %d bytes)\n", len(transferData))

	// Example 5: Build an approval calldata
	fmt.Println("\n📝 Building Approval Calldata:")
	spender := types.MustAddressFromHex("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D") // Uniswap V2 Router
	maxAmount, _ := tokens.ParseUnits("1000000", 6) // 1M USDT approval

	approveData := usdt.ApproveData(spender, maxAmount)
	fmt.Printf("  Approve Uniswap to spend 1M USDT\n")
	fmt.Printf("  Calldata: 0x%x\n", approveData[:20])
	fmt.Printf("  (truncated, full length: %d bytes)\n", len(approveData))

	fmt.Println("\n✨ Example complete!")
	fmt.Println("\n💡 Note: This example only reads data. To execute transfers,")
	fmt.Println("   you would need to:")
	fmt.Println("   1. Create a wallet with a private key")
	fmt.Println("   2. Build a transaction with the calldata")
	fmt.Println("   3. Sign and send the transaction")
}
