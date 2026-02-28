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
- Comprehensive test suite for types
- Balance query example
- Multi-chain transfer example
- Getting started documentation

### Supported Chains
- Ethereum (Mainnet, Sepolia)
- Polygon (Mainnet, Mumbai)
- Arbitrum (One, Sepolia)
- Optimism (Mainnet, Sepolia)
- Base (Mainnet, Sepolia)

### Supported Operations
- Get balance (current and historical)
- Get block number
- Get transaction count (nonce)
- Estimate gas
- Suggest gas price
- Suggest gas tip cap (EIP-1559)
- Send raw transaction
- Network ID query

### Coming Soon
- Complete transaction signing with RLP encoding
- Event log filtering and subscriptions
- Contract interaction (ABI support)
- ERC20/ERC721 token helpers
- WebSocket support for real-time events
- Solana support
- Account abstraction (ERC-4337)

## [0.1.0] - TBD

Initial MVP release (planned).


