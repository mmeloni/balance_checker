package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/xlab/suplog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"os"
	"portfolio-bank-balance-checks/model"

	"github.com/cosmos/cosmos-sdk/types"

	indexerModels "github.com/InjectiveLabs/injective-indexer/db/model"
	cosmtypes "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

func GetComparableFromMongodb(db *mongo.Database, address string) (comparable *model.ComparablePortfolio, err error) {
	var portfolio []model.Portfolio
	cur, err := db.Collection("portfolio").Find(context.Background(), bson.M{"accountAddress": address})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.NewComparablePortfolio(), nil
		}
		panic(err)
	}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		var p model.Portfolio
		err := cur.Decode(&p)
		if err != nil {
			panic(err)
		}
		portfolio = append(portfolio, p)
	}
	comparable = model.NewComparablePortfolio()
	for _, p := range portfolio {
		switch p.Type {
		case 1:
			if _, ok := comparable.AvailableBalances[p.SubaccountId]; !ok {
				comparable.AvailableBalances[p.SubaccountId] = make(map[string]string)
			}
			comparable.AvailableBalances[p.SubaccountId][p.Denom] = PrimitiveDecimal128ToString(p.Amount)
		case 2:
			if _, ok := comparable.TotalBalances[p.SubaccountId]; !ok {
				comparable.TotalBalances[p.SubaccountId] = make(map[string]string)
			}
			comparable.TotalBalances[p.SubaccountId][p.Denom] = PrimitiveDecimal128ToString(p.Amount)
		default:
			comparable.BankBalances[p.Denom] = PrimitiveDecimal128ToString(p.Amount)
		}
	}
	comparable.AccountAddress = address

	return comparable, nil
}

func GetComparableFromExchangeAPI(address string) (comparable *model.ComparablePortfolio, err error) {

	portfolioAPIResp, err := getPortfolioAPI(address)
	if err != nil {
		return comparable, err
	}

	comparable = model.NewComparablePortfolio()

	for _, p := range portfolioAPIResp.Portfolio.BankBalances {
		bankBalanceDecimal, err := types.NewDecFromStr(p.Amount)
		if err != nil {
			return comparable, err
		}
		comparable.BankBalances[p.Denom] = bankBalanceDecimal.String()
	}
	for _, subaccount := range portfolioAPIResp.Portfolio.Subaccounts {
		if _, ok := comparable.AvailableBalances[subaccount.Subaccount]; !ok {
			comparable.AvailableBalances[subaccount.Subaccount] = make(map[string]string)
		}
		if _, ok := comparable.TotalBalances[subaccount.Subaccount]; !ok {
			comparable.TotalBalances[subaccount.Subaccount] = make(map[string]string)
		}
		availableBalanceDecimal, err := types.NewDecFromStr(subaccount.Deposit.AvailableBalance)
		if err != nil {
			return comparable, err
		}
		totalBalanceDecimal, err := types.NewDecFromStr(subaccount.Deposit.TotalBalance)
		if err != nil {
			return comparable, err
		}
		comparable.AvailableBalances[subaccount.Subaccount][subaccount.Denom] = availableBalanceDecimal.String()
		comparable.TotalBalances[subaccount.Subaccount][subaccount.Denom] = totalBalanceDecimal.String()
	}
	comparable.AccountAddress = address

	return comparable, nil
}

func getPortfolioAPI(address string) (portApiResp model.PortfolioApiResponse, err error) {
	//23.88.5.151
	exchangeAddress, ok := os.LookupEnv("INDEXER_EXCHANGE_URL")
	if !ok {
		return portApiResp, fmt.Errorf("INDEXER_EXCHANGE_URL env is not set")
	}

	suplog.Infoln("exchangeAddress: ", exchangeAddress)
	//api/exchange/portfolio/v2/portfolio/inj1cml96vmptgw99syqrrz8az79xer2pcgp0a885r
	resp, err := http.Get(fmt.Sprintf("%s/api/exchange/portfolio/v2/portfolio/%s", exchangeAddress, address))
	if err != nil {
		return portApiResp, err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return portApiResp, nil
		}
		PrintBody(resp)
		panic(fmt.Errorf("status code is not 200, it is %d", resp.StatusCode))
	}

	err = json.NewDecoder(resp.Body).Decode(&portApiResp)
	if err != nil {
		return portApiResp, err
	}
	return portApiResp, nil
}

