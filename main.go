package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/xlab/suplog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"os"
	"portfolio-bank-balance-checks/model"
)

var SubaccountBalances map[string]*model.ComparablePortfolio

func main() {
	_, ok := os.LookupEnv("INDEXER_COSMOS_LCD_URL")
	if !ok {
		panic(fmt.Errorf("INDEXER_COSMOS_LCD_URL env is not set"))
	}
	_, ok = os.LookupEnv("BLOCK_NUMBER")
	if !ok {
		panic(fmt.Errorf("BLOCK_NUMBER env is not set"))
	}
	_, ok = os.LookupEnv("INDEXER_EXCHANGE_URL")
	if !ok {
		panic(fmt.Errorf("INDEXER_EXCHANGE_URL env is not set"))
	}
	//##-> stop main net stage exchange process and retrieve {latest_block}
	//##-> fetch all accounts
	accounts, err := GetAllAccounts()
	if err != nil {
		panic(err)
	}
	//##-> mongo connect
	db, err := mongoConnect()
	if err != nil {
		panic(err)
	}

	//##-> fetch all subaccount balances(total, available) from chain
	SubaccountBalances, err = initChainTotalAndAvailableBalances()
	if err != nil {
		panic(err)
	}

	for id, account := range accounts {
		//##-> fetch exchange mongodb account balances, bank, total, available on account x
		comp1, err := GetComparableFromMongodb(db, account.BaseAccount.Address)
		if err != nil {
			panic(err)
		}
		//##-> fetch exchange portfolio on account x
		comp2, err := GetComparableFromExchangeAPI(account.BaseAccount.Address)
		if err != nil {
			panic(err)
		}
		//##-> fetch chain bank, total, available on account x using 'x-cosmos-block-height: {latest_block}'
		comp3, err := GetComparableFromChain(account.BaseAccount.Address)

		fmt.Printf("comp1: %+v\n", comp1)
		fmt.Printf("comp2: %+v\n", comp2)
		fmt.Printf("comp3: %+v\n", comp3)

		fmt.Printf("processing account %d/%d: %s\n", id, len(accounts), account.BaseAccount.Address)
	}

	//##-> compare results and store discrepancies in a file
}

func GetAllAccounts() (accounts []model.Account, err error) {
	chainAddress, ok := os.LookupEnv("INDEXER_COSMOS_LCD_URL")
	if !ok {
		return accounts, fmt.Errorf("INDEXER_COSMOS_LCD_URL env is not set")
	}
	suplog.Infoln("chainAddress: ", chainAddress)

	resp, err := http.Get(fmt.Sprintf("%s/cosmos/auth/v1beta1/accounts", chainAddress))
	if err != nil {
		return accounts, err
	}

	if resp.StatusCode != http.StatusOK {
		panic(err)
	}

	err = json.NewDecoder(resp.Body).Decode(&accounts)
	if err != nil {
		return accounts, err
	}
	suplog.Infof("fond %d accounts", len(accounts))
	return accounts, nil
}

// connect to mongo repset db or return error
func mongoConnect() (db *mongo.Database, err error) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017,mongodb://localhost:27018")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}
	suplog.Infoln("Connected to MongoDB!")

	return client.Database("exchangeV2"), nil
}
