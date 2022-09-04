package http_client

import (
	"net/http"
	"time"
)

var (
	HttpClient = &http.Client{
		Timeout: time.Second * 5,
	}
)

// Language: go
// Path: components\http_client\http-client.go
// Compare this snippet from web\app-conf.go:
