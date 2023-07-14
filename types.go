package main

type Company struct {
	Name               string `json:"name"`
	CompaniesHouseUrl  string `json:"companiesHouseUrl"`
	Address            string `json:"address",omitempty`
	Town               string `json:"town",omitempty`
	Postcode           string `json:"postcode",omitempty`
	IncorporationDate  string `json:"incorporationDate",omitempty`
	MortgageCharges    int    `json:"mortgageCharges",omitempty`
	AverageAgeOfOwners int    `json:"averageAgeOfOwners",omitempty`
	LastAccountsDate   string `json:"lastAccountsDate",omitempty`
	Employees          int    `json:"employees",omitempty`
	Equity             int    `json:"equity",omitempty`
	NetCurrentAssets   int    `json:"netAssets",omitempty`
	FixedAssets        int    `json:"fixedAssets",omitempty`
	Cash               int    `json:"cash",omitempty`
}
