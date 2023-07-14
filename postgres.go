package main

import (
	"database/sql"
	"fmt"
	"strings"
)

type Postgres struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func (d *Postgres) GetListOfCompanies(industry string, isSample bool) ([]Company, error) {
	conn, err := d.getDatabaseConnection()
	if err != nil {
		return nil, fmt.Errorf("error fetching database connection: %s", err)
	}

	defer conn.Close()

	var companies []Company

	rows, err := conn.Query(COMPANY_QUERY, industry)
	if err != nil {
		return nil, fmt.Errorf("query error: %s", err)
	}

	for rows.Next() {
		var row CompanyDbRow

		err = rows.Scan(
			&row.CompanyName,
			&row.CompanyNumber,
			&row.AddressLine1,
			&row.AddressLine2,
			&row.PostTown,
			&row.PostCode,
			&row.IncorporationDate,
			&row.Size,
			&row.MortgageCharges,
			&row.MortgagesOutstanding,
			&row.MortgagesPartSatisfied,
			&row.MortgagesSatisfied,
			&row.LastAccountsDate,
			&row.NextAccountsDate,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning db row: %s", err)
		}

		company := Company{
			Name:                   row.CompanyName,
			CompaniesHouseUrl:      fmt.Sprintf("https://find-and-update.company-information.service.gov.uk/company/%s", row.CompanyNumber),
			Address:                generateAddress(row.AddressLine1, row.AddressLine2, row.PostTown, row.PostCode),
			IncorporationDate:      row.IncorporationDate,
			Size:                   row.Size,
			MortgageCharges:        row.MortgageCharges,
			MortgagesOutstanding:   row.MortgagesOutstanding,
			MortgagesPartSatisfied: row.MortgagesPartSatisfied,
			MortgagesSatisfied:     row.MortgagesSatisfied,
			LastAccountsDate:       row.LastAccountsDate,
			NextAccountsDate:       row.NextAccountsDate,
		}

		companies = append(companies, company)
	}

	return companies, nil
}

type CompanyDbRow struct {
	CompanyName            string
	CompanyNumber          string
	AddressLine1           sql.NullString
	AddressLine2           sql.NullString
	PostTown               sql.NullString
	PostCode               sql.NullString
	Size                   string
	IncorporationDate      string
	MortgageCharges        int
	MortgagesOutstanding   int
	MortgagesPartSatisfied int
	MortgagesSatisfied     int
	LastAccountsDate       string
	NextAccountsDate       string
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

type PersonWithSignificantControl struct{}

func (d *Postgres) getDatabaseConnection() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", d.Host, d.Port, d.User, d.Password, d.Name)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %s", err)
	}

	return db, nil
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

const COMPANY_QUERY = `
SELECT 
	co."CompanyName",
	co."CompanyNumber",
	co."RegAddress.AddressLine1",
	co."RegAddress.AddressLine2",
	co."RegAddress.PostTown",
	co."RegAddress.PostCode",
	co."IncorporationDate",
	acs."size",
	co."Mortgages.NumMortCharges",
	co."Mortgages.NumMortOutstanding",
	co."Mortgages.NumMortPartSatisfied",
	co."Mortgages.NumMortSatisfied",
	co."Accounts.LastMadeUpDate",
	co."Accounts.NextDueDate"
FROM "company_ch" co
TABLESAMPLE SYSTEM (30)
JOIN "accounts_to_size" acs on co."Accounts.AccountCategory" = acs."accountcategory"
WHERE
	$1 IN (
		co."SICCode.SicText_1", co."SICCode.SicText_2", co."SICCode.SicText_3", co."SICCode.SicText_4"
	)
	AND co."CompanyStatus" = 'Active'
	AND acs."size" <> 'no accounts available / dormant'
ORDER BY RANDOM()
LIMIT 10
;
`

func (d *Postgres) GetListOfPersonsWithSignificantControl(*[]Company) ([]PersonWithSignificantControl, error) {
	return nil, nil
}
