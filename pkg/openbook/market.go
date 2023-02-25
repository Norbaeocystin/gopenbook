package openbook

import (
	"context"
	"encoding/binary"
	"errors"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/serum"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"log"
	"math"
	"math/big"
)

type PriceAndQuantity struct {
	Price    *big.Float
	PriceRaw *big.Int
	Quantity *big.Float
}

type Market struct {
	Address       solana.PublicKey
	Metadata      serum.MarketV2
	BaseDecimals  *big.Float
	QuoteDecimals *big.Float
	Client        *rpc.Client
	VaultSigner   solana.PublicKey
}

func NewMarket(address solana.PublicKey, client *rpc.Client) (Market, error) {
	var market Market
	market.Address = address
	market.Client = client
	err := market.FetchMarket()
	if err != nil {
		return market, err
	}
	err = market.Decimals()
	if err != nil {
		return market, err
	}
	buffer := make([]byte, 8)
	binary.LittleEndian.PutUint64(buffer, uint64(market.Metadata.VaultSignerNonce))
	key, _ := solana.CreateProgramAddress([][]byte{market.Address.Bytes(), buffer}, solana.MustPublicKeyFromBase58(OpenBookAddress))
	market.VaultSigner = key
	return market, nil
}

func (m *Market) FetchMarket() error {
	market, err := serum.FetchMarket(context.TODO(), m.Client, m.Address)
	m.Metadata = market.MarketV2
	return err
}

func (m *Market) Decimals() error {
	// best way to use token.Mint - for fetching decimals data
	accountBase, _ := m.Client.GetAccountInfo(context.TODO(), m.Metadata.BaseMint)
	var baseData token.Mint
	decoder := bin.NewBinDecoder(accountBase.GetBinary())
	decoder.Decode(&baseData)
	// log.Println(baseData)
	m.BaseDecimals = big.NewFloat(math.Pow(10, float64(baseData.Decimals)))
	accountQuote, _ := m.Client.GetAccountInfo(context.TODO(), m.Metadata.QuoteMint)
	var quoteData token.Mint
	decoder = bin.NewBinDecoder(accountQuote.GetBinary())
	decoder.Decode(&quoteData)
	m.QuoteDecimals = big.NewFloat(math.Pow(10, float64(quoteData.Decimals)))
	return nil
}

func (m Market) LoadBids() {
	info, _ := m.Client.GetAccountInfo(context.TODO(), m.Metadata.Bids)
	var orderbook serum.Orderbook
	decoder := bin.NewBinDecoder(info.GetBinary())
	decoder.Decode(&orderbook)
	orderbook.Items(false, func(node *serum.SlabLeafNode) error {
		priceBI := new(big.Int).Rsh(node.Key.BigInt(), 64)
		priceRaw := new(big.Float).SetInt(priceBI)
		price := new(big.Float).Mul(new(big.Float).Quo(priceRaw, m.BaseDecimals), m.QuoteDecimals)
		quantity := new(big.Float).Mul(new(big.Float).Quo(big.NewFloat(float64(node.Quantity)), m.BaseDecimals), m.QuoteDecimals)
		log.Println(price, quantity)
		return nil
	})
}

func (m Market) LoadBidsForOwner(openOrderAccount solana.PublicKey) []*serum.SlabLeafNode {
	info, _ := m.Client.GetAccountInfoWithOpts(context.TODO(), m.Metadata.Bids, &rpc.GetAccountInfoOpts{
		Commitment: rpc.CommitmentProcessed,
		Encoding:   "",
		DataSlice:  nil,
	})
	var orderbook serum.Orderbook
	decoder := bin.NewBinDecoder(info.GetBinary())
	decoder.Decode(&orderbook)
	nodes := []*serum.SlabLeafNode{}
	orderbook.Items(true, func(node *serum.SlabLeafNode) error {
		if node.Owner.String() == openOrderAccount.String() {
			nodes = append(nodes, node)
		}
		return nil
	})
	return nodes
}

