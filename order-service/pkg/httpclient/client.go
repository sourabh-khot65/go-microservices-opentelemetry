package httpclient

import "net/http"

func New() *http.Client {
	return &http.Client{}
}
