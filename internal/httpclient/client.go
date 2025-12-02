package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

// New creates a new HTTP client with the given configuration
func New(config *Config) Client {
	return &client{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		config: config,
	}
}

func (c *client) Get(ctx context.Context, urlStr string, params map[string]string, result interface{}) error {
	return c.makeRequest(ctx, http.MethodGet, urlStr, params, result)
}

func (c *client) Post(ctx context.Context, urlStr string, body interface{}, result interface{}) error {
	return c.makeRequest(ctx, http.MethodPost, urlStr, body, result)
}

func (c *client) Put(ctx context.Context, urlStr string, body interface{}, result interface{}) error {
	return c.makeRequest(ctx, http.MethodPut, urlStr, body, result)
}

func (c *client) Patch(ctx context.Context, urlStr string, body interface{}, result interface{}) error {
	return c.makeRequest(ctx, http.MethodPatch, urlStr, body, result)
}

func (c *client) Delete(ctx context.Context, urlStr string, params map[string]string, result interface{}) error {
	return c.makeRequest(ctx, http.MethodDelete, urlStr, params, result)
}

func (c *client) PostRaw(ctx context.Context, urlStr string, body []byte, contentType string, contentEncoding string, result interface{}) error {
	return c.executeRawWithRetry(ctx, http.MethodPost, urlStr, body, contentType, contentEncoding, result)
}

func (c *client) PostMultipart(ctx context.Context, urlStr string, fieldName string, fileName string, fileData []byte, result interface{}) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(fileData); err != nil {
		return fmt.Errorf("failed to write file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	return c.executeRawWithRetry(ctx, http.MethodPost, urlStr, buf.Bytes(), writer.FormDataContentType(), "", result)
}

func (c *client) makeRequest(ctx context.Context, method, urlStr string, data interface{}, result interface{}) error {
	var bodyBytes []byte
	var err error

	// Prepare URL and body based on method
	if method == http.MethodGet || method == http.MethodDelete {
		// Handle query parameters
		if data != nil {
			var queryParams string
			// Check if data is map[string]string or a struct
			if paramsMap, ok := data.(map[string]string); ok {
				// Convert map to query string
				values := url.Values{}
				for k, v := range paramsMap {
					values.Add(k, v)
				}
				queryParams = values.Encode()
			} else {
				// Use StructToQueryParams for structs
				queryParams = StructToQueryParams(data)
			}

			if queryParams != "" {
				// Properly handle existing query parameters
				parsedURL, err := url.Parse(urlStr)
				if err != nil {
					return fmt.Errorf("invalid URL: %w", err)
				}

				if parsedURL.RawQuery != "" {
					parsedURL.RawQuery += "&" + queryParams
				} else {
					parsedURL.RawQuery = queryParams
				}
				urlStr = parsedURL.String()
			}
		}
	} else if data != nil {
		// Prepare JSON body for POST, PUT, PATCH
		bodyBytes, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	// Execute request with retry logic
	return c.executeWithRetry(ctx, method, urlStr, bodyBytes, result)
}

func (c *client) executeWithRetry(ctx context.Context, method, urlStr string, bodyBytes []byte, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retrying with exponential backoff
			time.Sleep(c.config.RetryDelay * time.Duration(attempt))
		}

		// Create fresh request for each attempt
		var body io.Reader
		if len(bodyBytes) > 0 {
			body = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		if len(bodyBytes) > 0 {
			req.Header.Set("Content-Type", "application/json")
		}

		req.Header.Set("User-Agent", c.config.UserAgent)

		if c.config.Token != "" {
			req.Header.Set("Authorization", "Bearer "+c.config.Token)
		}

		// Execute request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		// Handle response
		lastErr = c.handleResponse(resp, result)

		// Close response body immediately
		resp.Body.Close()

		// Check if we should retry
		if lastErr == nil {
			return nil
		}

		// Don't retry on client errors (4xx except 429)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
			return lastErr
		}

		// Retry on server errors (5xx) and rate limiting (429)
		if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
			continue
		}

		// For other errors, don't retry
		return lastErr
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *client) executeRawWithRetry(ctx context.Context, method, urlStr string, bodyBytes []byte, contentType string, contentEncoding string, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(c.config.RetryDelay * time.Duration(attempt))
		}

		var body io.Reader
		if len(bodyBytes) > 0 {
			body = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		if contentEncoding != "" {
			req.Header.Set("Content-Encoding", contentEncoding)
		}
		if c.config.Token != "" {
			req.Header.Set("Authorization", "Bearer "+c.config.Token)
		}
		req.Header.Set("User-Agent", c.config.UserAgent)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		lastErr = c.handleResponse(resp, result)
		resp.Body.Close()

		if lastErr == nil {
			return nil
		}

		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
			return lastErr
		}

		if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
			continue
		}

		return lastErr
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *client) handleResponse(resp *http.Response, result interface{}) error {
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Success - unmarshal into result if provided and not 204 No Content
		if result != nil && resp.StatusCode != http.StatusNoContent && len(body) > 0 {
			if err := json.Unmarshal(body, result); err != nil {
				return fmt.Errorf("failed to unmarshal response: %w", err)
			}
		}
		return nil
	}

	// Error response - include body in error message
	if len(body) > 0 {
		return fmt.Errorf("HTTP %d: %s - %s", resp.StatusCode, resp.Status, string(body))
	}
	return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
}
