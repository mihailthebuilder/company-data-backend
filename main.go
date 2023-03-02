package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	loadEnvironmentVariables()

	connectToDatabase()

	runApplication()
}

func loadEnvironmentVariables() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

var dbConn *gorm.DB

func connectToDatabase() {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", getEnv("DB_HOST"), getEnv("DB_PORT"), getEnv("DB_USER"), getEnv("DB_PASSWORD"), getEnv("DB_NAME"))

	db, err := gorm.Open(postgres.Open(connStr))
	if err != nil {
		log.Panic("Couldn't connect to database:", err)
	}

	dbConn = db
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

	r.GET("/companies/sic_code/:sic_code", handleCompaniesBySicCodeRequest)

	r.Run()
}

type SicCompany struct {
	Index          string `gorm:"column:index"`
	CompanyNumber  string `gorm:"column:CompanyNumber"`
	SicCode        string `gorm:"column:SicCode"`
	SicDescription string `gorm:"column:SicDescription"`
}

func (SicCompany) TableName() string {
	return "sic_company"
}

func handleCompaniesBySicCodeRequest(c *gin.Context) {
	sic := c.Param("sic_code")

	valid := isValidSicFormat(&sic)
	if !valid {
		c.JSON(http.StatusBadRequest, fmt.Sprintf("Invalid SIC code: %s", sic))
		return
	}

	processCompaniesBySicCodeRequest(&sic, c)
}

func isValidSicFormat(sic *string) bool {
	pattern := "^[0-9]+$"
	match, err := regexp.MatchString(pattern, *sic)

	if err != nil {
		log.Panic("Error validating SIC code:", err)
	}

	return match
}

func processCompaniesBySicCodeRequest(sic *string, c *gin.Context) {
	var sicCompanies []SicCompany

	result := dbConn.Model(&SicCompany{}).Where(&SicCompany{SicCode: *sic}).Order("RANDOM()").Limit(10).Find(&sicCompanies)
	if result.Error != nil {
		log.Panic("Query failure", result.Error)
	}

	c.JSON(http.StatusOK, sicCompanies)
}
