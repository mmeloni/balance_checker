package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Portfolio struct {
	AccountAddress string               `json:"accountAddress" bson:"accountAddress"`
	SubaccountId   string               `json:"subaccountId" bson:"subaccountId"`
	Denom          string               `json:"denom" bson:"denom"`
	Amount         primitive.Decimal128 `json:"amount" bson:"amount"`
	Type           int8                 `json:"type" bson:"type"`
	Timestamp      int64                `json:"timestamp" bson:"timestamp"`
}

/*
	{
	  "portfolio": {
	    "accountAddress": "inj1cml96vmptgw99syqrrz8az79xer2pcgp0a885r",
	    "bankBalances": [
	      {
	        "amount": "1000000000000000000",
	        "denom": "USDT"
	      },
	      {
	        "amount": "1000000000000000000",
	        "denom": "USDT"
	      },
	      {
	        "amount": "1000000000000000000",
	        "denom": "USDT"
	      },
	      {
	        "amount": "1000000000000000000",
	        "denom": "USDT"
	      }
	    ],
	    "positionsWithUPNL": [
	      {
	        "position": {
	          "aggregateReduceOnlyQuantity": "0",
	          "createdAt": 1544614248000,
	          "direction": "long",
	          "entryPrice": "15333333.333333333333333333",
	          "liquidationPrice": "23420479",
	          "margin": "77000000",
	          "markPrice": "14000000",
	          "marketId": "0x3bdb3d8b5eb4d362371b72cf459216553d74abdb55eb0208091f7777dd85c8bb",
	          "quantity": "9",
	          "subaccountId": "0x90f8bf6a479f320ead074411a4b0e7944ea8c9c1000000000000000000000002",
	          "ticker": "INJ/USDT-PERP",
	          "updatedAt": 1544614248000
	        },
	        "unrealizedPNL": "0"
	      },
	      {
	        "position": {
	          "aggregateReduceOnlyQuantity": "0",
	          "createdAt": 1544614248000,
	          "direction": "long",
	          "entryPrice": "15333333.333333333333333333",
	          "liquidationPrice": "23420479",
	          "margin": "77000000",
	          "markPrice": "14000000",
	          "marketId": "0x3bdb3d8b5eb4d362371b72cf459216553d74abdb55eb0208091f7777dd85c8bb",
	          "quantity": "9",
	          "subaccountId": "0x90f8bf6a479f320ead074411a4b0e7944ea8c9c1000000000000000000000002",
	          "ticker": "INJ/USDT-PERP",
	          "updatedAt": 1544614248000
	        },
	        "unrealizedPNL": "0"
	      }
	    ],
	    "subaccounts": [
	      {
	        "denom": "peggy0xdAC17F958D2ee523a2206206994597C13D831ec7",
	        "deposit": {
	          "availableBalance": "1000000000000000000",
	          "totalBalance": "1960000000000000000"
	        },
	        "subaccountId": "0x90f8bf6a479f320ead074411a4b0e7944ea8c9c1000000000000000000000002"
	      },
	      {
	        "denom": "peggy0xdAC17F958D2ee523a2206206994597C13D831ec7",
	        "deposit": {
	          "availableBalance": "1000000000000000000",
	          "totalBalance": "1960000000000000000"
	        },
	        "subaccountId": "0x90f8bf6a479f320ead074411a4b0e7944ea8c9c1000000000000000000000002"
	      },
	      {
	        "denom": "peggy0xdAC17F958D2ee523a2206206994597C13D831ec7",
	        "deposit": {
	          "availableBalance": "1000000000000000000",
	          "totalBalance": "1960000000000000000"
	        },
	        "subaccountId": "0x90f8bf6a479f320ead074411a4b0e7944ea8c9c1000000000000000000000002"
	      },
	      {
	        "denom": "peggy0xdAC17F958D2ee523a2206206994597C13D831ec7",
	        "deposit": {
	          "availableBalance": "1000000000000000000",
	          "totalBalance": "1960000000000000000"
	        },
	        "subaccountId": "0x90f8bf6a479f320ead074411a4b0e7944ea8c9c1000000000000000000000002"
	      }
	    ]
	  }
	}
*/
type PortfolioApiResponse struct {
	Portfolio PortfolioAPI `json:"portfolio"`
}

type PortfolioAPI struct {
	AccountAddress string       `json:"accountAddress"`
	BankBalances   []Balance    `json:"bankBalances"`
	Subaccounts    []Subaccount `json:"subaccounts"`
}

type Balance struct {
	Amount string `json:"amount"`
	Denom  string `json:"denom"`
}

type Subaccount struct {
	Denom      string `json:"denom"`
	Subaccount string `json:"subaccountId"`
	Deposit    struct {
		AvailableBalance string `json:"availableBalance"`
		TotalBalance     string `json:"totalBalance"`
	}
}
