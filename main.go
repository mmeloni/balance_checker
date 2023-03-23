package main

import (
	"context"
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
var TotalAccounts int

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

	//##-> stop main net stage exchange process and retrieve {latest_block}
	//##-> fetch all accounts
	ch := make(chan []model.Account)
	done := make(chan bool)
	go getMoreAccounts(ch, done)

	processedCount := 0
	discrepanciesCount := 0

	for {
		select {
		case <-done:
			suplog.Infof("processed %d accounts", processedCount)
			suplog.Infof("found %d discrepancies", discrepanciesCount)
			return
		case accounts := <-ch:
			for _, account := range accounts {
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
				if err != nil {
					panic(err)
				}
				//##-> compare results
				eq1, diff1 := comp1.IsEqual(comp2)
				diffMsg1, diffMsg2, diffMsg3 := "", "", ""
				if !eq1 {
					diffMsg1 = fmt.Sprintf("account %s is not equal in mongodb and portfolio api: %s", account.BaseAccount.Address, diff1)
					suplog.Errorf("account %s is not equal in mongodb and portfolio api: %s", account.BaseAccount.Address, diff1)
				}
				eq2, diff2 := comp2.IsEqual(comp3)
				if !eq2 {
					diffMsg2 = fmt.Sprintf("account %s is not equal in portfolio api and chain api:%s ", account.BaseAccount.Address, diff2)
					suplog.Warningf("account %s is not equal in portfolio api and chain api:%s ", account.BaseAccount.Address, diff2)
				}
				eq3, diff3 := comp1.IsEqual(comp3)
				if !eq3 {
					diffMsg3 = fmt.Sprintf("account %s is not equal in mongodb and chain api: %s", account.BaseAccount.Address, diff3)
					suplog.Warningf("account %s is not equal in mongodb and chain api: %s", account.BaseAccount.Address, diff3)
				}

				if !eq1 || !eq2 || !eq3 {
					suplog.Debugf("mongoRaw:\t\t%+v\n", comp1)
					suplog.Debugf("portfolioAPI:\t\t%+v\n", comp2)
					suplog.Debugf("chain:\t\t\t%+v\n", comp3)
					appendDiffToFile(diffMsg1, diffMsg2, diffMsg3)
					discrepanciesCount++
				}
				processedCount++
				fmt.Printf("processing account %d/%d: %s\n", processedCount, TotalAccounts, account.BaseAccount.Address)
			}
		}
	}
}

// connect to mongo repset db or return error
func mongoConnect() (db *mongo.Database, err error) {
	mongoUrl := "mongodb://localhost:27017,mongodb://localhost:27018"
	if os.Getenv("MONGO_URL") != "" {
		mongoUrl = os.Getenv("MONGO_URL")
	}

	clientOptions := options.Client().ApplyURI(mongoUrl)
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
	f, err := os.OpenFile("balance_check.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString("Detected discrepancy:\n")
	if err != nil {
		panic(err)
	}
	for id, c := range comp {
		//##-> append to file
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

func appendDiffToFile(diff ...string) {
	f, err := os.OpenFile("balance_check.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString("Detected discrepancy:\n")
	if err != nil {
		panic(err)
	}
	for id, c := range diff {
		//##-> append to file
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
