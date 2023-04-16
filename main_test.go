package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestRegisterRoute_ShouldReturn500WhenFormDataNotGiven(t *testing.T) {
	r := createRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/register", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestRegisterRoute_ShouldReturn500WhenPartialFormDataGiven(t *testing.T) {
	r := createRouter()

	w := httptest.NewRecorder()

	requestBody, _ := json.Marshal(RegistrationRequestBody{EmailAddress: "hello@world.com"})

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(requestBody))
	r.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestRegisterRoute_ShouldReturnJwtTokenWhenFullFormDataGiven(t *testing.T) {
	r := createRouter()

	w := httptest.NewRecorder()

	requestBody, _ := json.Marshal(
		RegistrationRequestBody{
			EmailAddress:         "hello@world.com",
			ReasonForWantingData: "power",
			ProblemBeingSolved:   "more power",
		},
	)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(requestBody))
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
