package main

import (
	"database/sql"
	"fmt"
	"strings"
)

type IDatabase interface {
	GetListOfCompanies(industry *string, isSample bool) ([]ProcessedCompany, error)
}

type Database struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func (d *Database) GetListOfCompanies(industry *string, isSample bool) ([]ProcessedCompany, error) {
	conn, err := d.getDatabaseConnection()
	if err != nil {
		return nil, fmt.Errorf("error fetching database connection: %s", err)
	}

	defer conn.Close()

	var companies []ProcessedCompany

	template := getQueryTemplate(isSample)

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

	return companies, nil
}

type CompanyRow struct {
	CompanyName       string
	CompanyNumber     string
	AddressLine1      sql.NullString
	AddressLine2      sql.NullString
	PostTown          sql.NullString
	PostCode          sql.NullString
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

func generateAddress(addressEntries ...sql.NullString) string {
	nonEmptyAddressEntries := []string{}

	for _, entry := range addressEntries {
		if !entry.Valid {
			continue
		}

		entryWithoutWhitespace := strings.TrimSpace(entry.String)

		if len(entryWithoutWhitespace) > 0 {
			nonEmptyAddressEntries = append(nonEmptyAddressEntries, entry.String)
		}
	}

	return strings.Join(nonEmptyAddressEntries, ", ")
}

var AccountRankingToSize = map[int]string{
	0: "no accounts available / dormant",
	1: "micro",
	2: "small",
	3: "medium",
	4: "large",
	5: "very large",
}

/*
Small, unaudited abridged, total exemption full and total exemption small – has a turnover of GBP 10.2m or less, GBP 5.1m or less on its balance sheet and has 50 employees or less (#809) – two or more must apply
Microentity – has a  turnover of GBP 632k or less, GBP 316k or less on its balance sheet and has 10 employees or less (#733) – two or more must apply
Dormant – not doing business and doesn’t have any other income (#147)
Full – has a turnover of above GBP 10.2m or does not satisfy two or more of the criteria required to be a micro-entity or small company (#14)
*/

var CompanyAccountRanking = map[string]int{
	"ACCOUNTS TYPE NOT AVAILABLE": 0,
	"NO ACCOUNTS FILED":           0,
	"DORMANT":                     0,
	"MICRO ENTITY":                1,
	"SMALL":                       2,
	"TOTAL EXEMPTION SMALL":       2,
	"TOTAL EXEMPTION FULL":        2,
	"MEDIUM":                      3,
	"FULL":                        4,
	"GROUP":                       5,

	"AUDITED ABRIDGED":            4,
	"AUDIT EXEMPTION SUBSIDIARY":  2,
	"FILING EXEMPTION SUBSIDIARY": 2,
	"INITIAL":                     2,
	"PARTIAL EXEMPTION":           2,
	"UNAUDITED ABRIDGED":          2,
}

func getQueryTemplate(sample bool) string {
	var template string

	if sample {
		template = fmt.Sprintf(QUERY_TEMPLATE, "TABLESAMPLE SYSTEM (10)", "ORDER BY RANDOM()", "LIMIT 10")
	} else {
		template = fmt.Sprintf(QUERY_TEMPLATE, "", "", "")
	}

	return template
}

const QUERY_TEMPLATE = `
SELECT 
	"CompanyName",
	"CompanyNumber",
	"RegAddress.AddressLine1",
	"RegAddress.AddressLine2",
	"RegAddress.PostTown",
	"RegAddress.PostCode",
	"IncorporationDate",
	"Accounts.AccountCategory"
FROM "ch_company_2023_05_01"
%s
WHERE
	$1 IN (
		"SICCode.SicText_1", "SICCode.SicText_2", "SICCode.SicText_3", "SICCode.SicText_4"
	)
	AND "CompanyStatus" = 'Active'
	AND "Accounts.AccountCategory" NOT IN ('DORMANT','NO ACCOUNTS FILED','ACCOUNTS TYPE NOT AVAILABLE')
%s
%s
;
`
