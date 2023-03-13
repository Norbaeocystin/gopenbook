package gopenbook

import (
	"context"
	"fmt"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/rpc"
	"log"
)

// find or create ata
func FindOrCreateATA(client *rpc.Client, wallet solana.PrivateKey, mint solana.PublicKey) solana.PublicKey {
	ataAccount, _, _ := solana.FindAssociatedTokenAddress(wallet.PublicKey(), mint)
	tokens, err := client.GetTokenAccountsByOwner(context.TODO(), wallet.PublicKey(),
		&rpc.GetTokenAccountsConfig{
			Mint: &mint,
			// ProgramId: solana.TokenProgramID.ToPointer(),
		},

		&rpc.GetTokenAccountsOpts{
			Commitment: "",
			Encoding:   solana.EncodingBase64,
			DataSlice:  nil,
		})
	if err != nil {
		log.Println("got error during searching", err)
	}
	for _, token := range tokens.Value {
		// log.Println("found ata", token.Pubkey)
		return token.Pubkey
	}
	log.Println("not found ata, creating ata")
	i := associatedtokenaccount.NewCreateInstruction(
		wallet.PublicKey(),
		wallet.PublicKey(),
		mint,
	).Build()
	recent, err := client.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			i,
		},
		recent.Value.Blockhash, //NONCE
		solana.TransactionPayer(wallet.PublicKey()),
	)
	// log.Println(tx, err)
	// TODO intiliaze those 2 accounts
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
	sig, err := client.SendTransactionWithOpts(context.TODO(), tx,
		rpc.TransactionOpts{
			Encoding:            "",
			SkipPreflight:       false,
			PreflightCommitment: "",
			MaxRetries:          nil,
			MinContextSlot:      nil,
		},
	)
	if err != nil {
		panic(err)
	}
	log.Println("tx for creating ata:", sig)
	// return sig
	return ataAccount
}
