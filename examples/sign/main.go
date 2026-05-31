package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/khizar/bc-lib/internal/encoding"
	"github.com/khizar/bc-lib/pkg/chains"
	"github.com/khizar/bc-lib/pkg/transaction"
	"github.com/khizar/bc-lib/pkg/types"
	"github.com/khizar/bc-lib/pkg/wallet"
)

// This example demonstrates the full transaction signing flow:
//
//  1. Create (or import) a wallet.
//  2. Build an EIP-1559 transaction.
//  3. Sign it locally with the wallet's private key.
//  4. Inspect the raw, broadcast-ready transaction bytes.
//  5. Recover the sender from the signature to verify correctness.
//
// No network connection is required - everything runs locally.
func main() {
	// Import a deterministic key so the output is reproducible. In production
	// use wallet.NewWallet() to generate a fresh key, or FromPrivateKeyHex to
	// load one from a secret store.
	w, err := wallet.FromPrivateKeyHex("4646464646464646464646464646464646464646464646464646464646464646")
	if err != nil {
		log.Fatalf("failed to load wallet: %v", err)
	}
	fmt.Printf("Wallet address: %s\n", w.Address())

	to := types.MustAddressFromHex("0x3535353535353535353535353535353535353535")

	// Build an EIP-1559 transfer of 1 ETH.
	tx, err := transaction.Transfer(to, types.EtherToWei(big.NewInt(1)), chains.EthereumMainnet).
		Nonce(9).
		MaxFeePerGas(big.NewInt(30_000_000_000)).        // 30 gwei
		MaxPriorityFeePerGas(big.NewInt(2_000_000_000)). // 2 gwei
		Build()
	if err != nil {
		log.Fatalf("failed to build transaction: %v", err)
	}

	// Sign the transaction locally.
	signed, err := w.SignTransaction(tx)
	if err != nil {
		log.Fatalf("failed to sign transaction: %v", err)
	}

	fmt.Printf("\nSigned transaction:\n")
	fmt.Printf("  Hash:  %s\n", signed.Hash)
	fmt.Printf("  From:  %s\n", signed.From)
	fmt.Printf("  Raw:   0x%s\n", hex.EncodeToString(signed.RawTransaction))

	// Recover the signer to confirm the signature is valid.
	recovered, err := encoding.Sender(signed.Transaction)
	if err != nil {
		log.Fatalf("failed to recover sender: %v", err)
	}
	fmt.Printf("\nRecovered signer: %s (matches wallet: %v)\n", recovered, recovered == w.Address())

	// To broadcast, connect a client and call SendRawTransaction or use the
	// high-level SignAndSendTransaction helper:
	//
	//   bcClient, _ := client.New(config)
	//   bcClient.Connect(ctx, "ethereum")
	//   hash, _ := bcClient.SendRawTransaction(ctx, "ethereum", signed.RawTransaction)
	//
	//   // or, populate gas/nonce, sign and send in one call:
	//   hash, _ := bcClient.SignAndSendTransaction(ctx, "ethereum", w, tx)
}
