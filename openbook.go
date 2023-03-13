package gopenbook

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"log"
)

const RPCSOLANA = "https://api.mainnet-beta.solana.com"

var Address2Name = map[string]string{
	"9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin": "Serum Dex Program v3",
	"11111111111111111111111111111111":             "System Program",
	"5fNfvyp5czQVX77yoACa3JJVEhdRaWjPuazuWgjhTqEH": "Mango Markets V2",
	"So1endDq2YkqhipRh3WViPa8hdiSpxWy6z3Z6tMCpAo":  "Solend Program",
	"perpke6JybKfRDitCmnazpCrGN5JRApxxukhA9Js6E6":  "Bonfida",
}

func GetAllSignatures() []*rpc.TransactionSignature {
	client := rpc.New(RPCSOLANA)
	signatures := GetLastSignatures(client, OpenBookAddress)
	log.Println("Got signatures", len(signatures))
	if len(signatures) > 0 {
		for {
			lastSignature := signatures[len(signatures)-1]
			nextSignatures := GetSignaturesBefore(client, OpenBookAddress, lastSignature.Signature)
			if len(nextSignatures) > 0 {
				signatures = append(signatures, nextSignatures...)
				log.Println("signatures", len(signatures))
			} else {
				break
			}
		}
	}
	return signatures
}

func GetLastSignatures(client *rpc.Client, address string) []*rpc.TransactionSignature {
	key, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		log.Fatal(err)
	}
	signatures, _ := client.GetSignaturesForAddress(context.TODO(), key)
	return signatures
}

func GetSignaturesBefore(client *rpc.Client, address string, signature solana.Signature) []*rpc.TransactionSignature {
	opts := rpc.GetSignaturesForAddressOpts{
		Limit:  nil,
		Before: signature,
	}
	key, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		log.Fatal(err)
	}
	signatures, _ := client.GetSignaturesForAddressWithOpts(context.TODO(),
		key,
		&opts,
	)
	return signatures
}

func GetTransactionFromSignature(signature solana.Signature) (*rpc.GetTransactionResult, error) {
	endpoint := RPCSOLANA // rpc.MainNetBetaSerum_RPC
	client := rpc.New(endpoint)
	opts := rpc.GetTransactionOpts{
		Encoding:   solana.EncodingBase58,
		Commitment: "",
	}
	txs, err := client.GetTransaction(context.TODO(), signature, &opts)
	return txs, err
}
