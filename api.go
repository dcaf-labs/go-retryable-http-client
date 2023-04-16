package api

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	defaultCallsPerSecond = 1.0
	defaultRetryDuration  = time.Second * 10
	defaultMaxRetries     = 3
)

type RetryableHTTPClient struct {
	*retryablehttp.Client
}

type RateLimitHTTPClientOptions struct {
	CallsPerSecond   float64
	HttpClient       *http.Client
	MaxRetries       int
	MinRetryDuration time.Duration
	CheckRetry       retryablehttp.CheckRetry
}

type RetryableHTTPClientProvider func(options RateLimitHTTPClientOptions) RetryableHTTPClient

func (r RetryableHTTPClient) Do(request *http.Request) (*http.Response, error) {
	retryableRequest, err := retryablehttp.FromRequest(request)
	if err != nil {
		return nil, err
	}
	return r.Client.Do(retryableRequest)
}

func (r RetryableHTTPClient) CloseIdleConnections() {
	r.HTTPClient.CloseIdleConnections()
}
