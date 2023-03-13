package gopenbook

import (
	"context"
	"errors"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/serum"
	"github.com/gagliardetto/solana-go/rpc"
)

// get openbook address for openbook programid
func GetOpenBookAddresses(client rpc.Client, owner solana.PublicKey) ([]solana.PublicKey, error) {
	var results = []solana.PublicKey{}
	filters := []rpc.RPCFilter{
		{ // Memcmp: &memcmp,
			DataSize: uint64(3228),
		}, {
			Memcmp: &rpc.RPCFilterMemcmp{
				Offset: 45,
				Bytes:  owner.Bytes(),
			},
		},
	}
	opts := rpc.GetProgramAccountsOpts{
		Commitment: "",
		Encoding:   "base64",
		Filters:    filters,
	}
	out, err := client.GetProgramAccountsWithOpts(context.TODO(), solana.MustPublicKeyFromBase58(OpenBookAddress), &opts)
	if err != nil {
		return results, err
	}

	for _, programAccount := range out {
		results = append(results, programAccount.Pubkey)
	}
	if len(results) == 0 {
		return results, errors.New("not found, create!")
	}
	return results, nil
}

// from publicKey to data in openorderaccount
func GetOpenOrderDataForAccount(client *rpc.Client, openOrderAccount solana.PublicKey) serum.OpenOrders {
	acc, _ := client.GetAccountInfo(context.TODO(), openOrderAccount)
	var openBookData serum.OpenOrders
	decoder := bin.NewBinDecoder(acc.GetBinary())
	decoder.Decode(&openBookData)
	return openBookData
}

// returns map of publickeys with openorderaccounts data
func GetOpenOrdersDataForAccounts(client *rpc.Client, openOrderAccounts []solana.PublicKey) map[solana.PublicKey]serum.OpenOrders {
	result := make(map[solana.PublicKey]serum.OpenOrders)
	out, _ := client.GetMultipleAccounts(context.TODO(), openOrderAccounts...)
	for idx, acc := range out.Value {
		var openBookData serum.OpenOrders
		decoder := bin.NewBinDecoder(acc.Data.GetBinary())
		decoder.Decode(&openBookData)
		result[openOrderAccounts[idx]] = openBookData
	}
	return result
}
