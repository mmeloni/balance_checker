package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/xlab/suplog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"log"
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

		//##-> compare results
		eq1 := comp1.IsEqual(comp2)
		if !eq1 {
			suplog.Errorf("account %s is not equal in mongodb and portfolio api\n", account.BaseAccount.Address)
		}
		eq2 := comp2.IsEqual(comp3)
		if !eq2 {
			suplog.Warningf("account %s is not equal in portfolio api and chain\n", account.BaseAccount.Address)
		}
		eq3 := comp1.IsEqual(comp3)
		if !eq3 {
			suplog.Warningf("account %s is not equal in mongodb and chain\n", account.BaseAccount.Address)
		}

		if !eq1 || !eq2 || !eq3 {
			suplog.Debugf("mongoRaw:\t\t%+v\n", comp1)
			suplog.Debugf("portfolioAPI:\t\t%+v\n", comp2)
			suplog.Debugf("chain:\t\t\t%+v\n", comp3)
			appendComparableToFile(comp1, comp2, comp3)
		}

		fmt.Printf("processing account %d/%d: %s\n", id+1, len(accounts), account.BaseAccount.Address)
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

	//PrintBody(resp)

	respAccount := model.RespAccounts{}
	err = json.NewDecoder(resp.Body).Decode(&respAccount)
	if err != nil {
		return accounts, err
	}
	suplog.Infof("found %d accounts", len(respAccount.Accounts))
	return respAccount.Accounts, nil
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

func PrintBody(resp *http.Response) {

	b, err := io.ReadAll(resp.Body)
	// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(string(b))
}

func appendComparableToFile(comp ...*model.ComparablePortfolio) {
	for id, c := range comp {
		//##-> append to file
		f, err := os.OpenFile("balance_check.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		_, err = f.WriteString("Detected discrepancy:\n")
		if err != nil {
			panic(err)
		}
		switch id {
		case 0:
			if _, err = f.WriteString("mongoRaw:\n"); err != nil {
				panic(err)
			}
		case 1:
			if _, err = f.WriteString("portfolioAPI:\n"); err != nil {
				panic(err)
			}
		case 2:
			if _, err = f.WriteString("chain:\n"); err != nil {
				panic(err)
			}
		}
		if _, err = f.WriteString(fmt.Sprintf("%+v\n", c)); err != nil {
			panic(err)
		}
	}
}
