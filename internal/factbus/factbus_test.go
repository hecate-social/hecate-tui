package factbus

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestParsesSSEFact(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("ResponseWriter does not support Flusher")
		}
		fmt.Fprint(w, "data: {\"fact_type\":\"venture_setup_v1\",\"data\":{\"name\":\"test\"}}\n\n")
		flusher.Flush()
		// Keep connection open briefly so client can read
		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()

	conn := NewConnection("", server.URL)
	defer conn.Close()

	go conn.connectLoop()

	select {
	case fact := <-conn.factChan:
		if fact.FactType != "venture_setup_v1" {
			t.Errorf("Expected fact_type 'venture_setup_v1', got '%s'", fact.FactType)
		}
		if string(fact.Data) != `{"name":"test"}` {
			t.Errorf("Expected data '{\"name\":\"test\"}', got '%s'", string(fact.Data))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for fact")
	}
}

func TestSkipsCommentsAndHeartbeats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		flusher := w.(http.Flusher)
		// Send heartbeat comment first
		fmt.Fprint(w, ": heartbeat\n\n")
		flusher.Flush()
		// Then a real fact
		fmt.Fprint(w, "data: {\"fact_type\":\"real_fact\",\"data\":{}}\n\n")
		flusher.Flush()
		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()

	conn := NewConnection("", server.URL)
	defer conn.Close()

	go conn.connectLoop()

	select {
	case fact := <-conn.factChan:
		if fact.FactType != "real_fact" {
			t.Errorf("Expected 'real_fact', got '%s'", fact.FactType)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for fact")
	}
}

func TestSkipsDONE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		flusher := w.(http.Flusher)
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
		fmt.Fprint(w, "data: {\"fact_type\":\"after_done\",\"data\":{}}\n\n")
		flusher.Flush()
		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()

	conn := NewConnection("", server.URL)
	defer conn.Close()

	go conn.connectLoop()

	select {
	case fact := <-conn.factChan:
		// [DONE] should be skipped, first real fact should be after_done
		if fact.FactType != "after_done" {
			t.Errorf("Expected 'after_done', got '%s'", fact.FactType)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for fact")
	}
}

func TestSkipsMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		flusher := w.(http.Flusher)
		// Malformed JSON — should be skipped, no panic
		fmt.Fprint(w, "data: not-json\n\n")
		flusher.Flush()
		// Valid fact follows
		fmt.Fprint(w, "data: {\"fact_type\":\"valid\",\"data\":{}}\n\n")
		flusher.Flush()
		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()

	conn := NewConnection("", server.URL)
	defer conn.Close()

	go conn.connectLoop()

	select {
	case fact := <-conn.factChan:
		if fact.FactType != "valid" {
			t.Errorf("Expected 'valid', got '%s'", fact.FactType)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for fact")
	}
}

func TestPollCmdReturnsContinueWhenEmpty(t *testing.T) {
	conn := NewConnection("", "http://localhost:99999")

	cmd := conn.PollCmd()
	msg := cmd()

	if _, ok := msg.(FactContinueMsg); !ok {
		t.Errorf("Expected FactContinueMsg, got %T", msg)
	}
}

func TestCloseDisconnects(t *testing.T) {
	// Server that stays open until client disconnects
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		flusher := w.(http.Flusher)
		fmt.Fprint(w, ": connected\n\n")
		flusher.Flush()
		// Block until client disconnects
		<-r.Context().Done()
	}))
	defer server.Close()

	conn := NewConnection("", server.URL)
	go conn.connectLoop()

	// Give time for connection to establish
	time.Sleep(200 * time.Millisecond)

	// Close should cancel context and close channel
	conn.Close()

	// Channel should eventually close, producing FactDisconnectedMsg
	select {
	case _, ok := <-conn.factChan:
		if ok {
			// Got a fact — that's fine, drain until closed
			for range conn.factChan {
			}
		}
		// Channel closed — expected
	case <-time.After(3 * time.Second):
		t.Fatal("Timed out waiting for channel close")
	}
}
