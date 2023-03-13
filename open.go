package gopenbook

import (
	"context"
	"fmt"
	serumgo "github.com/gagliardetto/serum-go"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"log"
	"time"
)

// open - sell and buy position in one tx
func Open(sellPrice, buyPrice, amount uint64, market Market, openOrders, tokenOwnerBuyAddress, tokenOwnerSellAddress, feeDiscount solana.PublicKey, wallet solana.PrivateKey) (solana.Signature, error) {
	serumgo.SetProgramID(solana.MustPublicKeyFromBase58(OpenBookAddress))
	matchIx := serumgo.NewMatchOrdersInstruction(5,
		market.Address,
		market.Metadata.RequestQueue,
		market.Metadata.EventQueue,
		market.Metadata.Bids,
		market.Metadata.Asks,
	).Build()
	buyNoiV3 := serumgo.NewOrderInstructionV3{
		Side:                        serumgo.SideBid,
		LimitPrice:                  buyPrice,
		MaxCoinQty:                  amount,
		MaxNativePcQtyIncludingFees: buyPrice * amount,
		SelfTradeBehavior:           0,
		OrderType:                   serumgo.OrderTypeLimit,
		ClientOrderId:               uint64(time.Now().Unix()),
		Limit:                       65535,
	}
	iBuy := serumgo.NewNewOrderV3Instruction(
		buyNoiV3,
		market.Address,
		openOrders,
		market.Metadata.RequestQueue,
		market.Metadata.EventQueue,
		market.Metadata.Bids,
		market.Metadata.Asks,
		tokenOwnerBuyAddress,
		wallet.PublicKey(),
		market.Metadata.BaseVault,
		market.Metadata.QuoteVault,
		solana.TokenProgramID,
		solana.SysVarRentPubkey,
		feeDiscount,
	).Build()
	log.Println(buyPrice, amount, buyPrice*amount)
	log.Println(sellPrice, amount, sellPrice*amount)
	sellNoiV3 := serumgo.NewOrderInstructionV3{
		Side:                        serumgo.SideAsk,
		LimitPrice:                  sellPrice,
		MaxCoinQty:                  amount,
		MaxNativePcQtyIncludingFees: sellPrice * amount,
		SelfTradeBehavior:           0,
		OrderType:                   serumgo.OrderTypeLimit,
		ClientOrderId:               uint64(time.Now().Unix()),
		Limit:                       65535,
	}
	iSell := serumgo.NewNewOrderV3Instruction(
		sellNoiV3,
		market.Address,
		openOrders,
		market.Metadata.RequestQueue,
		market.Metadata.EventQueue,
		market.Metadata.Bids,
		market.Metadata.Asks,
		tokenOwnerSellAddress,
		wallet.PublicKey(),
		market.Metadata.BaseVault,
		market.Metadata.QuoteVault,
		solana.TokenProgramID,
		solana.SysVarRentPubkey,
		feeDiscount,
	).Build()
	//dataSell, _ := iSell.Data()
	//sellI := solana.NewInstruction(
	//	iSell.ProgramID(),
	//	iSell.Accounts()[:12],
	//	dataSell,
	//)
	//dataBuy, _ := iBuy.Data()
	//buyI := solana.NewInstruction(iBuy.ProgramID(),
	//	iBuy.Accounts()[:12],
	//	dataBuy,
	//)
	recent, err := market.Client.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	tx, err := solana.NewTransaction(
		[]solana.Instruction{matchIx, iSell,
			iBuy, matchIx,
		},
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
	sig, err := market.Client.SendTransactionWithOpts(context.TODO(), tx,
		rpc.TransactionOpts{
			Encoding:            "",
			SkipPreflight:       false,
			PreflightCommitment: rpc.CommitmentProcessed,
			MaxRetries:          nil,
			MinContextSlot:      nil,
		},
	)
	log.Println("bidandasktx", sig, err)
	return sig, err
}

// if openbookaccount address does not exists creates tx
func OpenOpenBookAccount(client *rpc.Client, wallet solana.PrivateKey) (solana.Signature, error, solana.PublicKey) {
	nw := solana.NewWallet()
	i := system.NewCreateAccountInstruction(
		26357760,
		3328,
		solana.MustPublicKeyFromBase58(OpenBookAddress),
		wallet.PublicKey(),
		nw.PublicKey(),
	).Build()
	recent, err := client.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
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
			if nw.PublicKey().Equals(key) {
				return &nw.PrivateKey
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
	sig, err := client.SendTransactionWithOpts(context.TODO(), tx,
		rpc.TransactionOpts{
			Encoding:            "",
			SkipPreflight:       false,
			PreflightCommitment: rpc.CommitmentProcessed,
			MaxRetries:          nil,
			MinContextSlot:      nil,
		},
	)
	log.Println("opening", sig, err)
	return sig, err, wallet.PublicKey()
}
