package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// HTTPClient defines an interface for HTTP clients
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	// Client is the default HTTP client
	Client HTTPClient = &http.Client{
		Timeout: 10 * time.Second,
	}
)

// HTTPError represents an error response from an HTTP request
type HTTPError struct {
	Status     int
	StatusText string
	Message    string
}

// Error implements the error interface
func (e *HTTPError) Error() string {
	return fmt.Sprintf("%d %s: %s", e.Status, e.StatusText, e.Message)
}

// Get makes a GET request to the specified URL
func Get(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add headers
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// Execute request
	res, err := Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer res.Body.Close()

	// Read response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Check status code
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, &HTTPError{
			Status:     res.StatusCode,
			StatusText: res.Status,
			Message:    string(body),
		}
	}

	return body, nil
}

// Post makes a POST request to the specified URL with the given body
func Post(ctx context.Context, url string, headers map[string]string, body interface{}) ([]byte, error) {
	// Convert body to JSON
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, io.NopCloser(io.Reader(io.New(jsonBody))))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add headers
	req.Header.Add("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// Execute request
	res, err := Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer res.Body.Close()

	// Read response body
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Check status code
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, &HTTPError{
			Status:     res.StatusCode,
			StatusText: res.Status,
			Message:    string(resBody),
		}
	}

	return resBody, nil
}

// GetJSON makes a GET request to the specified URL and unmarshals the response into the result
func GetJSON(ctx context.Context, url string, headers map[string]string, result interface{}) error {
	body, err := Get(ctx, url, headers)
	if err != nil {
		return err
	}

	// Unmarshal response
	if err := json.Unmarshal(body, result); err != nil {
		log.Error().Err(err).Str("url", url).Str("response", string(body)).Msg("Error unmarshaling response")
		return fmt.Errorf("error unmarshaling response: %w", err)
	}

	return nil
}

// PostJSON makes a POST request to the specified URL and unmarshals the response into the result
func PostJSON(ctx context.Context, url string, headers map[string]string, body, result interface{}) error {
	resBody, err := Post(ctx, url, headers, body)
	if err != nil {
		return err
	}

	// Unmarshal response
	if err := json.Unmarshal(resBody, result); err != nil {
		log.Error().Err(err).Str("url", url).Str("response", string(resBody)).Msg("Error unmarshaling response")
		return fmt.Errorf("error unmarshaling response: %w", err)
	}

	return nil
}

// VerifyTokenWithAuthService verifies a token with the Auth Service
func VerifyTokenWithAuthService(ctx context.Context, token string, authServiceURL string) (map[string]interface{}, error) {
	// Create URL
	url := fmt.Sprintf("%s/verify", authServiceURL)

	// Set headers
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}

	// Make request
	var result struct {
		Status  string                 `json:"status"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data"`
	}

	if err := GetJSON(ctx, url, headers, &result); err != nil {
		return nil, fmt.Errorf("error verifying token: %w", err)
	}

	// Check status
	if result.Status != "success" {
		return nil, fmt.Errorf("token verification failed: %s", result.Message)
	}

	return result.Data, nil
}
