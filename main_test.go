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

func TestRegisterRoute_ShouldReturnInvalidRequestWhenFormDataNotGiven(t *testing.T) {
	var handler = RouteHandler{}
	r := createRouter(&handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/register", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterRoute_ShouldReturnInvalidRequestWhenPartialFormDataGiven(t *testing.T) {
	var handler = RouteHandler{}
	r := createRouter(&handler)

	w := httptest.NewRecorder()

	requestBody, _ := json.Marshal(RegistrationRequestBody{EmailAddress: "hello@world.com"})

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(requestBody))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterRoute_ShouldReturnJwtTokenWhenFullFormDataGiven(t *testing.T) {
	body := RegistrationRequestBody{
		EmailAddress:         "hello@world.com",
		ReasonForWantingData: "power",
		ProblemBeingSolved:   "more power",
	}

	e := MockEmailer{}
	e.On("SendEmail", &EmailDetails{EmailAddress: "hello@world.com", Title: "Company Data - Registration request", Message: fmt.Sprintf("Reason for wanting data: %s . Problem being solved: %s", body.ReasonForWantingData, body.ProblemBeingSolved)}).Return(nil)

	var handler = RouteHandler{
		Emailer:                   e,
		JwtTokenLifespanInMinutes: "60",
		ApiSecret:                 "helloWorld",
	}
	r := createRouter(&handler)

	w := httptest.NewRecorder()

	requestBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(requestBody))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

type MockEmailer struct {
	mock.Mock
}

func (e MockEmailer) SendEmail(d *EmailDetails) error {
	args := e.Called(d)
	return args.Error(0)
}

func TestSampleRoute_ShouldReturnInvalidRequestWhenNoBody(t *testing.T) {
	var handler = RouteHandler{}
	r := createRouter(&handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/companies/sample", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestSampleRoute_ShouldReturnInvalidRequestWhenInvalidJsonBody(t *testing.T) {
	var handler = RouteHandler{}
	r := createRouter(&handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/companies/sample", bytes.NewReader([]byte("hello world")))
	r.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestSampleRoute_ShouldReturnData(t *testing.T) {
	d := MockDatabase{}

	industry := "Extraction of salt"
	d.On("GetSampleListOfCompaniesForIndustry", &industry).Return(&[]ProcessedCompany{{}})

	handler := RouteHandler{
		Database: d,
	}

	r := createRouter(&handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/companies/sample", bytes.NewReader([]byte(`{"SicDescription":"Extraction of salt"}`)))
	r.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

type MockDatabase struct {
	mock.Mock
}

func (e MockDatabase) GetSampleListOfCompaniesForIndustry(i *string) (*[]ProcessedCompany, error) {
	args := e.Called(i)
	return args.Get(0).(*[]ProcessedCompany), args.Error(1)
}
