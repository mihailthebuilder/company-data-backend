package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/gin-gonic/gin"
)

var dbConn *sql.DB

func main() {
	err := connectToDatabase()
	if err != nil {
		log.Panicf("Couldn't connect to database: %s", err)
	}

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

	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func isValidSicFormat(sic string) (bool, error) {
	pattern := "^[0-9]+$"
	match, err := regexp.MatchString(pattern, sic)
	return match, err
}

func connectToDatabase() error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", GetEnv("DB_HOST"), GetEnv("DB_PORT"), GetEnv("DB_USER"), GetEnv("DB_PASSWORD"), GetEnv("DB_NAME"))
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	*dbConn = *db
	return nil
}

func GetEnv(env string) string {
	val := os.Getenv(env)
	if val == "" {
		log.Fatalf("Environment variable not set: %s", env)
	}
	return val
}
