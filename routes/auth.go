package routes

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func (h *RouteHandler) VerifyAuthorization(c *gin.Context) {
	token, err := h.extractBearerTokenFromHeader(c.Request.Header.Get("Authorization"))
	if err != nil {
		log.Println("error extracting token: ", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	valid, err := h.isValidTokenString(token)
	if err != nil {
		log.Println("error validating token: ", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if !valid {
		log.Println("invalid token")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
}

func (h *RouteHandler) extractBearerTokenFromHeader(header string) (*string, error) {
	split := strings.Split(header, " ")

	if len(split) != 2 {
		return nil, fmt.Errorf("no authorization token: %s", split)
	}

	if split[0] != "Bearer" {
		return nil, fmt.Errorf("no bearer token given: %s", split)
	}

	return &split[1], nil
}

func (h *RouteHandler) isValidTokenString(tokenString *string) (bool, error) {
	token, err := h.getToken(tokenString)
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

func (h *RouteHandler) getToken(tokenString *string) (*jwt.Token, error) {
	return jwt.Parse(*tokenString, func(token *jwt.Token) (interface{}, error) {
		if validMethod := validateSigningMethod(token); !validMethod {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Header["alg"])
		}

		return []byte(h.ApiSecret), nil
	})
}

func validateSigningMethod(token *jwt.Token) bool {
	_, ok := token.Method.(*jwt.SigningMethodHMAC)
	return ok
}

func claimsAreValid(token *jwt.Token) bool {
	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		log.Println("token claim type is invalid: ", claims)
		return false
	}

	expirationTime, err := claims.GetExpirationTime()
	if err != nil {
		log.Println("unable to get expiration time from token claim: ", claims)
	}

	expired := expirationTime.Before(time.Now())

	if expired {
		log.Println("token expired: ", claims)
	}

	return !expired
}
