package oauth

import (
	"net/url"
	"sync"
	"testing"
	"time"
)

func TestSessionStore_Stop_Idempotent(t *testing.T) {
	store := NewSessionStore()

	store.Stop()
	store.Stop()

	select {
	case <-store.stopCh:
		// ok
	case <-time.After(time.Second):
		t.Fatal("stopCh 未关闭")
	}
}

func TestSessionStore_Stop_Concurrent(t *testing.T) {
	store := NewSessionStore()

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.Stop()
		}()
	}

	wg.Wait()

	select {
	case <-store.stopCh:
		// ok
	case <-time.After(time.Second):
		t.Fatal("stopCh 未关闭")
	}
}

func TestBuildAuthorizationURL_IncludesNonEmptyRedirectURI(t *testing.T) {
	if RedirectURI == "" {
		t.Fatal("RedirectURI must not be empty")
	}

	authURL := BuildAuthorizationURL("state123", "challenge123", ScopeOAuth)
	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("BuildAuthorizationURL() returned invalid URL: %v", err)
	}

	redirect := parsed.Query().Get("redirect_uri")
	if redirect == "" {
		t.Fatal("redirect_uri query parameter must not be empty")
	}
	if redirect != RedirectURI {
		t.Fatalf("redirect_uri = %q, want %q", redirect, RedirectURI)
	}
}
