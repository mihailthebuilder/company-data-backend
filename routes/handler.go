package routes

import (
	"io"
	"net/http"
)

type RouteHandler struct {
	EmailAPI                  IEmailAPI
	JwtTokenLifespanInMinutes string
	ApiSecret                 string
	Database                  IDatabase
}

type IDatabase interface {
	GetCompaniesAndOwnershipForIndustry(industry *string, isSample bool) (CompaniesAndOwnershipQueryResults, error)
}

type IEmailAPI interface {
	SendRequest(body io.Reader) (*http.Response, error)
}

type CompaniesAndOwnershipQueryResults struct {
	Companies []Company `json:"companies"`
	PSCs      []PSC     `json:"personsWithSignificantControl"`
}

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

type PSC struct{}
