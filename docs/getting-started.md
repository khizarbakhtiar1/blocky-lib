# Getting Started with bc-lib

## Installation

```bash
go get github.com/khizar/bc-lib
```

## Quick Start

### 1. Create a Client

```go
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
    // Configure your chains
    config := client.Config{
        Chains: map[string]chains.ChainConfig{
            "ethereum": {
                Name:    "Ethereum Mainnet",
                ChainID: chains.EthereumMainnet,
                RPCURLs: []string{
                    "https://eth.llamarpc.com",
                    "https://ethereum-rpc.publicnode.com",
                },
                Timeout: 30,
            },
        },
    }

    // Create the client
    bcClient, err := client.New(config)
    if err != nil {
        log.Fatal(err)
    }
    defer bcClient.Close()

    // Connect to the chain
    ctx := context.Background()
    if err := bcClient.Connect(ctx, "ethereum"); err != nil {
        log.Fatal(err)
    }

    fmt.Println("Connected to Ethereum!")
}
```

### 2. Query Balances

```go
// Parse an address
address := types.MustAddressFromHex("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")

// Get balance in wei
balance, err := bcClient.GetBalance(ctx, "ethereum", address)
if err != nil {
    log.Fatal(err)
}

// Convert to ether
etherBalance := types.WeiToEther(balance)
fmt.Printf("Balance: %s ETH\n", etherBalance)
```

### 3. Build Transactions

```go
import (
    "github.com/khizar/bc-lib/pkg/transaction"
    "math/big"
)

// Build a simple transfer
tx, err := transaction.Transfer(
    toAddress,
    types.EtherToWei(big.NewInt(1)), // 1 ETH
    chains.EthereumMainnet,
).
    Nonce(0).
    MaxFeePerGas(big.NewInt(30000000000)).        // 30 gwei
    MaxPriorityFeePerGas(big.NewInt(2000000000)). // 2 gwei
    Build()

if err != nil {
    log.Fatal(err)
}

// Estimate gas
estimatedGas, err := bcClient.EstimateGas(ctx, "ethereum", tx)
```

### 4. Work with Multiple Chains

```go
config := client.Config{
    Chains: map[string]chains.ChainConfig{
        "ethereum": {
            ChainID: chains.EthereumMainnet,
            RPCURLs: []string{"https://eth.llamarpc.com"},
        },
        "polygon": {
            ChainID: chains.PolygonMainnet,
            RPCURLs: []string{"https://polygon-rpc.com"},
        },
        "arbitrum": {
            ChainID: chains.ArbitrumOne,
            RPCURLs: []string{"https://arb1.arbitrum.io/rpc"},
        },
    },
}

bcClient, _ := client.New(config)

// Connect to all chains
bcClient.ConnectAll(ctx)

// Query same address on multiple chains
for _, chainName := range bcClient.ChainNames() {
    balance, _ := bcClient.GetBalance(ctx, chainName, address)
    fmt.Printf("%s: %s\n", chainName, types.FormatEther(balance))
}
```

## Core Concepts

### Addresses

```go
// From hex string
addr, err := types.AddressFromHex("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")

// Must version (panics on error - use for constants)
addr := types.MustAddressFromHex("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")

// Get string representation
fmt.Println(addr.String()) // "0x742d35..."

// Check if zero address
if addr.IsZero() {
    fmt.Println("Zero address!")
}
```

### Hashes

```go
// Transaction or block hash
hash, err := types.HashFromHex("0x1234...")

// Convert to string
fmt.Println(hash.String())
```

### Units

```go
// Convert between units
wei := types.EtherToWei(big.NewInt(1))
gwei := types.WeiToGwei(wei)
ether := types.WeiToEther(wei)

// Parse ether from string
amount, err := types.ParseEther("1.5") // Returns wei

// Format wei as ether
formatted := types.FormatEther(wei) // Returns string like "1.500000..."
```

### Transaction Builder

```go
// Simple transfer
tx := transaction.Transfer(toAddr, value, chainID).
    Nonce(nonce).
    MaxFeePerGas(maxFee).
    MaxPriorityFeePerGas(priorityFee).
    Build()

// Contract call
tx := transaction.ContractCall(contractAddr, data, chainID).
    Nonce(nonce).
    GasLimit(100000).
    Build()

// Contract deployment
tx := transaction.ContractDeploy(bytecode, chainID).
    Nonce(nonce).
    GasLimit(2000000).
    Build()
```

## Common Patterns

### Error Handling

```go
balance, err := bcClient.GetBalance(ctx, "ethereum", address)
if err != nil {
    log.Printf("Failed to get balance: %v", err)
    return
}
```

### Context with Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

balance, err := bcClient.GetBalance(ctx, "ethereum", address)
```

### RPC Failover

The library automatically handles RPC failover when you provide multiple URLs:

```go
RPCURLs: []string{
    "https://primary-rpc.com",
    "https://backup-rpc.com",
    "https://fallback-rpc.com",
},
```

If one fails, it automatically tries the next.

## Examples

See the [examples](../examples/) directory for complete examples:

- [Balance Query](../examples/balance/) - Query address balance
- [Multi-Chain](../examples/transfer/) - Work with multiple chains
- More examples coming soon!

## Next Steps

- Read the [API documentation](./API.md)
- Check out [examples](../examples/)
- Learn about [supported chains](./CHAINS.md)
- See the [roadmap](../README.md#roadmap)

## Need Help?

- Open an issue on GitHub
- Check the examples
- Read the documentation
- See the test files for usage patterns

