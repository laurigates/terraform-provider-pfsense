package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/laurigates/terraform-provider-pfsense/internal/provider/resource_firewall_alias"
)

// applyIfRequested is the apply-staging decision + POST that the shared CRUD
// helper runs after each mutating call. These tests exercise it directly
// against an httptest server (no Terraform runtime), through the real firewall
// alias mapper, covering apply-on-true, apply-on-default (null/unknown), and
// skip-on-false.

func aliasResourceFor(url string) *envelopeResource[resource_firewall_alias.FirewallAliasModel] {
	return &envelopeResource[resource_firewall_alias.FirewallAliasModel]{
		client: NewClient(url, "test-key", false),
		mapper: firewallAliasMapper{},
	}
}

func TestApplyIfRequested(t *testing.T) {
	cases := []struct {
		name    string
		apply   types.Bool
		wantHit int
	}{
		{"explicit true applies", types.BoolValue(true), 1},
		{"null defaults to apply", types.BoolNull(), 1},
		{"unknown defaults to apply", types.BoolUnknown(), 1},
		{"explicit false skips", types.BoolValue(false), 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var hits int
			var mu sync.Mutex
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v2/firewall/apply" {
					t.Errorf("unexpected path %q", r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("apply method = %q, want POST", r.Method)
				}
				mu.Lock()
				hits++
				mu.Unlock()
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"code":200,"status":"ok","response_id":"SUCCESS","message":"","data":{"applied":true,"pending_subsystems":[]}}`))
			}))
			defer server.Close()

			r := aliasResourceFor(server.URL)
			if err := r.applyIfRequested(context.Background(), tc.apply); err != nil {
				t.Fatalf("applyIfRequested: %v", err)
			}
			if hits != tc.wantHit {
				t.Errorf("apply POST count = %d, want %d", hits, tc.wantHit)
			}
		})
	}
}

// noApplyMapper wraps the alias mapper but declares no apply endpoint, so
// applyIfRequested must never issue a request.
type noApplyMapper struct{ firewallAliasMapper }

func (noApplyMapper) applyPath() string { return "" }

func TestApplyIfRequestedNoApplyPath(t *testing.T) {
	var hits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	r := &envelopeResource[resource_firewall_alias.FirewallAliasModel]{
		client: NewClient(server.URL, "k", false),
		mapper: noApplyMapper{},
	}
	if err := r.applyIfRequested(context.Background(), types.BoolValue(true)); err != nil {
		t.Fatalf("expected no-op, got %v", err)
	}
	if hits != 0 {
		t.Errorf("empty applyPath issued %d requests, want 0", hits)
	}
}
