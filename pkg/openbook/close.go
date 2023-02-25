package openbook

import (
	"context"
	"fmt"
	serumgo "github.com/gagliardetto/serum-go"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"log"
)

func Close(clientId uint64, openOrdersAccount solana.PublicKey, wallet solana.PrivateKey, m Market) (solana.Signature, error) {
	i := serumgo.NewCancelOrderByClientIdV2Instruction(
		clientId,
		m.Address,
		m.Metadata.Bids,
		m.Metadata.Asks,
		openOrdersAccount,
		wallet.PublicKey(),
		m.Metadata.EventQueue,
	).Build()
	recent, err := m.Client.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	tx, err := solana.NewTransaction(
		[]solana.Instruction{i},
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

func CloseSettleAndSwap(jupIn []solana.Instruction, clientOrderId uint64, openOrdersAccount, tokenBaseAddress, tokenQuoteAddress solana.PublicKey, wallet solana.PrivateKey, m Market) (solana.Signature, error) {
	instructions := make([]solana.Instruction, 0)
	// compute budget instruction
	i0 := solana.NewInstruction(COMPUTE_BUDGET,
		[]*solana.AccountMeta{},
		// fee 1, u
		[]uint8{0, 32, 161, 7, 0, 1, 0, 0, 0},
	)
	instructions = append(instructions, i0)
	serumgo.ProgramID = solana.MustPublicKeyFromBase58(OpenBookAddress)
	closeIx := serumgo.NewCancelOrderByClientIdV2Instruction(
		clientOrderId,
		m.Address,
		m.Metadata.Bids,
		m.Metadata.Asks,
		openOrdersAccount,
		wallet.PublicKey(),
		m.Metadata.EventQueue,
	).Build()
	instructions = append(instructions, closeIx)
	settleIx := serumgo.NewSettleFundsInstruction(
		m.Address,
		openOrdersAccount,
		wallet.PublicKey(),
		m.Metadata.BaseVault,
		m.Metadata.QuoteVault,
		tokenBaseAddress,
		tokenQuoteAddress,
		m.VaultSigner,
		solana.TokenProgramID,
		solana.PublicKey{},
	).Build()
	b, _ := settleIx.Data()
	settle := solana.NewInstruction(
		solana.MustPublicKeyFromBase58(OpenBookAddress),
		settleIx.Accounts()[:9],
		b,
	)
	instructions = append(instructions, settle)
	instructions = append(instructions, jupIn...)
	recent, err := m.Client.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	tx, err := solana.NewTransaction(
		instructions,
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
	log.Println("closing, settling, swaping", sig, err)
	return sig, err
}
