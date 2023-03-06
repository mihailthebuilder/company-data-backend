package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

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

	r.GET("/companies/sic_code/:sic_code", handleCompaniesBySicCodeRequest)

	r.Run()
}

func serverRecoversFromAnyPanicAndWrites500(engine *gin.Engine) {
	engine.Use(gin.Recovery())
}

func handleCompaniesBySicCodeRequest(c *gin.Context) {
	connectToDatabase()
	defer dbConn.Close()

	sic := c.Param("sic_code")

	valid := isValidSicFormat(&sic)
	if !valid {
		c.JSON(http.StatusBadRequest, fmt.Sprintf("Invalid SIC code: %s", sic))
		return
	}

	processCompaniesBySicCodeRequest(&sic, c)
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

func isValidSicFormat(sic *string) bool {
	pattern := "^[0-9]+$"
	match, err := regexp.MatchString(pattern, *sic)

	if err != nil {
		log.Panic("Error validating SIC code:", err)
	}

	return match
}

type Company struct {
	CompanyName       string
	CompanyNumber     string
	AddressLine1      string
	AddressLine2      string
	PostTown          string
	PostCode          string
	CompanyStatus     string
	IncorporationDate string
}

func processCompaniesBySicCodeRequest(sic *string, c *gin.Context) {
	var company Company
	var companies []Company

	template := `
	SELECT cb."CompanyName", 
		sc."CompanyNumber",
		cb."RegAddress.AddressLine1",
		cb."RegAddress.AddressLine2",
		cb."RegAddress.PostTown",
		cb."RegAddress.PostCode",
		cb."CompanyStatus",
		cb."IncorporationDate"
	FROM "sic_company" sc
	TABLESAMPLE SYSTEM (0.5)
	JOIN "company_base" cb
		ON sc."SicCode" = $1
		AND sc."Snapshot" = '2023-03-01'
		AND cb."CompanyNumberSnapshot" = sc."CompanyNumberSnapshot"
	ORDER BY RANDOM()
	LIMIT 10
	`

	rows, err := dbConn.Query(template, *sic)
	if err != nil {
		log.Panic("Query failure", err)
	}

	for rows.Next() {
		err = rows.Scan(
			&company.CompanyName,
			&company.CompanyNumber,
			&company.AddressLine1,
			&company.AddressLine2,
			&company.PostTown,
			&company.PostCode,
			&company.CompanyStatus,
			&company.IncorporationDate,
		)
		if err != nil {
			log.Panic("Error scanning db row: ", err)
		}

		companies = append(companies, company)
	}

	c.JSON(http.StatusOK, companies)
}
