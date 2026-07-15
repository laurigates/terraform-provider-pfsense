package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client is a minimal pfSense REST API v2 client.
//
// Every API response is wrapped in a common envelope:
//
//	{"code": 200, "status": "ok", "response_id": "...", "message": "...", "data": {...}}
//
// Do unwraps the envelope and returns the raw `data` payload; non-2xx
// envelope codes are surfaced as errors carrying status/message/response_id.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// NewClient builds a client for the pfSense box at baseURL (e.g.
// "https://192.168.0.1"). When insecure is true, TLS certificate
// verification is skipped (self-signed box certificates).
func NewClient(baseURL string, apiKey string, insecure bool) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http: &http.Client{
			Transport: transport,
			// Some endpoints on this API are known to hang; never wait forever.
			Timeout: 30 * time.Second,
		},
	}
}

// apiEnvelope is the pfSense REST API v2 response wrapper.
type apiEnvelope struct {
	Code       int             `json:"code"`
	Status     string          `json:"status"`
	ResponseID string          `json:"response_id"`
	Message    string          `json:"message"`
	Data       json.RawMessage `json:"data"`
}

// APIError is a non-2xx pfSense API response.
type APIError struct {
	Code       int
	Status     string
	ResponseID string
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("pfSense API error %d (%s): %s [response_id=%s]",
		e.Code, e.Status, e.Message, e.ResponseID)
}

// Do performs an API call and returns the unwrapped `data` payload.
// path is the endpoint path (e.g. "/api/v2/firewall/alias"), query may be
// nil, and body (when non-nil) is JSON-encoded as the request body.
func (c *Client) Do(ctx context.Context, method string, path string, query url.Values, body any) (json.RawMessage, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encoding request body: %w", err)
		}
		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reader)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-API-Key", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var envelope apiEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("%s %s: HTTP %d with non-envelope body: %w",
			method, path, resp.StatusCode, err)
	}

	if envelope.Code < 200 || envelope.Code >= 300 {
		return nil, &APIError{
			Code:       envelope.Code,
			Status:     envelope.Status,
			ResponseID: envelope.ResponseID,
			Message:    envelope.Message,
		}
	}

	return envelope.Data, nil
}
