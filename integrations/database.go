package integrations

import (
	"company-data-backend/routes"
	"context"
	"strconv"

	"database/sql"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

/*
use new library
do the multiple query thing
*/

type Database struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func (d *Database) GetCompaniesAndOwnershipForIndustry(industry *string, isSample bool) (routes.CompaniesAndOwnershipQueryResults, error) {
	results := routes.CompaniesAndOwnershipQueryResults{}

	conn, err := d.getDatabaseConnection()
	if err != nil {
		return results, fmt.Errorf("error fetching database connection: %s", err)
	}

	defer conn.Close(context.Background())

	err = createCompanyListTemporaryTable(conn, industry, isSample)
	if err != nil {
		return results, fmt.Errorf("create company list temp table error: %s", err)
	}

	err = addCompanyDataToResults(conn, &results)
	if err != nil {
		return results, fmt.Errorf("get companies error: %s", err)
	}

	err = addPSCDataToResults(conn, &results)
	if err != nil {
		return results, fmt.Errorf("get PSC error: %s", err)
	}

	return results, nil
}

type CompanyRow struct {
	CompanyName            string
	CompanyNumber          string
	AddressLine1           sql.NullString
	AddressLine2           sql.NullString
	PostTown               sql.NullString
	PostCode               sql.NullString
	IncorporationDate      string
	Size                   string
	MortgageCharges        string
	MortgagesOutstanding   string
	MortgagesPartSatisfied string
	MortgagesSatisfied     string
	LastAccountsDate       string
	NextAccountsDate       string
}

func (d *Database) getDatabaseConnection() (*pgx.Conn, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", d.Host, d.Port, d.User, d.Password, d.Name)

	db, err := pgx.Connect(context.Background(), connStr)
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

func getCompanyListQuery(sample bool) string {
	var template string

	const createFilteredCompanyListTableQuery = `
select 
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
into
	temp table cdr
from
	"ch_company_2023_05_01" co
%s
join "accounts_to_size" acs on
	co."Accounts.AccountCategory" = acs."accountcategory"
where
	$1 in (
		co."SICCode.SicText_1", co."SICCode.SicText_2", co."SICCode.SicText_3", co."SICCode.SicText_4"
	)
	and co."CompanyStatus" = 'Active'
	and acs."size" <> 'no accounts available / dormant'
%s
%s;
`

	if sample {
		template = fmt.Sprintf(createFilteredCompanyListTableQuery, "TABLESAMPLE SYSTEM (10)", "ORDER BY RANDOM()", "LIMIT 10")
	} else {
		template = fmt.Sprintf(createFilteredCompanyListTableQuery, "", "", "")
	}

	return template
}

func createCompanyListTemporaryTable(conn *pgx.Conn, industry *string, isSample bool) error {
	template := getCompanyListQuery(isSample)

	_, err := conn.Exec(context.Background(), template, *industry)

	return err
}

func addCompanyDataToResults(conn *pgx.Conn, results *routes.CompaniesAndOwnershipQueryResults) error {

	rows, err := conn.Query(context.Background(), `SELECT * FROM cdr;`)
	if err != nil {
		return fmt.Errorf("get companies error: %s", err)
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
			&companyRow.Size,
			&companyRow.MortgageCharges,
			&companyRow.MortgagesOutstanding,
			&companyRow.MortgagesPartSatisfied,
			&companyRow.MortgagesSatisfied,
			&companyRow.LastAccountsDate,
			&companyRow.NextAccountsDate,
		)
		if err != nil {
			return fmt.Errorf("error scanning db row: %s", err)
		}

		mortgageCharges, err := strconv.Atoi(companyRow.MortgageCharges)
		if err != nil {
			return fmt.Errorf("unable to convert mortgage charges column value to int: %s", err)
		}

		mortgagesOutstanding, err := strconv.Atoi(companyRow.MortgagesOutstanding)
		if err != nil {
			return fmt.Errorf("unable to convert mortgage charges column value to int: %s", err)
		}

		mortgagesPartSatisfied, err := strconv.Atoi(companyRow.MortgagesPartSatisfied)
		if err != nil {
			return fmt.Errorf("unable to convert mortgage charges column value to int: %s", err)
		}

		mortgagesSatisfied, err := strconv.Atoi(companyRow.MortgagesSatisfied)
		if err != nil {
			return fmt.Errorf("unable to convert mortgage charges column value to int: %s", err)
		}

		processedCompany := routes.Company{
			Name:                   companyRow.CompanyName,
			CompaniesHouseUrl:      fmt.Sprintf("https://find-and-update.company-information.service.gov.uk/company/%s", companyRow.CompanyNumber),
			Address:                generateAddress(companyRow.AddressLine1, companyRow.AddressLine2, companyRow.PostTown, companyRow.PostCode),
			IncorporationDate:      companyRow.IncorporationDate,
			Size:                   companyRow.Size,
			MortgageCharges:        mortgageCharges,
			MortgagesOutstanding:   mortgagesOutstanding,
			MortgagesPartSatisfied: mortgagesPartSatisfied,
			MortgagesSatisfied:     mortgagesSatisfied,
			LastAccountsDate:       companyRow.LastAccountsDate,
			NextAccountsDate:       companyRow.NextAccountsDate,
		}

		results.Companies = append(results.Companies, processedCompany)
	}

	return nil
}

func addPSCDataToResults(conn *pgx.Conn, results *routes.CompaniesAndOwnershipQueryResults) error {
	query := `
select
	psc.company_number,
	psc."data.address.premises" ,
	psc."data.address.address_line_1" ,
	psc."data.address.address_line_2" ,
	psc."data.address.locality" ,
	psc."data.address.postal_code",
	psc."data.country_of_residence" ,
	psc."data.date_of_birth.month" ,
	psc."data.date_of_birth.year",
	psc."data.kind" ,
	psc."data.name" ,
	psc."data.nationality" ,
	psc."data.natures_of_control.0" ,
	psc."data.notified_on"
from
	"ch_psc_2023_05_03" psc
join cdr on
	psc."company_number" = cdr."CompanyNumber"
where
	psc."data.ceased" is null
	and psc."data.ceased_on" is null
;
`

	_, err := conn.Query(context.Background(), query)
	if err != nil {
		return fmt.Errorf("get companies error: %s", err)
	}

	return nil
}
