package main

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

// TestGracefulShutdown_CompletesInflightRequests verifies that http.Server.Shutdown
// waits for an in-flight request to finish before returning.
func TestGracefulShutdown_CompletesInflightRequests(t *testing.T) {
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{Handler: slow}

	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}

	go srv.Serve(ln) //nolint:errcheck

	reqDone := make(chan struct{})
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String() + "/")
		if err == nil {
			resp.Body.Close()
		}
		close(reqDone)
	}()

	// Let the request reach the handler, then initiate shutdown.
	time.Sleep(10 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}

	select {
	case <-reqDone:
		// correct: in-flight request completed before shutdown returned
	case <-time.After(200 * time.Millisecond):
		t.Fatal("in-flight request did not complete within 200ms of Shutdown")
	}
}
