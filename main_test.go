package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/mock"
)

type MockEmailer struct {
	mock.Mock
}

func (e MockEmailer) SendEmail(d *EmailDetails) error {
	args := e.Called(d)
	return args.Error(0)
}

func TestRegisterRoute_ShouldReturn500WhenFormDataNotGiven(t *testing.T) {
	var config = RouterConfig{}
	r := createRouter(&config)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/register", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestRegisterRoute_ShouldReturn500WhenPartialFormDataGiven(t *testing.T) {
	var config = RouterConfig{}
	r := createRouter(&config)

	w := httptest.NewRecorder()

	requestBody, _ := json.Marshal(RegistrationRequestBody{EmailAddress: "hello@world.com"})

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(requestBody))
	r.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestRegisterRoute_ShouldReturnJwtTokenWhenFullFormDataGiven(t *testing.T) {
	body := RegistrationRequestBody{
		EmailAddress:         "hello@world.com",
		ReasonForWantingData: "power",
		ProblemBeingSolved:   "more power",
	}

	e := MockEmailer{}
	e.On("SendEmail", &EmailDetails{EmailAddress: "hello@world.com", Title: "Company Data - Registration request", Message: fmt.Sprintf("Reason for wanting data: %s . Problem being solved: %s", body.ReasonForWantingData, body.ProblemBeingSolved)})

	var config = RouterConfig{
		Emailer: e,
	}
	r := createRouter(&config)

	w := httptest.NewRecorder()

	requestBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(requestBody))
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
