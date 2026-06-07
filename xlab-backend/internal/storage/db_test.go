package storage

import (
	"context"
	"strings"
	"testing"
)

func TestOpenPostgresWrapsPingErrors(t *testing.T) {
	_, err := OpenPostgres(context.Background(), "postgres://%")
	if err == nil {
		t.Fatal("expected OpenPostgres error")
	}
	if !strings.Contains(err.Error(), "ping postgres") {
		t.Fatalf("error = %q, want ping context", err.Error())
	}
}
