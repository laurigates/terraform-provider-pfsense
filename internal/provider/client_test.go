package provider

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestClientDoUnwrapsEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-API-Key"); got != "test-key" {
			t.Errorf("X-API-Key = %q, want %q", got, "test-key")
		}
		if r.URL.Path != "/api/v2/firewall/alias" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("id"); got != "3" {
			t.Errorf("id query = %q, want 3", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"code": 200,
			"status": "ok",
			"response_id": "SUCCESS",
			"message": "",
			"data": {"id": 3, "name": "web_ports", "type": "port", "descr": "", "address": ["80", "443"], "detail": ["http", "https"]}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", false)
	data, err := client.Do(context.Background(), http.MethodGet, "/api/v2/firewall/alias",
		url.Values{"id": []string{"3"}}, nil)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}

	var alias firewallAliasAPI
	if err := json.Unmarshal(data, &alias); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if alias.ID == nil || *alias.ID != 3 {
		t.Errorf("id = %v, want 3", alias.ID)
	}
	if alias.Name != "web_ports" || alias.Type != "port" {
		t.Errorf("name/type = %q/%q", alias.Name, alias.Type)
	}
	if len(alias.Address) != 2 || alias.Address[0] != "80" {
		t.Errorf("address = %v", alias.Address)
	}
}

func TestClientDoSurfacesEnvelopeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": 404,
			"status": "not found",
			"response_id": "MODEL_OBJECT_NOT_FOUND",
			"message": "Object with ID 99 was not found.",
			"data": []
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", false)
	_, err := client.Do(context.Background(), http.MethodGet, "/api/v2/firewall/alias",
		url.Values{"id": []string{"99"}}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Code != 404 {
		t.Errorf("code = %d, want 404", apiErr.Code)
	}
	if apiErr.ResponseID != "MODEL_OBJECT_NOT_FOUND" {
		t.Errorf("response_id = %q", apiErr.ResponseID)
	}
}

func TestClientDoSendsJSONBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q", got)
		}
		var body firewallAliasAPI
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if body.Name != "test_alias" {
			t.Errorf("body name = %q", body.Name)
		}
		if body.ID != nil {
			t.Errorf("create body must omit id, got %v", *body.ID)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code": 200, "status": "ok", "response_id": "SUCCESS", "message": "", "data": {"id": 8, "name": "test_alias", "type": "host", "descr": "", "address": [], "detail": []}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", false)
	data, err := client.Do(context.Background(), http.MethodPost, "/api/v2/firewall/alias", nil,
		&firewallAliasAPI{Name: "test_alias", Type: "host", Address: []string{}, Detail: []string{}})
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	var alias firewallAliasAPI
	if err := json.Unmarshal(data, &alias); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if alias.ID == nil || *alias.ID != 8 {
		t.Errorf("returned id = %v, want 8", alias.ID)
	}
}
