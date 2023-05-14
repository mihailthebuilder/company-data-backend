package integrations

import (
	"database/sql"
	"fmt"
	"strings"
)

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
			&companyRow.Size,
			&companyRow.MortgageCharges,
			&companyRow.MortgagesOutstanding,
			&companyRow.MortgagesPartSatisfied,
			&companyRow.MortgagesSatisfied,
			&companyRow.LastAccountsDate,
			&companyRow.NextAccountsDate,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning db row: %s", err)
		}

		processedCompany := ProcessedCompany{
			Name:                   companyRow.CompanyName,
			CompaniesHouseUrl:      fmt.Sprintf("https://find-and-update.company-information.service.gov.uk/company/%s", companyRow.CompanyNumber),
			Address:                generateAddress(companyRow.AddressLine1, companyRow.AddressLine2, companyRow.PostTown, companyRow.PostCode),
			IncorporationDate:      companyRow.IncorporationDate,
			Size:                   companyRow.Size,
			MortgageCharges:        companyRow.MortgageCharges,
			MortgagesOutstanding:   companyRow.MortgagesOutstanding,
			MortgagesPartSatisfied: companyRow.MortgagesPartSatisfied,
			MortgagesSatisfied:     companyRow.MortgagesSatisfied,
			LastAccountsDate:       companyRow.LastAccountsDate,
			NextAccountsDate:       companyRow.NextAccountsDate,
		}

		companies = append(companies, processedCompany)
	}

	return companies, nil
}

type CompanyRow struct {
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

type ProcessedCompany struct {
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

func (d *Database) getDatabaseConnection() (*sql.DB, error) {
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
	'Accounting and auditing activities' in (
		co."SICCode.SicText_1", co."SICCode.SicText_2", co."SICCode.SicText_3", co."SICCode.SicText_4"
	)
	and co."CompanyStatus" = 'Active'
	and acs."size" <> 'no accounts available / dormant'
%s
%s;

SELECT * FROM cdr;

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
