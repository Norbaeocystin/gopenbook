package jup

import (
	"encoding/json"
	"github.com/gagliardetto/solana-go"
	"math/big"
	"net/url"
)

var JUP_API = "https://quote-api.jup.ag/v4/"

func GetQuote(inputMint, outputMint, user solana.PublicKey, amount *big.Int) (QuoteGenerated, error) {
	urlParsed, _ := url.Parse(JUP_API + "quote")
	values := urlParsed.Query()
	values.Add("inputMint", inputMint.String())
	values.Add("outputMint", outputMint.String())
	values.Add("amount", amount.String())
	values.Add("swapMode", "ExactIn") // ExactIn or ExactOut
	values.Add("slippageBps", "1")
	values.Add("onlyDirectRoutes", "true")
	values.Add("asLegacyTransaction", "true")
	values.Add("userPublicKey", user.String())
	urlParsed.RawQuery = values.Encode()
	b, err := Get(urlParsed.String())
	var quote QuoteGenerated
	json.Unmarshal(b, &quote)
	return quote, err
}
