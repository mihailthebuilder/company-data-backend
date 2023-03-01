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
	"gorm.io/gorm/schema"
)

var dbConn *gorm.DB

type SicCompany struct {
	Index          string `gorm:"index"`
	CompanyNumber  string `gorm:"CompanyNumber"`
	SicCode        string `gorm:"SicCode"`
	SicDescription string `gorm:"SicDescription"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	connectToDatabase()

	r := gin.Default()

	r.GET("/companies/sic_code/:sic_code", handleCompaniesBySicCodeRequest)

	r.Run()
}

func handleCompaniesBySicCodeRequest(c *gin.Context) {
	sic := c.Param("sic_code")

	valid := isValidSicFormat(sic)
	if !valid {
		c.JSON(http.StatusBadRequest, fmt.Sprintf("Invalid SIC code: %s", sic))
	}

	processCompaniesBySicCodeRequest(sic, c)
}

func isValidSicFormat(sic string) bool {
	pattern := "^[0-9]+$"
	match, err := regexp.MatchString(pattern, sic)

	if err != nil {
		log.Panic("Error validating SIC code:", err)
	}

	return match
}

func connectToDatabase() {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", getEnv("DB_HOST"), getEnv("DB_PORT"), getEnv("DB_USER"), getEnv("DB_PASSWORD"), getEnv("DB_NAME"))

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
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

func processCompaniesBySicCodeRequest(sic string, c *gin.Context) {
	var sicCompany SicCompany

	dbConn.First(&sicCompany, 1)

	c.JSON(200, gin.H{
		"message": sicCompany,
	})
}
