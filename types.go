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
	Employees          int    `json:"employees,omitempty"`
	Equity             int    `json:"equity,omitempty"`
	NetCurrentAssets   int    `json:"netCurrentAssets,omitempty"`
	FixedAssets        int    `json:"fixedAssets,omitempty"`
	Cash               int    `json:"cash,omitempty"`
}
