package jup

type Route struct {
	InAmount       string `json:"inAmount"`
	OutAmount      string `json:"outAmount"`
	PriceImpactPct int    `json:"priceImpactPct"`
	MarketInfos    []struct {
		ID                 string `json:"id"`
		Label              string `json:"label"`
		InputMint          string `json:"inputMint"`
		OutputMint         string `json:"outputMint"`
		NotEnoughLiquidity bool   `json:"notEnoughLiquidity"`
		InAmount           string `json:"inAmount"`
		OutAmount          string `json:"outAmount"`
		PriceImpactPct     int    `json:"priceImpactPct"`
		LpFee              struct {
			Amount string  `json:"amount"`
			Mint   string  `json:"mint"`
			Pct    float64 `json:"pct"`
		} `json:"lpFee"`
		PlatformFee struct {
			Amount string `json:"amount"`
			Mint   string `json:"mint"`
			Pct    int    `json:"pct"`
		} `json:"platformFee"`
	} `json:"marketInfos"`
	Amount               string `json:"amount"`
	SlippageBps          int    `json:"slippageBps"`
	OtherAmountThreshold string `json:"otherAmountThreshold"`
	SwapMode             string `json:"swapMode"`
	Fees                 struct {
		SignatureFee             int           `json:"signatureFee"`
		OpenOrdersDeposits       []interface{} `json:"openOrdersDeposits"`
		AtaDeposits              []interface{} `json:"ataDeposits"`
		TotalFeeAndDeposits      int           `json:"totalFeeAndDeposits"`
		MinimumSOLForTransaction int           `json:"minimumSOLForTransaction"`
	} `json:"fees"`
}

type QuoteGenerated struct {
	Data        []Route `json:"data"`
	TimeTaken   float64 `json:"timeTaken"`
	ContextSlot int     `json:"contextSlot"`
}

type RouteTx struct {
	Route               Route  `json:"route"`
	UserPublickKey      string `json:"userPublicKey"`
	WrapUnwrapSOL       bool   `json:"wrapUnwrapSOL"`
	AsLegacyTransaction bool   `json:"asLegacyTransaction"`
}

type TxData struct {
	SwapTransaction string `json:"swapTransaction"`
}
