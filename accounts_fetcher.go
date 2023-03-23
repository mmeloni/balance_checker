package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/xlab/suplog"
	"net/http"
	"os"
	"portfolio-bank-balance-checks/model"
	"strconv"
)

func GetAllAccounts(page string) (nextPage string, accounts []model.Account, err error) {
	chainAddress, ok := os.LookupEnv("INDEXER_COSMOS_LCD_URL")
	if !ok {
		return "", accounts, fmt.Errorf("INDEXER_COSMOS_LCD_URL env is not set")
	}

	suplog.Infoln("chainAddress: ", chainAddress)

	queryString := fmt.Sprintf("%s/cosmos/auth/v1beta1/accounts", chainAddress)
	if page != "" {
		page = base64.StdEncoding.EncodeToString([]byte(page))
		queryString = fmt.Sprintf("%s?pagination.key=%s", queryString, page)
	}

	resp, err := http.Get(queryString)
	if err != nil {
		return "", accounts, err
	}

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Errorf("status code is not 200, it is %d", resp.StatusCode))
	}

	//PrintBody(resp)

	respAccount := model.RespAccounts{}
	err = json.NewDecoder(resp.Body).Decode(&respAccount)
	if err != nil {
		return "", accounts, err
	}
	suplog.Infof("found %d accounts", len(respAccount.Accounts))

	nextPage = ""
	if respAccount.Pagination.NextKey != "" {
		nextPage = respAccount.Pagination.NextKey
	}

	if TotalAccounts == 0 {
		TotalAccounts, err = strconv.Atoi(respAccount.Pagination.Total)
		if err != nil {
			return "", accounts, fmt.Errorf("error converting total accounts to int: %w", err)
		}
	}

	return nextPage, respAccount.Accounts, nil
}

func getMoreAccounts(ch chan []model.Account, done chan bool) {
	nextPage, accounts, err := GetAllAccounts("")
	if err != nil {
		panic(err)
	}
	ch <- accounts
	for nextPage != "" {
		nextPage, accounts, err = GetAllAccounts(nextPage)
		if err != nil {
			panic(err)
		}
		ch <- accounts
	}
	close(ch)
	done <- true
}
