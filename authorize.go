package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func verifyAuthorization(c *gin.Context) {
	token, err := extractBearerToken(c)
	if err != nil {
		log.Panic("error extracting token: ", err)
	}

	valid, err := isValidTokenString(token)
	if err != nil {
		log.Panic("error validating token: ", err)
	}

	if !valid {
		c.AbortWithStatus(401)
	}
}

func extractBearerToken(c *gin.Context) (*string, error) {
	header := c.Request.Header.Get("Authorization")
	split := strings.Split(header, " ")

	if len(split) != 2 {
		return nil, fmt.Errorf("no authorization token: %s", split)
	}

	if split[0] != "Bearer" {
		return nil, fmt.Errorf("no bearer token given: %s", split)
	}

	return &split[1], nil
}

func isValidTokenString(tokenString *string) (bool, error) {
	token, err := getToken(tokenString)
	if err != nil {
		return false, fmt.Errorf("unable to get token from string: %s", *tokenString)
	}

	if !token.Valid {
		log.Println("parsing returns invalid token")
		return false, nil
	}

	valid := claimsAreValid(token)

	return valid, nil
}

func getToken(tokenString *string) (*jwt.Token, error) {
	return jwt.Parse(*tokenString, func(token *jwt.Token) (interface{}, error) {
		if validMethod := validateSigningMethod(token); !validMethod {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Header["alg"])
		}

		return []byte(getEnv("API_SECRET")), nil
	})
}

func validateSigningMethod(token *jwt.Token) bool {
	_, ok := token.Method.(*jwt.SigningMethodHMAC)
	return ok
}

func claimsAreValid(token *jwt.Token) bool {
	claims, ok := token.Claims.(jwt.RegisteredClaims)

	if !ok {
		log.Println("token claim type is invalid")
		return false
	}

	expired := claims.ExpiresAt.Before(time.Now())
	if expired {
		log.Println("token expired")
	}

	return !expired
}
