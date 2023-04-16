package main

import (
	"database/sql"
	"fmt"
	"strings"
)

type IDatabase interface {
	GetSampleListOfCompaniesForIndustry(industry *string) (*[]ProcessedCompany, error)
}

type Database struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func (d *Database) GetSampleListOfCompaniesForIndustry(industry *string) (*[]ProcessedCompany, error) {
	conn, err := d.getDatabaseConnection()
	if err != nil {
		return nil, fmt.Errorf("error fetching database connection: %s", err)
	}

	defer conn.Close()

	var companies []ProcessedCompany

	template := `
	SELECT cb."CompanyName", 
		sc."CompanyNumber",
		cb."RegAddress.AddressLine1",
		cb."RegAddress.AddressLine2",
		cb."RegAddress.PostTown",
		cb."RegAddress.PostCode",
		cb."IncorporationDate",
		cb."Accounts.AccountCategory"
	FROM "sic_company" sc
	TABLESAMPLE SYSTEM (10)
	JOIN "company_base" cb
		ON sc."SicDescription" = $1
		AND sc."Snapshot" = '2023-03-01'
		AND cb."CompanyNumberSnapshot" = sc."CompanyNumberSnapshot"
		AND cb."CompanyStatus" = 'Active'
		AND cb."Accounts.AccountCategory" != 'DORMANT'
	ORDER BY RANDOM()
	LIMIT 10
	`

	rows, err := conn.Query(template, *industry)
	if err != nil {
		return nil, fmt.Errorf("query error: %s", err)
	}

	for rows.Next() {
		var companyRow CompanyRow

		err = rows.Scan(
			&companyRow.CompanyName,
			&companyRow.CompanyNumber,
			&companyRow.AddressLine1,
			&companyRow.AddressLine2,
			&companyRow.PostTown,
			&companyRow.PostCode,
			&companyRow.IncorporationDate,
			&companyRow.AccountCategory,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning db row: %s", err)
		}

		processedCompany := ProcessedCompany{
			Name:              companyRow.CompanyName,
			CompaniesHouseUrl: fmt.Sprintf("https://find-and-update.company-information.service.gov.uk/company/%s", companyRow.CompanyNumber),
			Address:           generateAddress(companyRow.AddressLine1, companyRow.AddressLine2, companyRow.PostTown, companyRow.PostCode),
			IncorporationDate: companyRow.IncorporationDate,
			Size:              calculateCompanySize(companyRow.AccountCategory),
		}

		companies = append(companies, processedCompany)
	}

	return &companies, nil
}

type CompanyRow struct {
	CompanyName       string
	CompanyNumber     string
	AddressLine1      string
	AddressLine2      string
	PostTown          string
	PostCode          string
	AccountCategory   string
	IncorporationDate string
}

type ProcessedCompany struct {
	Name              string `json:"name"`
	CompaniesHouseUrl string `json:"companiesHouseUrl"`
	Address           string `json:"address"`
	Size              string `json:"size"`
	IncorporationDate string `json:"incorporationDate"`
}

func (d *Database) getDatabaseConnection() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", d.Host, d.Port, d.User, d.Password, d.Name)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %s", err)
	}

	return db, nil
}

func calculateCompanySize(accountCategory string) string {
	return AccountRankingToSize[CompanyAccountRanking[accountCategory]]
}

func generateAddress(addressEntries ...string) string {
	nonEmptyAddressEntries := []string{}

	for _, entry := range addressEntries {
		entryWithoutWhitespace := strings.TrimSpace(entry)

		if len(entryWithoutWhitespace) > 0 {
			nonEmptyAddressEntries = append(nonEmptyAddressEntries, entry)
		}
	}

	return strings.Join(nonEmptyAddressEntries, ", ")
}

var AccountRankingToSize = map[int]string{
	0: "very small",
	1: "small",
	2: "medium",
	3: "large",
}

/*
Small, unaudited abridged, total exemption full and total exemption small – has a turnover of GBP 10.2m or less, GBP 5.1m or less on its balance sheet and has 50 employees or less (#809) – two or more must apply
Microentity – has a  turnover of GBP 632k or less, GBP 316k or less on its balance sheet and has 10 employees or less (#733) – two or more must apply
Dormant – not doing business and doesn’t have any other income (#147)
Full – has a turnover of above GBP 10.2m or does not satisfy two or more of the criteria required to be a micro-entity or small company (#14)
*/

var CompanyAccountRanking = map[string]int{
	"ACCOUNTS TYPE NOT AVAILABLE": 0,
	"AUDITED ABRIDGED":            2,
	"AUDIT EXEMPTION SUBSIDIARY":  3,
	"DORMANT":                     0,
	"FILING EXEMPTION SUBSIDIARY": 3,
	"FULL":                        3,
	"GROUP":                       3,
	"INITIAL":                     3,
	"MEDIUM":                      3,
	"MICRO ENTITY":                1,
	"NO ACCOUNTS FILED":           0,
	"PARTIAL EXEMPTION":           2,
	"SMALL":                       2,
	"TOTAL EXEMPTION FULL":        2,
	"TOTAL EXEMPTION SMALL":       2,
	"UNAUDITED ABRIDGED":          2,
}
