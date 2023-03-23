package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func handleRequestToSendEmail(c *gin.Context) {

	plainTextContent := "and easy to do anywhere, even with Go"
	htmlContent := "<strong>and easy to do anywhere, even with Go</strong>"

	from := mail.NewEmail("Sender", getEnv("EMAIL_SENDER"))
	subject := "Company Database - Email"
	to := mail.NewEmail("Recipient", getEnv("EMAIL_RECIPIENT"))
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(getEnv("EMAIL_API_KEY"))

	response, err := client.Send(message)
	if err != nil {
		log.Panic("error sending email: ", err)
	}

	if response.StatusCode != 202 {
		log.Panic("non-200 status code when sending email: ", response.StatusCode, "; status text: ", response.Body)
	}

	c.Status(http.StatusOK)
}
