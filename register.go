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

func (h *RouteHandler) Register(c *gin.Context) {
	err := saveRegistration(c, h.Emailer)
	if err != nil {
		log.Println("Registration error: ", err)
		c.AbortWithStatus(400)
		return
	}

	token, err := generateJwtToken(&h.ApiSecret, &h.JwtTokenLifespanInMinutes)
	if err != nil {
		log.Println("Token generation error: ", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

type RegistrationController struct {
	Context       *gin.Context
	RouterConfing *RouteHandler
}

func saveRegistration(c *gin.Context, e IEmailer) error {
	var body RegistrationRequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return fmt.Errorf("failed parsing request body: %s", err)
	}

	if len(body.EmailAddress) == 0 || len(body.ReasonForWantingData) == 0 || len(body.ProblemBeingSolved) == 0 {
		return fmt.Errorf("request body doesn't have all required attributes: %s", body)
	}

	details := EmailDetails{
		EmailAddress: body.EmailAddress,
		Title:        "Company Data - Registration request",
		Message:      fmt.Sprintf("Reason for wanting data: %s . Problem being solved: %s", body.ReasonForWantingData, body.ProblemBeingSolved),
	}

	return e.SendEmail(&details)
}

type RegistrationRequestBody struct {
	EmailAddress         string
	ReasonForWantingData string
	ProblemBeingSolved   string
}

func generateJwtToken(secret *string, lifespanInMinutes *string) (*string, error) {
	tokenLifespanInMinutes, err := strconv.Atoi(*lifespanInMinutes)

	if err != nil {
		return nil, fmt.Errorf("failed to parse token lifespan: %s", err)
	}

	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(tokenLifespanInMinutes))),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	str, err := token.SignedString([]byte(*secret))

	return &str, err
}
