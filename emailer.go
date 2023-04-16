package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type IEmailer interface {
	SendEmail(details EmailDetails) error
}

type EmailDetails struct {
	EmailAddress string
	Title        string
	Message      string
}

type Emailer struct {
	EmailApiUrl string
}

func (e *Emailer) SendEmail(details EmailDetails) error {
	requestBody, err := json.Marshal(details)
	if err != nil {
		return fmt.Errorf("failed marshalling request: %s", err)
	}

	response, err := http.Post(e.EmailApiUrl, "application/json", bytes.NewBuffer(requestBody))

	if err != nil {
		return fmt.Errorf("failed sending email request: %s", err)
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response from email request. status code %d ; status %s", response.StatusCode, response.Status)
	}

	return nil
}
