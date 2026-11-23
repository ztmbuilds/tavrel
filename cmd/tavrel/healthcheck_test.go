package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthcheckHealthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	if err := healthcheck(srv.Listener.Addr().String()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestHealthcheckUnhealthyStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	if err := healthcheck(srv.Listener.Addr().String()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestHealthcheckUnreachable(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	addr := lis.Addr().String()
	lis.Close()

	if err := healthcheck(addr); err == nil {
		t.Fatal("expected error, got nil")
	}
}
