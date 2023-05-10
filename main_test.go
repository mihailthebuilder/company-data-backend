package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	registrationRequest := RegistrationRequestBody{
		EmailAddress:         "hello@world.com",
		ReasonForWantingData: "power",
		ProblemBeingSolved:   "more power",
	}

	e := MockEmailAPI{}

	emailRequest, _ := json.Marshal(EmailDetails{
		EmailAddress: registrationRequest.EmailAddress,
		Title:        "Company Data - Registration request",
		Message:      fmt.Sprintf("Reason for wanting data: %s . Problem being solved: %s", registrationRequest.ReasonForWantingData, registrationRequest.ProblemBeingSolved),
	})

	e.On("SendRequest", bytes.NewReader(emailRequest)).Return(&http.Response{StatusCode: 200}, nil)

	var handler = RouteHandler{
		EmailAPI:                  e,
		JwtTokenLifespanInMinutes: "60",
		ApiSecret:                 "helloWorld",
	}
	r := createRouter(&handler)

	w := httptest.NewRecorder()

	requestBody, _ := json.Marshal(registrationRequest)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(requestBody))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responseBody map[string]string
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(t, true, len(responseBody["token"]) > 0)
}

type MockEmailAPI struct {
	mock.Mock
}

func (e MockEmailAPI) SendRequest(body io.Reader) (*http.Response, error) {
	args := e.Called(body)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestSampleRoute_ShouldReturnInvalidRequestWhenNoBody(t *testing.T) {
	var handler = RouteHandler{}
	r := createRouter(&handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/companies/sample", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSampleRoute_ShouldReturnInvalidRequestWhenInvalidJsonBody(t *testing.T) {
	var handler = RouteHandler{}
	r := createRouter(&handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/companies/sample", bytes.NewReader([]byte("hello world")))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSampleRoute_ShouldReturnData(t *testing.T) {
	d := MockDatabase{}

	industry := "Extraction of salt"
	d.On("GetListOfCompanies", &industry, true).Return([]ProcessedCompany{{}}, nil)

	handler := RouteHandler{
		Database: d,
	}

	r := createRouter(&handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/companies/sample", bytes.NewReader([]byte(`{"SicDescription":"Extraction of salt"}`)))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

type MockDatabase struct {
	mock.Mock
}

func (e MockDatabase) GetListOfCompanies(i *string, s bool) ([]ProcessedCompany, error) {
	args := e.Called(i, s)
	return args.Get(0).([]ProcessedCompany), args.Error(1)
}

func TestFullRoute_ShouldReturnUnauthorizedWhenNoJwtToken(t *testing.T) {
	handler := RouteHandler{}
	r := createRouter(&handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/companies/authorized/full", bytes.NewReader([]byte(`{"SicDescription":"Extraction of salt"}`)))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthorizationRouteToFullRouteFlow_HappyPath(t *testing.T) {
	d := MockDatabase{}
	industry := "Extraction of salt"
	d.On("GetListOfCompanies", &industry, false).Return([]ProcessedCompany{{}}, nil)

	registrationRequest := RegistrationRequestBody{
		EmailAddress:         "hello@world.com",
		ReasonForWantingData: "power",
		ProblemBeingSolved:   "more power",
	}

	e := MockEmailAPI{}

	emailRequest, _ := json.Marshal(EmailDetails{
		EmailAddress: registrationRequest.EmailAddress,
		Title:        "Company Data - Registration request",
		Message:      fmt.Sprintf("Reason for wanting data: %s . Problem being solved: %s", registrationRequest.ReasonForWantingData, registrationRequest.ProblemBeingSolved),
	})

	e.On("SendRequest", bytes.NewReader(emailRequest)).Return(&http.Response{StatusCode: 200}, nil)

	var handler = RouteHandler{
		EmailAPI:                  e,
		JwtTokenLifespanInMinutes: "60",
		ApiSecret:                 "helloWorld",
		Database:                  d,
	}

	r := createRouter(&handler)

	w := httptest.NewRecorder()

	requestBody, _ := json.Marshal(registrationRequest)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(requestBody))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var responseBody map[string]string
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/companies/authorized/full", bytes.NewReader([]byte(`{"SicDescription":"Extraction of salt"}`)))
	req.Header = map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", responseBody["token"])},
	}
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
