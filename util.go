package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/time/rate"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
)

func DecodeRequestBody[V any](resp *http.Response, res V) (V, error) {
	var err error
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			if err != nil {
				err = fmt.Errorf("err %s, failed to close request body %s", err.Error(), closeErr.Error())
			} else {
				err = closeErr
			}
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return res, err
	}
	return res, err
}

func GetDefaultRateLimitHTTPClientOptions() RateLimitHTTPClientOptions {
	return RateLimitHTTPClientOptions{
		CallsPerSecond:   defaultCallsPerSecond,
		HttpClient:       *retryablehttp.NewClient().HTTPClient,
		MaxRetries:       defaultMaxRetries,
		MinRetryDuration: defaultRetryDuration,
		CheckRetry:       GetDefaultCheckRetry,
	}
}

func GetDefaultClientProvider() RetryableHTTPClientProvider {
	return func(options RateLimitHTTPClientOptions) RetryableHTTPClient {
		return GetDefaultRateLimitedHTTPClient(options)
	}
}
func GetDefaultRateLimitedHTTPClient(options RateLimitHTTPClientOptions) RetryableHTTPClient {
	rateLimiter := rate.NewLimiter(rate.Every(time.Minute/time.Duration(options.CallsPerSecond*60)), 1)

	client := retryablehttp.NewClient()
	client.HTTPClient = &options.HttpClient
	client.Logger = nil
	client.CheckRetry = options.CheckRetry
	client.RetryWaitMin = options.MinRetryDuration
	client.RetryMax = options.MaxRetries
	client.RequestLogHook = func(logger retryablehttp.Logger, req *http.Request, retry int) {
		if err := rateLimiter.Wait(context.Background()); err != nil {
			logger.Printf("failed to wait for rate limiter %w", err)
			return
		}
	}
	return RetryableHTTPClient{client}
}

func GetDefaultCheckRetry(_ context.Context, resp *http.Response, err error) (bool, error) {
	log := logrus.NewEntry(logrus.New())
	if err != nil {
		log = log.WithError(err)
	}
	if resp != nil && resp.Request != nil && resp.Request.URL != nil {
		log = log.WithField("url", resp.Request.URL.String())
	}
	if resp != nil {
		log = log.WithField("statusCode", resp.StatusCode)
	}
	if resp != nil && resp.StatusCode >= http.StatusTooManyRequests {
		log.Error("waiting for rate limit")
		return true, err
	}
	return false, nil
}
