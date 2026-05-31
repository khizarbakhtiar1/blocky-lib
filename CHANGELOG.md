# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial project setup and architecture
- Core types: Address, Hash, Transaction, Block, Receipt, Log
- Unit conversion utilities (Wei, Gwei, Ether)
- Chain interface for blockchain implementations
- Ethereum client with JSON-RPC support
- RPC provider with automatic failover
- Multi-chain client orchestration
- Transaction builder with fluent API
- Wallet functionality for key management
- Comprehensive test suite
- Balance, transfer, ERC-20, and signing examples
- Getting started documentation

### Cryptography
- Real **secp256k1** curve implementation (replacing the earlier P-256 stand-in),
  producing Ethereum-valid keys and addresses
- Deterministic **RFC 6979** ECDSA signing with low-S normalization (EIP-2) and
  recovery-id computation
- Public-key recovery (`Ecrecover`/`SigToPub`) and signature verification
- Validated against published vectors (private-key→address vectors and the
  EIP-155 signed-transaction example, byte-for-byte)

### Transaction Signing
- RLP serialization for signing and broadcasting
- Legacy (EIP-155), EIP-2930 (access list), and EIP-1559 transaction support
- `wallet.SignTransaction` returns broadcast-ready raw bytes
- `Sender` recovery to derive the signer address from a signed transaction
- `ethereum.SendTransaction` broadcasts signed transactions
- High-level client helpers: `PopulateTransaction` and `SignAndSendTransaction`

### Supported Chains
- Ethereum (Mainnet, Sepolia)
- Polygon (Mainnet, Mumbai)
- Arbitrum (One, Sepolia)
- Optimism (Mainnet, Sepolia)
- Base (Mainnet, Sepolia)

### Supported Operations
- Get balance (current and historical)
- Get block (by number/hash) with full transaction decoding
- Get transaction and transaction receipt
- Get transaction count (nonce)
- Estimate gas, suggest gas price / tip cap (EIP-1559)
- Sign, send, and send-raw transactions
- Contract calls (`eth_call`) and ERC-20 helpers
- Event log filtering (`FilterLogs`) and polling-based subscriptions (`SubscribeToLogs`)
- Network ID and sync status queries

### Coming Soon
- Full ABI encoding/decoding
- WebSocket transport for real-time events
- Batch operations and benchmarks
- ERC-721 helpers
- Constant-time / hardened signing
- Account abstraction (ERC-4337)
- Non-EVM chains (e.g. Solana)

## [0.1.0] - TBD

Initial MVP release (planned).


