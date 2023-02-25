package jup

import (
	"encoding/base64"
	"encoding/json"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"net/url"
)

func GetTX(user solana.PublicKey, route Route) (*solana.Transaction, error) {
	var tx *solana.Transaction
	urlParsed, _ := url.Parse(JUP_API + "swap")
	routeTx := RouteTx{route, user.String(), false, true}
	body, _ := json.Marshal(routeTx)
	b, e := Post(urlParsed.String(), body)
	if e != nil {
		return tx, e
	}
	var swapTx TxData
	json.Unmarshal(b, &swapTx)
	out, err := base64.StdEncoding.DecodeString(swapTx.SwapTransaction)
	if err != nil {
		return tx, err
	}
	err = bin.NewBinDecoder(out).Decode(&tx)
	return tx, err
}
