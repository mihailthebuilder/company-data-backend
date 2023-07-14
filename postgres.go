package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

type Postgres struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func (d *Postgres) GetListOfCompanies(industry string, isSample bool) (*[]Company, error) {
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
		company, err := getCompany(rows)
		if err != nil {
			return nil, fmt.Errorf("error getting company from row: %s", err)
		}

		companies = append(companies, company)
	}

	return &companies, nil
}

func (d *Postgres) getDatabaseConnection() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", d.Host, d.Port, d.User, d.Password, d.Name)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %s", err)
	}

	return db, nil
}

const COMPANY_QUERY = `
select
	cc."CompanyName" ,
	cc."CompanyNumber" ,
	cc."RegAddress.AddressLine1" ,
	cc."RegAddress.AddressLine2" ,
	cc."RegAddress.PostTown",
	cc."RegAddress.PostCode",
	cc."IncorporationDate" ,
	cc."Mortgages.NumMortCharges" ,
	caa.average_age ,
	lad.end_date ,
	lad.employees ,
	coalesce(lad.equity,
	lad.net_assets_liabilities ) "equity",
	lad.net_current_assets_liabilities ,
	lad.total_assets_less_current_liabilities ,
	lad.fixed_assets ,
	lad.cash
from
	company_ch cc
join company_average_age caa on
	caa.company_number = cc."CompanyNumber"
join latest_accounts_data lad on
	lad.reference_number = cc."CompanyNumber"
where
	cc."CompanyStatus" = 'Active'
	and cc."Accounts.AccountCategory" not in ('DORMANT', 'NO ACCOUNTS FILED', 'ACCOUNTS TYPE NOT AVAILABLE')
	and cc."CompanyCategory" = 'Private Limited Company'
	and $1 in (
		cc."SICCode.SicText_1", cc."SICCode.SicText_2", cc."SICCode.SicText_3", cc."SICCode.SicText_4"
	)
limit 10
;
`

type CompanyDbRow struct {
	CompanyName                       string
	CompanyNumber                     string
	AddressLine1                      sql.NullString
	AddressLine2                      sql.NullString
	PostTown                          sql.NullString
	PostCode                          sql.NullString
	IncorporationDate                 string
	MortgageCharges                   int
	AverageAge                        int
	EndDate                           string
	Employees                         sql.NullString
	Equity                            sql.NullString
	NetCurrentAssets                  sql.NullString
	TotalAssetsLessCurrentLiabilities sql.NullString
	FixedAssets                       sql.NullString
	Cash                              sql.NullString
}

func getCompany(rows *sql.Rows) (Company, error) {
	var row CompanyDbRow

	err := rows.Scan(
		&row.CompanyName,
		&row.CompanyNumber,
		&row.AddressLine1,
		&row.AddressLine2,
		&row.PostTown,
		&row.PostCode,
		&row.IncorporationDate,
		&row.MortgageCharges,
		&row.AverageAge,
		&row.EndDate,
		&row.Employees,
		&row.Equity,
		&row.NetCurrentAssets,
		&row.TotalAssetsLessCurrentLiabilities,
		&row.FixedAssets,
		&row.Cash,
	)
	if err != nil {
		return Company{}, err
	}

	company := Company{
		Name:               row.CompanyName,
		CompaniesHouseUrl:  fmt.Sprintf("https://find-and-update.company-information.service.gov.uk/company/%s", row.CompanyNumber),
		Address:            generateAddress(row.AddressLine1, row.AddressLine2),
		IncorporationDate:  row.IncorporationDate,
		MortgageCharges:    row.MortgageCharges,
		AverageAgeOfOwners: row.AverageAge,
		LastAccountsDate:   row.EndDate,
	}

	if row.PostTown.Valid {
		company.Town = row.PostTown.String
	}

	if row.PostCode.Valid {
		company.Postcode = row.PostCode.String
	}

	if row.Employees.Valid {
		employees, err := strconv.ParseFloat(row.Employees.String, 64)
		if err == nil {
			company.Employees = int(employees)
		}
	}

	if row.Equity.Valid {
		equity, err := strconv.ParseFloat(row.Equity.String, 64)
		if err == nil {
			company.Equity = int(equity)
		}
	}

	if row.NetCurrentAssets.Valid {
		netCurrentAssets, err := strconv.ParseFloat(row.NetCurrentAssets.String, 64)
		if err == nil {
			company.NetCurrentAssets = int(netCurrentAssets)
		}
	}

	if row.Cash.Valid {
		cash, err := strconv.ParseFloat(row.Cash.String, 64)
		if err == nil {
			company.Cash = int(cash)
		}
	}

	if row.FixedAssets.Valid {
		fixedAssets, err := strconv.ParseFloat(row.FixedAssets.String, 64)
		if err == nil {
			company.FixedAssets = int(fixedAssets)
		}
	}

	if company.FixedAssets == 0 && row.TotalAssetsLessCurrentLiabilities.Valid {
		totalAssetsLessCurrentLiabilities, err := strconv.ParseFloat(row.TotalAssetsLessCurrentLiabilities.String, 64)
		if err == nil {
			company.FixedAssets = int(totalAssetsLessCurrentLiabilities)
		}

		if company.NetCurrentAssets > 0 {
			company.FixedAssets = company.FixedAssets - company.NetCurrentAssets
		} else if company.Cash > 0 {
			company.FixedAssets = company.FixedAssets - company.Cash
		}
	}

	return company, nil
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
