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
