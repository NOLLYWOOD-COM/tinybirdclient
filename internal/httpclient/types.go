package httpclient

import (
	"net/http"
	"time"
)

// client is the internal HTTP client implementation
type client struct {
	httpClient *http.Client
	config     *Config
}

// Config holds configuration for the HTTP client
type Config struct {
	Timeout    time.Duration
	RetryDelay time.Duration
	MaxRetries int
	UserAgent  string
	Token      string
}
