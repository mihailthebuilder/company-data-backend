package main

import (
	"io"
	"net/http"
)

type IEmailAPI interface {
	SendRequest(body io.Reader) (*http.Response, error)
}

type EmailAPI struct {
	URL string
}

func (e *EmailAPI) SendRequest(body io.Reader) (*http.Response, error) {
	return http.Post(e.URL, "application/json", body)
}
