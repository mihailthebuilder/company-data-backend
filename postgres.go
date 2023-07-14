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
		var row CompanyDbRow

		err = rows.Scan(
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
			return nil, fmt.Errorf("error scanning db row: %s", err)
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

		setValidStringField(row.PostTown, &company.Town)
		setValidStringField(row.PostCode, &company.Postcode)
		setValidInt32Field(row.Employees, &company.Employees)
		setValidInt32Field(row.Equity, &company.Equity)
		setValidInt32Field(row.NetCurrentAssets, &company.NetCurrentAssets)
		setValidInt32Field(row.FixedAssets, &company.FixedAssets)

		if company.FixedAssets == 0 {
			if row.TotalAssetsLessCurrentLiabilities.Valid {
				company.FixedAssets = int(row.TotalAssetsLessCurrentLiabilities.Int32)

				if row.NetCurrentAssets.Valid {
					company.FixedAssets = company.FixedAssets - int(row.NetCurrentAssets.Int32)
				} else if row.Cash.Valid {
					company.FixedAssets = company.FixedAssets - int(row.Cash.Int32)
				}
			}
		}

		setValidInt32Field(row.Cash, &company.Cash)

		companies = append(companies, company)
	}

	return &companies, nil
}

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
	Employees                         sql.NullInt32
	Equity                            sql.NullInt32
	NetCurrentAssets                  sql.NullInt32
	TotalAssetsLessCurrentLiabilities sql.NullInt32
	FixedAssets                       sql.NullInt32
	Cash                              sql.NullInt32
}

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

func setValidStringField(valid sql.NullString, field *string) {
	if valid.Valid {
		*field = valid.String
	}
}

func setValidInt32Field(valid sql.NullInt32, field *int) {
	if valid.Valid {
		*field = int(valid.Int32)
	}
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
;
`
