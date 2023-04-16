package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func (h *RouteHandler) Register(c *gin.Context) {
	err := h.saveRegistration(c)
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

func (h *RouteHandler) saveRegistration(c *gin.Context) error {
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

	return h.sendEmail(&details)
}

type EmailDetails struct {
	EmailAddress string
	Title        string
	Message      string
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

func (h *RouteHandler) sendEmail(details *EmailDetails) error {
	requestBody, err := json.Marshal(*details)
	if err != nil {
		return fmt.Errorf("failed marshalling request: %s", err)
	}

	response, err := h.EmailAPI.SendRequest(bytes.NewReader(requestBody))

	if err != nil {
		return fmt.Errorf("failed sending email request: %s", err)
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response from email request. status code %d ; status %s", response.StatusCode, response.Status)
	}

	return nil
}