func (m Market) LoadAsks() {
	info, _ := m.Client.GetAccountInfo(context.TODO(), m.Metadata.Asks)
	var orderbook serum.Orderbook
	decoder := bin.NewBinDecoder(info.GetBinary())
	decoder.Decode(&orderbook)
	orderbook.Items(false, func(node *serum.SlabLeafNode) error {
		priceBI := new(big.Int).Rsh(node.Key.BigInt(), 64)
		priceRaw := new(big.Float).SetInt(priceBI)
		price := new(big.Float).Mul(new(big.Float).Quo(priceRaw, m.BaseDecimals), m.QuoteDecimals)
		quantity := new(big.Float).Mul(new(big.Float).Quo(big.NewFloat(float64(node.Quantity)), m.BaseDecimals), m.QuoteDecimals)
		log.Println(price, quantity)
		return nil
	})
}

func (m Market) LoadAsksForOwner(openOrderAccount solana.PublicKey) []*serum.SlabLeafNode {
	info, _ := m.Client.GetAccountInfoWithOpts(context.TODO(), m.Metadata.Asks, &rpc.GetAccountInfoOpts{
		Commitment: rpc.CommitmentProcessed,
		Encoding:   "",
		DataSlice:  nil,
	})
	var orderbook serum.Orderbook
	decoder := bin.NewBinDecoder(info.GetBinary())
	decoder.Decode(&orderbook)
	nodes := []*serum.SlabLeafNode{}
	orderbook.Items(false, func(node *serum.SlabLeafNode) error {
		if node.Owner.String() == openOrderAccount.String() {
			nodes = append(nodes, node)
		}
		return nil
	})
	return nodes
}

func (m Market) LoadNBids(n int) []PriceAndQuantity {
	info, _ := m.Client.GetAccountInfo(context.TODO(), m.Metadata.Bids)
	var orderbook serum.Orderbook
	decoder := bin.NewBinDecoder(info.GetBinary())
	decoder.Decode(&orderbook)
	prices := make([]PriceAndQuantity, 0)
	count := 0
	orderbook.Items(true, func(node *serum.SlabLeafNode) error {
		priceBI := new(big.Int).Rsh(node.Key.BigInt(), 64)
		priceRaw := new(big.Float).SetInt(priceBI)
		price := new(big.Float).Mul(new(big.Float).Quo(priceRaw, m.BaseDecimals), m.QuoteDecimals)
		quantity := new(big.Float).Mul(new(big.Float).Quo(big.NewFloat(float64(node.Quantity)), m.BaseDecimals), m.QuoteDecimals)
		// log.Println(price, quantity)
		prices = append(prices, PriceAndQuantity{price, priceBI, quantity})
		count += 1
		if count == n {
			return errors.New("not real error heh")
		}
		return nil
	})
	return prices
}

func (m Market) LoadBidSumGTE(sum *big.Float) PriceAndQuantity {
	info, _ := m.Client.GetAccountInfo(context.TODO(), m.Metadata.Bids)
	var orderbook serum.Orderbook
	decoder := bin.NewBinDecoder(info.GetBinary())
	decoder.Decode(&orderbook)
	sumOrder := big.NewFloat(0)
	var priceAndQuantity PriceAndQuantity
	orderbook.Items(true, func(node *serum.SlabLeafNode) error {
		priceBI := new(big.Int).Rsh(node.Key.BigInt(), 64)
		priceRaw := new(big.Float).SetInt(priceBI)
		price := new(big.Float).Mul(new(big.Float).Quo(priceRaw, m.BaseDecimals), m.QuoteDecimals)
		quantity := new(big.Float).Mul(new(big.Float).Quo(big.NewFloat(float64(node.Quantity)), m.BaseDecimals), m.QuoteDecimals)
		// log.Println(price, quantity)
		sumOrder = new(big.Float).Add(quantity, sumOrder)
		if sumOrder.Cmp(sum) == 0 || sumOrder.Cmp(sum) == 1 {
			priceAndQuantity = PriceAndQuantity{price, priceBI, quantity}
			return errors.New("not real error heh")
		}
		return nil
	})
	return priceAndQuantity
}

