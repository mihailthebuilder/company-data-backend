package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	if isRunningLocally() {
		loadEnvironmentVariablesFromDotEnvFile()
	}

	runApplication()
}

func isRunningLocally() bool {
	return os.Getenv("GIN_MODE") == ""
}

func loadEnvironmentVariablesFromDotEnvFile() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func getEnv(env string) string {
	val := os.Getenv(env)
	if val == "" {
		log.Fatal("Environment variable not set:", env)
	}
	return val
}

func runApplication() {
	r := gin.Default()

	serverRecoversFromAnyPanicAndWrites500(r)
	allowAllOriginsForCORS(r)

	r.GET("/companies/sample", handleRequestForCompaniesSample)

	r.Run()
}

func serverRecoversFromAnyPanicAndWrites500(engine *gin.Engine) {
	engine.Use(gin.Recovery())
}

func allowAllOriginsForCORS(engine *gin.Engine) {
	engine.Use(cors.Default())
}

func handleRequestForCompaniesSample(c *gin.Context) {

	sicDescription := c.Query("SicDescription")
	if len(sicDescription) == 0 || len(sicDescription) > 158 {
		log.Println("Invalid industry request: ", sicDescription)
		c.JSON(http.StatusBadRequest, fmt.Sprintf("Invalid industry: %s", sicDescription))
		return
	}

	sample := getCompaniesSample(&sicDescription)

	c.JSON(http.StatusOK, sample)
}

type RequestBody struct {
	SicDescription string `json:"sicDescription"`
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

func getCompaniesSample(sic *string) []ProcessedCompany {
	connectToDatabase()
	defer dbConn.Close()

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
	LIMIT 20
	`

	rows, err := dbConn.Query(template, *sic)
	if err != nil {
		log.Panic("Query failure", err)
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
			log.Panic("Error scanning db row: ", err)
		}

		processedCompany := ProcessedCompany{
			Name:              companyRow.CompanyName,
			CompaniesHouseUrl: fmt.Sprintf("https://find-and-update.company-information.service.gov.uk/company/%s", companyRow.CompanyNumber),
			Address:           fmt.Sprintf("%s,%s,%s,%s", companyRow.AddressLine1, companyRow.AddressLine2, companyRow.PostTown, companyRow.PostCode),
			IncorporationDate: companyRow.IncorporationDate,
			Size:              getCompanySize(companyRow.AccountCategory),
		}

		companies = append(companies, processedCompany)
	}

	return companies
}

var dbConn *sql.DB

func connectToDatabase() {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", getEnv("DB_HOST"), getEnv("DB_PORT"), getEnv("DB_USER"), getEnv("DB_PASSWORD"), getEnv("DB_NAME"))

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Panic("Error opening database connection: ", err)
	}

	dbConn = db
}

func getCompanySize(accountCategory string) string {
	return AccountRankingToSize[CompanyAccountRanking[accountCategory]]
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
