package integrations

import (
	"io"
	"net/http"
)

type EmailAPI struct {
	URL string
}

func (e *EmailAPI) SendRequest(body io.Reader) (*http.Response, error) {
	return http.Post(e.URL, "application/json", body)
}
