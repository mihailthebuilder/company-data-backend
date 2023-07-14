package main

type Company struct {
	Name               string `json:"name"`
	CompaniesHouseUrl  string `json:"companiesHouseUrl"`
	Address            string `json:"address"`
	Town               string `json:"town"`
	Postcode           string `json:"postcode"`
	IncorporationDate  string `json:"incorporationDate"`
	MortgageCharges    int    `json:"mortgageCharges"`
	AverageAgeOfOwners int    `json:"averageAgeOfOwners"`
	LastAccountsDate   string `json:"lastAccountsDate"`
	Employees          int    `json:"employees"`
	Equity             int    `json:"equity"`
	NetCurrentAssets   int    `json:"netAssets"`
	FixedAssets        int    `json:"fixedAssets"`
	Cash               int    `json:"cash"`
}
