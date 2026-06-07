package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCurrentUserForwardsBearerTokenAndUnwrapsCoreEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/user/profile" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token-123" {
			t.Fatalf("Authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{"id":42,"email":"u@example.com","role":"user"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL+"/api/v1", time.Second)
	user, err := client.CurrentUser(context.Background(), "token-123")
	if err != nil {
		t.Fatalf("CurrentUser error: %v", err)
	}
	if user.ID != 42 || user.Email != "u@example.com" || user.Role != "user" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestCurrentUserRejectsCoreError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":401,"message":"bad token"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL+"/api/v1", time.Second)
	_, err := client.CurrentUser(context.Background(), "bad")
	if err == nil {
		t.Fatal("expected error")
	}
}
