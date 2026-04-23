package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const DefaultBaseURL = "https://api.qwilr.com"

type Client struct {
	token   string
	baseURL string
	http    *http.Client
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	switch e.StatusCode {
	case 401:
		return "authentication failed — run \"qwilr configure\" to set your API token"
	case 404:
		return fmt.Sprintf("resource not found (HTTP %d)", e.StatusCode)
	case 429:
		return "rate limited — wait a moment and try again"
	default:
		if e.StatusCode >= 500 {
			return fmt.Sprintf("Qwilr API error (HTTP %d) — try again or check https://status.qwilr.com", e.StatusCode)
		}
		return fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Message)
	}
}

func NewClient(token, baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		token:   token,
		baseURL: baseURL,
		http:    &http.Client{},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.baseURL + "/v1" + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}
	return resp, nil
}

func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) Post(ctx context.Context, path string, body, result interface{}) error {
	resp, err := c.do(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) Put(ctx context.Context, path string, body, result interface{}) error {
	resp, err := c.do(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) Delete(ctx context.Context, path string) error {
	resp, err := c.do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