func GetComparableFromChain(address string) (comparable *model.ComparablePortfolio, err error) {
	//##-> fetch bank balances
	chainAddress, ok := os.LookupEnv("INDEXER_COSMOS_LCD_URL")
	if !ok {
		return nil, fmt.Errorf("INDEXER_COSMOS_LCD_URL env is not set")
	}

	block, ok := os.LookupEnv("BLOCK_NUMBER")
	if !ok {
		return nil, fmt.Errorf("BLOCK_NUMBER env is not set")
	}
	client := http.Client{}
	// curl -X GET "/bank/v1beta1/balances/inj1qqqvuz86yfrfd2qesac8p0eh3693xxk0h83mqe" -H "accept: application/json"
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", chainAddress, address), nil)
	if err != nil {
		return nil, err
	}

	req.Header = http.Header{
		"x-cosmos-block-height": {block},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var balances model.BankBalances
	err = json.NewDecoder(resp.Body).Decode(&balances)
	if err != nil {
		return nil, err
	}
	comparable, ok = SubaccountBalances[address]
	if !ok {
		comparable = model.NewComparablePortfolio()
	}
	for _, b := range balances.Balances {
		intAmount, ok := cosmtypes.NewIntFromString(b.Amount)
		if !ok {
			return nil, fmt.Errorf("failed to parse int from string")
		}
		amount := cosmtypes.NewDecFromInt(intAmount)
		if err != nil {
			return nil, err
		}
		comparable.BankBalances[b.Denom] = amount.String()
	}
	comparable.AccountAddress = address

	return comparable, nil
}

func initChainTotalAndAvailableBalances() (comparables map[string]*model.ComparablePortfolio, err error) {
	chainAddress, ok := os.LookupEnv("INDEXER_COSMOS_LCD_URL")
	if !ok {
		return nil, fmt.Errorf("INDEXER_COSMOS_LCD_URL env is not set")
	}

	block, ok := os.LookupEnv("BLOCK_NUMBER")
	if !ok {
		return nil, fmt.Errorf("BLOCK_NUMBER env is not set")
	}
	client := http.Client{}
	//api/exchange/portfolio/v2/portfolio/inj1cml96vmptgw99syqrrz8az79xer2pcgp0a885r
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/injective/exchange/v1beta1/exchange/exchangeBalances", chainAddress), nil)
	if err != nil {
		return nil, err
	}

	req.Header = http.Header{
		"x-cosmos-block-height": {block},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var balances model.Balances
	err = json.NewDecoder(resp.Body).Decode(&balances)
	if err != nil {
		return nil, err
	}

	collNumber := int64(len(balances.Balances))
	if collNumber == 0 {
		suplog.Errorf("No subaccounts found")
		return nil, fmt.Errorf("no subaccounts found")
	}

	comparables = make(map[string]*model.ComparablePortfolio, collNumber)
	i := 0
	for _, subAccEntry := range balances.Balances {
		comp := model.NewComparablePortfolio()

		i++
		cosmtypes.GetConfig().SetBech32PrefixForAccount("inj", "injpub")

		hash := ethcommon.HexToHash(subAccEntry.SubaccountID)
		if len(hash.Bytes()) < 20 {
			return nil, fmt.Errorf("invalid subaccount id: %s", subAccEntry.SubaccountID)
		}
		slice := hash.Bytes()[:20]

		addr := cosmtypes.AccAddress(slice)
		accAddress := addr.String()

		availableBalancesDec, err := types.NewDecFromStr(subAccEntry.Deposits.AvailableBalance)
		if err != nil {
			return nil, fmt.Errorf("error while converting available balance to decimal: %s", err)
		}
		availableBalancesBigNum := indexerModels.NewBigNumFromDec(availableBalancesDec)
		availableBalancesDec128 := primitive.Decimal128(availableBalancesBigNum)

		portfolioAvailableBalance := model.Portfolio{
			AccountAddress: accAddress,
			SubaccountId:   subAccEntry.SubaccountID,
			Denom:          subAccEntry.Denom,
			Amount:         availableBalancesDec128,
			Type:           1,
		}

		totalBalancesDec, err := types.NewDecFromStr(subAccEntry.Deposits.TotalBalance)
		if err != nil {
			return nil, fmt.Errorf("error while converting total balance to decimal: %s", err)
		}
		totalBalancesBigNum := indexerModels.NewBigNumFromDec(totalBalancesDec)
		totalBalancesDec128 := primitive.Decimal128(totalBalancesBigNum)
		portfolioTotalBalance := model.Portfolio{
			AccountAddress: accAddress,
			SubaccountId:   subAccEntry.SubaccountID,
			Denom:          subAccEntry.Denom,
			Amount:         totalBalancesDec128,
			Type:           int8(2),
		}

		if _, ok := comp.AvailableBalances[portfolioAvailableBalance.SubaccountId]; !ok {
			comp.AvailableBalances[portfolioAvailableBalance.SubaccountId] = make(map[string]string)
		}
		if _, ok := comp.TotalBalances[portfolioTotalBalance.SubaccountId]; !ok {
			comp.TotalBalances[portfolioTotalBalance.SubaccountId] = make(map[string]string)
		}
		comp.AvailableBalances[portfolioAvailableBalance.SubaccountId][portfolioAvailableBalance.Denom] = portfolioAvailableBalance.Amount.String()
		comp.TotalBalances[portfolioTotalBalance.SubaccountId][portfolioTotalBalance.Denom] = portfolioTotalBalance.Amount.String()

		comp.AccountAddress = accAddress

		comparables[accAddress] = comp
	}

	suplog.Infof("Total comparables fetched from chain: %d\n", len(comparables))
	return comparables, nil
}

func DecToString(dec types.Dec) string {
	if dec.IsZero() {
		return "0"
	}
	return dec.String()
}
func PrimitiveDecimal128ToString(dec primitive.Decimal128) string {
	if dec.IsZero() {
		return "0"
	}
	return dec.String()
}
