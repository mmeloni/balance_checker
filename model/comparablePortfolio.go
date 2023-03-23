package model

import "github.com/cosmos/cosmos-sdk/types"

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

func (c *ComparablePortfolio) IsEqual(other *ComparablePortfolio) bool {
	if c.AccountAddress != other.AccountAddress {
		return false
	}
	if !c.isBankBalancesEqual(other) {
		return false
	}
	if !c.isAvailableBalancesEqual(other) {
		return false
	}
	if !c.isTotalBalancesEqual(other) {
		return false
	}
	return true
}

func (c *ComparablePortfolio) isBankBalancesEqual(other *ComparablePortfolio) bool {
	if len(c.BankBalances) != len(other.BankBalances) {
		println("bank balances length not equal")
	}
	for denom, amount := range c.BankBalances {
		if otherAmount, ok := other.BankBalances[denom]; !ok || !compareAmounts(amount, otherAmount) {
			if !ok {
				isZeroAmount(amount)
				continue
			}
			return false
		}
	}
	return true
}

func (c *ComparablePortfolio) isAvailableBalancesEqual(other *ComparablePortfolio) bool {
	if len(c.AvailableBalances) != len(other.AvailableBalances) {
		println("available balances length not equal")
	}
	for subaccount, balances := range c.AvailableBalances {
		if otherBalances, ok := other.AvailableBalances[subaccount]; !ok {
			return false
		} else {
			if len(balances) != len(otherBalances) {
				println("available balances length not equal")
			}
			for denom, amount := range balances {
				if otherAmount, ok := otherBalances[denom]; !ok || !compareAmounts(amount, otherAmount) {
					if !ok {
						isZeroAmount(amount)
						continue
					}
					return false
				}
			}
		}
	}
	return true
}

func (c *ComparablePortfolio) isTotalBalancesEqual(other *ComparablePortfolio) bool {
	if len(c.TotalBalances) != len(other.TotalBalances) {
		println("total balances length not equal")
	}
	for subaccount, balances := range c.TotalBalances {
		if otherBalances, ok := other.TotalBalances[subaccount]; !ok {
			return false
		} else {
			if len(balances) != len(otherBalances) {
				println("total balances length not equal")
			}
			for denom, amount := range balances {
				if otherAmount, ok := otherBalances[denom]; !ok || !compareAmounts(amount, otherAmount) {
					if !ok {
						isZeroAmount(amount)
						continue
					}
					return false
				}
			}
		}
	}
	return true
}

func compareAmounts(a, b string) bool {
	amountA, err := types.NewDecFromStr(a)
	if err != nil {
		return false
	}
	amountB, err := types.NewDecFromStr(b)
	if err != nil {
		return false
	}
	return amountA.Equal(amountB)
}

func isZeroAmount(amount string) bool {
	amountDec, err := types.NewDecFromStr(amount)
	if err != nil {
		return false
	}
	return amountDec.IsZero()
}
