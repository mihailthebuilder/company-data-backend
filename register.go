package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func handleRegistration(c *gin.Context) {
	err := saveRegistration(c)
	if err != nil {
		log.Panic("Registration failed:", err)
	}

	token, err := generateJwtToken()

	if err != nil {
		log.Panic("Token generation error:", err)
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func saveRegistration(c *gin.Context) error {
	return nil
}

func generateJwtToken() (string, error) {
	tokenLifespanInMinutes, err := strconv.Atoi(getEnv("TOKEN_MINUTE_LIFESPAN"))

	if err != nil {
		return "", fmt.Errorf("failed to parse token lifespan: %s", err)
	}

	claims := jwt.MapClaims{}
	claims["exp"] = time.Now().Add(time.Minute * time.Duration(tokenLifespanInMinutes)).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(getEnv("API_SECRET")))
}