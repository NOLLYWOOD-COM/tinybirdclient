package tinybird

import (
	"os"
	"time"

	"github.com/NOLLYWOOD-COM/tinybird/internal/httpclient"
)

const VERSION = "0.2.0"

// ApiVersion sets the Tinybird API version in ClientOptions.
func ApiVersion(version string) Option {
	return func(co *ClientOptions) {
		co.ApiVersion = version
	}
}

// Protocol sets the Tinybird protocol in ClientOptions.
func Protocol(protocol string) Option {
	return func(co *ClientOptions) {
		co.Protocol = protocol
	}
}

// Timeout sets the timeout duration in ClientOptions.
func Timeout(timeout time.Duration) Option {
	return func(co *ClientOptions) {
		co.Timeout = timeout
	}
}

// Host sets the Tinybird host URL in ClientOptions.
func Host(host string) Option {
	return func(co *ClientOptions) {
		co.Host = host
	}
}

// Token sets the Tinybird API token in ClientOptions.
func Token(token string) Option {
	return func(co *ClientOptions) {
		co.Token = token
	}
}

// MaxRetries sets the maximum number of retries in ClientOptions.
func MaxRetries(retries int) Option {
	return func(co *ClientOptions) {
		co.MaxRetries = retries
	}
}

// RetryDelay sets the delay between retries in ClientOptions.
func RetryDelay(delay time.Duration) Option {
	return func(co *ClientOptions) {
		co.RetryDelay = delay
	}
}

// NewClientOptions creates a new ClientOptions instance with the provided options.
func NewClientOptions(options ...Option) *ClientOptions {
	co := &ClientOptions{}

	for _, option := range options {
		option(co)
	}

	if co.Timeout == 0 {
		co.Timeout = 15 * time.Second
	}

	if co.Host == "" {
		co.Host = "https://api.tinybird.co"
	}

	if co.Token == "" {
		co.Token = os.Getenv("TINYBIRD_TOKEN")
	}

	if co.ApiVersion == "" {
		co.ApiVersion = "v0"
	}

	if co.Protocol == "" {
		co.Protocol = "https"
	}

	if co.MaxRetries == 0 {
		co.MaxRetries = 3
	}

	if co.RetryDelay == 0 {
		co.RetryDelay = 2 * time.Second
	}

	return co
}

func DefaultClientOptions() *ClientOptions {
	return &ClientOptions{
		Host:       "https://api.tinybird.co",
		Token:      os.Getenv("TINYBIRD_TOKEN"),
		Timeout:    15 * time.Second,
		ApiVersion: "v0",
		Protocol:   "https",
		MaxRetries: 3,
		RetryDelay: 2 * time.Second,
	}
}

// NewClient creates a new TinybirdClient with the given ClientOptions.
//
// options: Configuration options for the Tinybird client.
//
// http:    Optional custom HTTP client. If nil, a default client will be created.
func NewClient(options *ClientOptions, http httpclient.Client) Client {
	if http == nil {
		http = httpclient.New(&httpclient.Config{
			Timeout:    options.Timeout,
			RetryDelay: options.RetryDelay,
			MaxRetries: options.MaxRetries,
			UserAgent:  "com.nollywood/tinybirdclient/" + VERSION,
			Token:      options.Token,
		})
	}

	return &ClientImpl{
		httpClient: http,
		options:    options,
	}
}
