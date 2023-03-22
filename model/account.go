package model

/*
{
"@type": "/injective.types.v1beta1.EthAccount",
"base_account": {
"address": "inj1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqe2hm49",
"pub_key": null,
"account_number": "2",
"sequence": "0"
},
"code_hash": "xdJGAYb3IzySfn2y3McDwOUAtlPKgic7e/rYBF2FpHA="
},
*/
type Account struct {
	BaseAccount BaseAccount `json:"base_account"`
	CodeHash    string      `json:"code_hash"`
}
type BaseAccount struct {
	Address       string `json:"address"`
	PubKey        string `json:"pub_key"`
	AccountNumber string `json:"account_number"`
	Sequence      string `json:"sequence"`
}
