package sender_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func TestJiraSender_PostsCorrectPayload(t *testing.T) {
	var captured map[string]interface{}
	var capturedAuth string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		u, _, _ := r.BasicAuth()
		capturedAuth = u
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	s := sender.NewJiraSender(ts.URL, "OPS", "Bug", "admin", "secret")

	err := s.Send(alert.Alert{
		LeaseID: "aws/creds/my-role/abc123",
		TTL:     "2h",
		Level:   alert.LevelWarning,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if capturedAuth != "admin" {
		t.Errorf("expected basic auth user 'admin', got %q", capturedAuth)
	}

	fields, ok := captured["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("missing 'fields' in payload")
	}

	project, _ := fields["project"].(map[string]interface{})
	if project["key"] != "OPS" {
		t.Errorf("expected project key 'OPS', got %v", project["key"])
	}

	issuetype, _ := fields["issuetype"].(map[string]interface{})
	if issuetype["name"] != "Bug" {
		t.Errorf("expected issuetype 'Bug', got %v", issuetype["name"])
	}
}

func TestJiraSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	s := sender.NewJiraSender(ts.URL, "OPS", "Bug", "admin", "secret")

	err := s.Send(alert.Alert{
		LeaseID: "pki/issue/role/xyz",
		TTL:     "30m",
		Level:   alert.LevelCritical,
	})
	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestJiraSender_CriticalUsesHighPriority(t *testing.T) {
	var captured map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	s := sender.NewJiraSender(ts.URL, "OPS", "Task", "user", "token")

	_ = s.Send(alert.Alert{
		LeaseID: "database/creds/role/def456",
		TTL:     "10m",
		Level:   alert.LevelCritical,
	})

	fields, _ := captured["fields"].(map[string]interface{})
	priority, _ := fields["priority"].(map[string]interface{})
	if priority["name"] != "High" {
		t.Errorf("expected priority 'High' for critical alert, got %v", priority["name"])
	}
}
