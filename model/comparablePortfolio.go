package model

type ComparablePortfolio struct {
	AccountAddress string
	// denom -> amount
	BankBalances map[string]string
	// subaccount -> denom -> amount
	AvailableBalances map[string]map[string]string
	// subaccount -> denom -> amount
	TotalBalances map[string]map[string]string
}

func NewComparablePortfolio() *ComparablePortfolio {
	return &ComparablePortfolio{
		BankBalances:      make(map[string]string),
		AvailableBalances: make(map[string]map[string]string),
		TotalBalances:     make(map[string]map[string]string),
	}
}
