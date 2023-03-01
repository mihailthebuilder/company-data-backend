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

	valid, err := isValidSicFormat(sic)
	if err != nil {
		log.Println("Valid SIC format check failed:", err)
		c.JSON(http.StatusInternalServerError, "Internal server error")
	}

	if !valid {
		c.JSON(http.StatusBadRequest, fmt.Sprintf("Invalid SIC code: %s", sic))
	}

	processCompaniesBySicCodeRequest(sic, c)
}

func isValidSicFormat(sic string) (bool, error) {
	pattern := "^[0-9]+$"
	match, err := regexp.MatchString(pattern, sic)
	return match, err
}

func connectToDatabase() {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", GetEnv("DB_HOST"), GetEnv("DB_PORT"), GetEnv("DB_USER"), GetEnv("DB_PASSWORD"), GetEnv("DB_NAME"))

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		log.Panicf("Couldn't connect to database: %s", err)
	}

	dbConn = db
}

func GetEnv(env string) string {
	val := os.Getenv(env)
	if val == "" {
		log.Fatalf("Environment variable not set: %s", env)
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
