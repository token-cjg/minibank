package main

import (
	"net/http"
	"testing"
	"time"
)

func TestNewHTTPServer(t *testing.T) {
	// Create a dummy handler.
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	s := newHTTPServer(dummyHandler)

	if s.Addr != ":8080" {
		t.Errorf("expected Addr ':8080', got %q", s.Addr)
	}
	if s.Handler == nil {
		t.Error("expected non-nil Handler")
	}
	if s.ReadTimeout != 10*time.Second {
		t.Errorf("expected ReadTimeout 10s, got %v", s.ReadTimeout)
	}
	if s.WriteTimeout != 10*time.Second {
		t.Errorf("expected WriteTimeout 10s, got %v", s.WriteTimeout)
	}
	if s.IdleTimeout != 60*time.Second {
		t.Errorf("expected IdleTimeout 60s, got %v", s.IdleTimeout)
	}
}
