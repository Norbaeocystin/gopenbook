package gopenbook

import (
	"context"
	"fmt"
	serumgo "github.com/gagliardetto/serum-go"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"log"
)

// Settle transaction
func Settle(m Market, openOrders, tokenBaseAddress, tokenQuoteAddress solana.PublicKey, wallet solana.PrivateKey) (solana.Signature, error) {
	i := serumgo.NewSettleFundsInstruction(
		m.Address,
		openOrders,
		wallet.PublicKey(),
		m.Metadata.BaseVault,
		m.Metadata.QuoteVault,
		tokenBaseAddress,
		tokenQuoteAddress,
		m.VaultSigner,
		solana.TokenProgramID,
		solana.PublicKey{},
	).Build()
	b, _ := i.Data()
	settle := solana.NewInstruction(
		solana.MustPublicKeyFromBase58(OpenBookAddress),
		i.Accounts()[:9],
		b,
	)
	log.Println(len(i.Accounts()[:9]))
	recent, err := m.Client.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	tx, err := solana.NewTransaction(
		[]solana.Instruction{settle},
		recent.Value.Blockhash,
		solana.TransactionPayer(wallet.PublicKey()),
	)
	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if wallet.PublicKey().Equals(key) {
				return &wallet
			}
			return nil
		},
	)
	if err != nil {
		panic(fmt.Errorf("unable to sign transaction: %w", err))
	}
	// Pretty print the transaction:
	// spew.Dump(tx)

	// Send transaction, and wait for confirmation:
	sig, err := m.Client.SendTransactionWithOpts(context.TODO(), tx,
		rpc.TransactionOpts{
			Encoding:            "",
			SkipPreflight:       false,
			PreflightCommitment: rpc.CommitmentProcessed,
			MaxRetries:          nil,
			MinContextSlot:      nil,
		},
	)
	log.Println("settling", sig, err)
	return sig, err
}
