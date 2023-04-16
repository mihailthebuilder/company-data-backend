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
	r := RegistrationController{
		Context:       c,
		RouterConfing: c.MustGet("config").(*RouterConfig),
	}

	err := r.saveRegistration()
	if err != nil {
		log.Println("Registration error: ", err)
		r.Context.AbortWithStatus(400)
		return
	}

	token, err := r.generateJwtToken()
	if err != nil {
		log.Println("Token generation error: ", err)
		r.Context.AbortWithStatus(500)
		return
	}

	r.Context.JSON(http.StatusOK, gin.H{"token": token})
}

type RegistrationController struct {
	Context       *gin.Context
	RouterConfing *RouterConfig
}

func (r *RegistrationController) saveRegistration() error {
	var body RegistrationRequestBody
	if err := r.Context.ShouldBindJSON(&body); err != nil {
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

	return r.RouterConfing.Emailer.SendEmail(&details)
}

type RegistrationRequestBody struct {
	EmailAddress         string
	ReasonForWantingData string
	ProblemBeingSolved   string
}

func (r *RegistrationController) generateJwtToken() (*string, error) {
	tokenLifespanInMinutes, err := strconv.Atoi(r.RouterConfing.JwtTokenLifespanInMinutes)

	if err != nil {
		return nil, fmt.Errorf("failed to parse token lifespan: %s", err)
	}

	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(tokenLifespanInMinutes))),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	str, err := token.SignedString([]byte(r.RouterConfing.ApiSecret))

	return &str, err
}
