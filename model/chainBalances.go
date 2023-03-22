package model

type Balances struct {
	Balances []SubAccEntry `json:"balances"`
}

type SubAccEntry struct {
	SubaccountID string `json:"subaccount_id"`
	Denom        string `json:"denom"`
	Deposits     struct {
		AvailableBalance string `json:"available_balance"`
		TotalBalance     string `json:"total_balance"`
	} `json:"deposits"`
}

type BankBalances struct {
	Balances []BankBalance `json:"balances"`
}
type BankBalance struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}
