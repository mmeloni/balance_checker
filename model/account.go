package model

/*
	{
	  "accounts": [
	    {
	      "@type": "/cosmos.auth.v1beta1.ModuleAccount",
	      "base_account": {
	        "address": "inj17xpfvakm2amg962yls6f84z3kell8c5l6s5ye9",
	        "pub_key": null,
	        "account_number": "18",
	        "sequence": "0"
	      },
	      "name": "fee_collector",
	      "permissions": [
	      ]
	    }
	  ],
	  "pagination": {
	    "next_key": null,
	    "total": "30"
	  }
	}
*/
type RespAccounts struct {
	Accounts []Account `json:"accounts"`
}

type Account struct {
	BaseAccount BaseAccount `json:"base_account"`
}

type BaseAccount struct {
	Address       string `json:"address"`
	AccountNumber string `json:"account_number"`
}
