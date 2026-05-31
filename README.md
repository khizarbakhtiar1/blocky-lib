# bc-lib: Multi-Chain Blockchain SDK for Golang

A unified, developer-friendly Go library for interacting with multiple EVM blockchain networks.

## Vision

Provide Go developers with a simple, type-safe, and performant SDK for building multi-chain blockchain applications without the complexity of managing multiple chain-specific libraries.

## Key Features

- **Multi-Chain Support**: Ethereum, Polygon, Arbitrum, Optimism, Base, and other EVM chains
- **Unified Interface**: A single API for all supported blockchains
- **Real secp256k1 Cryptography**: Ethereum-correct key generation, address derivation, signing, and public-key recovery
- **Transaction Signing**: Full RLP encoding and signing for legacy (EIP-155), EIP-2930, and EIP-1559 transactions
- **Type Safety**: Leverages Go's type system for safe blockchain interactions
- **RPC Management**: Automatic failover and retry across multiple RPC endpoints
- **ERC-20 Helpers**: Read token metadata/balances and build transfer/approve calldata
- **Developer Experience**: Comprehensive examples, clear error messages, and a fluent transaction builder

## Installation

```bash
go get github.com/khizar/bc-lib
```

Requires Go 1.24+.

## Quick Start

### Query a balance

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
    bcClient, err := client.New(client.Config{
        Chains: map[string]chains.ChainConfig{
            "ethereum": {
                ChainID: chains.EthereumMainnet,
                RPCURLs: []string{"https://eth.llamarpc.com"},
            },
            "polygon": {
                ChainID: chains.PolygonMainnet,
                RPCURLs: []string{"https://polygon-rpc.com"},
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer bcClient.Close()

    ctx := context.Background()
    if err := bcClient.ConnectAll(ctx); err != nil {
        log.Fatal(err)
    }

    addr := types.MustAddressFromHex("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")
    balance, _ := bcClient.GetBalance(ctx, "ethereum", addr)
    fmt.Printf("Balance: %s ETH\n", types.FormatEther(balance))
}
```

### Sign and send a transaction

```go
import (
    "github.com/khizar/bc-lib/pkg/transaction"
    "github.com/khizar/bc-lib/pkg/wallet"
)

// Load a wallet (use wallet.NewWallet() to generate a fresh key).
w, _ := wallet.FromPrivateKeyHex("0x...")

// Build an EIP-1559 transfer.
tx, _ := transaction.Transfer(toAddress, value, chains.EthereumMainnet).
    MaxFeePerGas(big.NewInt(30_000_000_000)).
    MaxPriorityFeePerGas(big.NewInt(2_000_000_000)).
    Build()

// Populate nonce/gas, sign, and broadcast in one call.
hash, err := bcClient.SignAndSendTransaction(ctx, "ethereum", w, tx)

// ...or sign locally and broadcast the raw bytes yourself:
signed, _ := w.SignTransaction(tx)
hash, err = bcClient.SendRawTransaction(ctx, "ethereum", signed.RawTransaction)
```

## Roadmap

### Phase 1: Foundation ✅
- [x] Core client architecture
- [x] Ethereum mainnet support
- [x] Basic RPC operations (balance, transactions, blocks)
- [x] Transaction building and signing
- [x] Comprehensive test suite
- [x] Initial documentation

### Phase 2: Multi-Chain ✅
- [x] Polygon support
- [x] Arbitrum support
- [x] Optimism support
- [x] Base support
- [x] RPC provider failover
- [x] Connection management

### Phase 3: Advanced Features (in progress)
- [x] Event log filtering (`FilterLogs`)
- [x] Log subscriptions via polling (`SubscribeToLogs`)
- [x] Smart contract calls (`CallContract`)
- [x] ERC-20 helpers
- [ ] Full ABI encoding/decoding
- [ ] WebSocket transport for real-time data
- [ ] Batch operations

### Phase 4: Ecosystem (planned)
- [ ] ERC-721 helpers
- [ ] Account abstraction (ERC-4337) support
- [ ] CLI tool for common operations
- [ ] Performance benchmarks and hardening (constant-time signing)
- [ ] Additional non-EVM chains (e.g. Solana, Cosmos)

## Architecture

```
bc-lib/
├── pkg/                 # Public API
│   ├── client/          # Multi-chain client and orchestration
│   ├── chains/          # Chain interface + EVM (ethereum) implementation
│   ├── transaction/     # Transaction builder and confirmation waiter
│   ├── wallet/          # Wallet and key management
│   ├── tokens/          # ERC-20 helpers
│   ├── types/           # Common types (Address, Hash, Transaction, ...)
│   ├── errors/          # Typed errors
│   └── logging/         # Structured logging
├── internal/            # Implementation details
│   ├── crypto/          # secp256k1, Keccak-256, ECDSA signing/recovery
│   ├── encoding/        # RLP and transaction serialization
│   └── rpc/             # JSON-RPC provider with failover
└── examples/            # Runnable examples (balance, transfer, erc20, sign)
```

## Security Notes

- The secp256k1 implementation is written in pure Go using `math/big`. It is
  correct and produces Ethereum-valid keys, addresses, and signatures, but it
  is **not constant-time** and is not hardened against side-channel attacks.
  Use it in trusted environments. Hardening (or delegating to an audited
  native library) is tracked on the roadmap.
- Private keys never leave your process. Signing is performed locally; only the
  signed transaction bytes are sent over the network.

## Contributing

Contributions, ideas, and feedback are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License — see [LICENSE](LICENSE).

---

**Status**: Phase 1 & 2 complete; Phase 3 in progress.
