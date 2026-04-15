package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newMockVaultServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/v1/sys/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"sealed":      false,
			"initialized": true,
		})
	})

	// Lease lookup endpoint
	mux.HandleFunc("/v1/sys/leases/lookup", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":  "test/lease/123",
				"ttl": float64(3600),
			},
		})
	})

	return httptest.NewServer(mux)
}

func TestNewClient(t *testing.T) {
	_, err := NewClient("http://127.0.0.1:8200", "test-token")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestIsHealthy(t *testing.T) {
	server := newMockVaultServer(t)
	defer server.Close()

	client, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if err := client.IsHealthy(); err != nil {
		t.Errorf("expected healthy vault, got error: %v", err)
	}
}

func TestLookupLease(t *testing.T) {
	server := newMockVaultServer(t)
	defer server.Close()

	client, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	info, err := client.LookupLease("test/lease/123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if info.TTL != 3600*time.Second {
		t.Errorf("expected TTL 3600s, got %v", info.TTL)
	}

	if info.ExpiresAt.Before(time.Now()) {
		t.Errorf("expected ExpiresAt to be in the future")
	}
}
