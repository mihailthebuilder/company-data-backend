package main

import (
	"database/sql"
	"fmt"
	"log"
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
		company, err := getCompany(rows)
		if err != nil {
			return nil, fmt.Errorf("error getting company from row: %s", err)
		}

		companies = append(companies, company)
	}

	return companies, nil
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
tablesample system (20)
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

	url := fmt.Sprintf("https://find-and-update.company-information.service.gov.uk/company/%s", row.CompanyNumber)
	address := generateAddress(row.AddressLine1, row.AddressLine2)

	company := Company{
		Name:               &row.CompanyName,
		CompaniesHouseUrl:  &url,
		Address:            &address,
		IncorporationDate:  &row.IncorporationDate,
		MortgageCharges:    &row.MortgageCharges,
		AverageAgeOfOwners: &row.AverageAge,
		LastAccountsDate:   &row.EndDate,
	}

	if row.PostTown.Valid {
		company.Town = &row.PostTown.String
	}

	if row.PostCode.Valid {
		company.Postcode = &row.PostCode.String
	}

	company.Employees = getIntValueFromString(row.Employees.String)

	if company.Employees != nil && *company.Employees < 0 {
		*company.Employees = *company.Employees * -1
	}
	if company.Employees == nil {
		*company.Employees = 0
	}

	company.Equity = getIntValueFromString(row.Equity.String)
	company.NetCurrentAssets = getIntValueFromString(row.NetCurrentAssets.String)
	company.Cash = getIntValueFromString(row.Cash.String)
	company.FixedAssets = getIntValueFromString(row.FixedAssets.String)

	if company.FixedAssets == nil {
		talcl := getIntValueFromString(row.TotalAssetsLessCurrentLiabilities.String)

		if talcl != nil && company.NetCurrentAssets != nil {
			val := *talcl - *company.NetCurrentAssets
			company.FixedAssets = &val
		}
	}

	return company, nil
}

func getIntValueFromString(str string) *int {
	if str == "" {
		return nil
	}

	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		log.Println("error parsing int value from string: ", err)
		return nil
	}

	intVal := int(val)
	return &intVal
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