func (m Market) LoadNAsks(n int) []PriceAndQuantity {
	info, _ := m.Client.GetAccountInfo(context.TODO(), m.Metadata.Asks)
	var orderbook serum.Orderbook
	decoder := bin.NewBinDecoder(info.GetBinary())
	decoder.Decode(&orderbook)
	prices := make([]PriceAndQuantity, 0)
	count := 0
	orderbook.Items(false, func(node *serum.SlabLeafNode) error {
		priceBI := new(big.Int).Rsh(node.Key.BigInt(), 64)
		priceRaw := new(big.Float).SetInt(priceBI)
		price := new(big.Float).Mul(new(big.Float).Quo(priceRaw, m.BaseDecimals), m.QuoteDecimals)
		quantity := new(big.Float).Mul(new(big.Float).Quo(big.NewFloat(float64(node.Quantity)), m.BaseDecimals), m.QuoteDecimals)
		// log.Println(price, quantity)
		prices = append(prices, PriceAndQuantity{price, priceBI, quantity})
		count += 1
		if count == n {
			return errors.New("not real error heh")
		}
		return nil
	})
	return prices
}

func (m Market) LoadAskSumGTE(sum *big.Float) PriceAndQuantity {
	info, _ := m.Client.GetAccountInfo(context.TODO(), m.Metadata.Asks)
	var orderbook serum.Orderbook
	decoder := bin.NewBinDecoder(info.GetBinary())
	decoder.Decode(&orderbook)
	sumOrder := big.NewFloat(0)
	var priceAndQuantity PriceAndQuantity
	orderbook.Items(false, func(node *serum.SlabLeafNode) error {
		priceBI := new(big.Int).Rsh(node.Key.BigInt(), 64)
		priceRaw := new(big.Float).SetInt(priceBI)
		price := new(big.Float).Mul(new(big.Float).Quo(priceRaw, m.BaseDecimals), m.QuoteDecimals)
		quantity := new(big.Float).Mul(new(big.Float).Quo(big.NewFloat(float64(node.Quantity)), m.BaseDecimals), m.QuoteDecimals)
		// log.Println(price, quantity)
		sumOrder = new(big.Float).Add(quantity, sumOrder)
		if sumOrder.Cmp(sum) == 0 || sumOrder.Cmp(sum) == 1 {
			priceAndQuantity = PriceAndQuantity{price, priceBI, quantity}
			return errors.New("not real error heh")
		}
		return nil
	})
	return priceAndQuantity

}

func (m Market) GetOpenOrdersDataForAccount(ooa solana.PublicKey) serum.OpenOrders {
	acc, _ := m.Client.GetAccountInfo(context.TODO(), ooa)
	var openBookData serum.OpenOrders
	decoder := bin.NewBinDecoder(acc.GetBinary())
	decoder.Decode(&openBookData)
	return openBookData
}

// need to be created if does not exists to be able to open orders
// via systemrprogram create account ownership to openbook
func (m Market) GetOpenOrdersData(owner solana.PublicKey) (serum.OpenOrders, error) {

	filters := []rpc.RPCFilter{
		{ // Memcmp: &memcmp,
			DataSize: uint64(3228),
		}, {
			Memcmp: &rpc.RPCFilterMemcmp{
				Offset: 45,
				Bytes:  owner.Bytes(),
			},
		},
		{
			Memcmp: &rpc.RPCFilterMemcmp{
				Offset: 13,
				Bytes:  m.Address.Bytes(),
			},
		},
	}
	opts := rpc.GetProgramAccountsOpts{
		Commitment: "",
		Encoding:   "base64",
		Filters:    filters,
	}
	out, err := m.Client.GetProgramAccountsWithOpts(context.TODO(), solana.MustPublicKeyFromBase58(OpenBookAddress), &opts)
	var openBookData serum.OpenOrders
	if err != nil {
		return openBookData, err
	}
	for _, programAccount := range out {
		b := programAccount.Account.Data.GetBinary()
		var openBookData serum.OpenOrders
		decoder := bin.NewBinDecoder(b)
		decoder.Decode(&openBookData)
		return openBookData, nil
	}
	return openBookData, errors.New("not found, create!")
}

// need to be created if does not exists to be able to open orders
// via systemrprogram create account ownership to openbook
func (m Market) GetOpenOrdersAccount(owner solana.PublicKey) ([]solana.PublicKey, error) {
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
		{
			Memcmp: &rpc.RPCFilterMemcmp{
				Offset: 13,
				Bytes:  m.Address.Bytes(),
			},
		},
	}
	opts := rpc.GetProgramAccountsOpts{
		Commitment: "",
		Encoding:   "base64",
		Filters:    filters,
	}
	out, err := m.Client.GetProgramAccountsWithOpts(context.TODO(), solana.MustPublicKeyFromBase58(OpenBookAddress), &opts)
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
