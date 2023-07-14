package main

type Company struct {
	Name                   string `json:"name"`
	CompaniesHouseUrl      string `json:"companiesHouseUrl"`
	Address                string `json:"address"`
	Size                   string `json:"size"`
	IncorporationDate      string `json:"incorporationDate"`
	MortgageCharges        int    `json:"mortgageCharges"`
	MortgagesOutstanding   int    `json:"mortgagesOutstanding"`
	MortgagesPartSatisfied int    `json:"mortgagesPartSatisfied"`
	MortgagesSatisfied     int    `json:"mortgagesSatisfied"`
	LastAccountsDate       string `json:"lastAccountsDate"`
	NextAccountsDate       string `json:"nextAccountsDate"`
}

type PersonWithSignificantControl struct{}
