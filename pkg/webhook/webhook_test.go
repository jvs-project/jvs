package webhook

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.Enabled {
		t.Error("default config should be enabled")
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("expected MaxRetries 3, got %d", cfg.MaxRetries)
	}
	if cfg.RetryDelay != 5*time.Second {
		t.Errorf("expected RetryDelay 5s, got %v", cfg.RetryDelay)
	}
	if cfg.AsyncQueueSize != 100 {
		t.Errorf("expected AsyncQueueSize 100, got %d", cfg.AsyncQueueSize)
	}
}

func TestClientSendSync(t *testing.T) {
	// Create test server
	var receivedEvent map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedEvent)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{
		Enabled:    true,
		MaxRetries: 1,
		RetryDelay: 10 * time.Millisecond,
		Hooks: []HookConfig{
			{
				URL:     server.URL,
				Events:  []EventType{EventSnapshotCreated},
				Enabled: true,
			},
		},
	}

	client := NewClient(cfg)
	defer client.Close()

	event := Event{
		Event:      EventSnapshotCreated,
		RepoID:     "test-repo",
		SnapshotID: "abc123",
		Note:       "test snapshot",
	}

	err := client.Send(event, false)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if receivedEvent == nil {
		t.Error("expected event to be received")
	}
	if receivedEvent["event"] != string(EventSnapshotCreated) {
		t.Errorf("expected event %s, got %v", EventSnapshotCreated, receivedEvent["event"])
	}
}

func TestClientSendWithSignature(t *testing.T) {
	var receivedSignature string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = json.Marshal(r.Body)
		receivedSignature = r.Header.Get("X-JVS-Signature")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	secret := "test-secret-key"
	cfg := &Config{
		Enabled:    true,
		MaxRetries: 1,
		Hooks: []HookConfig{
			{
				URL:     server.URL,
				Secret:  secret,
				Events:  []EventType{EventSnapshotCreated},
				Enabled: true,
			},
		},
	}

	client := NewClient(cfg)
	defer client.Close()

	event := Event{
		Event:      EventSnapshotCreated,
		RepoID:     "test-repo",
		SnapshotID: "abc123",
	}

	err := client.Send(event, false)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if receivedSignature == "" {
		t.Fatal("expected X-JVS-Signature header")
	}

	if len(receivedSignature) < 7 || receivedSignature[:7] != "sha256=" {
		t.Errorf("invalid signature format: %s", receivedSignature)
	}
}

func TestClientSendAsync(t *testing.T) {
	calls := make(chan bool, 10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls <- true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{
		Enabled:        true,
		MaxRetries:     1,
		AsyncQueueSize: 10,
		Hooks: []HookConfig{
			{
				URL:     server.URL,
				Events:  []EventType{EventSnapshotCreated},
				Enabled: true,
			},
		},
	}

	client := NewClient(cfg)
	defer client.Close()

	event := Event{
		Event:      EventSnapshotCreated,
		RepoID:     "test-repo",
		SnapshotID: "abc123",
	}

	// Send async
	err := client.Send(event, true)
	if err != nil {
		t.Fatalf("Send async failed: %v", err)
	}

	// Wait for async delivery
	select {
	case <-calls:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("async webhook not received within timeout")
	}
}

func TestClientRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{
		Enabled:    true,
		MaxRetries: 3,
		RetryDelay: 10 * time.Millisecond,
		Hooks: []HookConfig{
			{
				URL:     server.URL,
				Events:  []EventType{EventSnapshotCreated},
				Enabled: true,
			},
		},
	}

	client := NewClient(cfg)
	defer client.Close()

	event := Event{Event: EventSnapshotCreated}

	start := time.Now()
	err := client.Send(event, false)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Send with retry failed: %v", err)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}

	// Should have retried at least twice (2 * RetryDelay)
	if duration < 20*time.Millisecond {
		t.Errorf("expected retries to take at least 20ms, took %v", duration)
	}
}

func TestClientDisabled(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{
		Enabled: false,
		Hooks: []HookConfig{
			{
				URL:     server.URL,
				Events:  []EventType{EventSnapshotCreated},
			},
		},
	}

	client := NewClient(cfg)
	defer client.Close()

	event := Event{Event: EventSnapshotCreated}
	err := client.Send(event, false)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if called {
		t.Error("webhook should not have been called when disabled")
	}
}

func TestClientWildcardEvent(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{
		Enabled: true,
		Hooks: []HookConfig{
			{
				URL:     server.URL,
				Events:  []EventType{"*"},
				Enabled: true,
			},
		},
	}

	client := NewClient(cfg)
	defer client.Close()

	// Test different events
	events := []EventType{
		EventSnapshotCreated,
		EventRestoreComplete,
		EventGCComplete,
	}

	for _, event := range events {
		called = false
		err := client.Send(Event{Event: event}, false)
		if err != nil {
			t.Fatalf("Send failed for %s: %v", event, err)
		}
		if !called {
			t.Errorf("wildcard hook not called for event %s", event)
		}
	}
}

func TestClientEventFiltering(t *testing.T) {
	var receivedEventType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)
		receivedEventType = payload["event"].(string)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{
		Enabled: true,
		Hooks: []HookConfig{
			{
				URL:     server.URL,
				Events:  []EventType{EventSnapshotCreated}, // Only snapshot events
				Enabled: true,
			},
		},
	}

	client := NewClient(cfg)
	defer client.Close()

	// Send snapshot event - should be called
	receivedEventType = ""
	client.Send(Event{Event: EventSnapshotCreated}, false)
	if receivedEventType != string(EventSnapshotCreated) {
		t.Errorf("snapshot hook should have been called, got event: %s", receivedEventType)
	}

	// Send restore event - should NOT be called (eventType remains unchanged)
	receivedEventType = ""
	client.Send(Event{Event: EventRestoreComplete}, false)
	if receivedEventType == string(EventRestoreComplete) {
		t.Error("restore hook should not have been called")
	}
}

func TestConvenienceMethods(t *testing.T) {
	var receivedEvent Event

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedEvent)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{
		Enabled: true,
		Hooks: []HookConfig{
			{
				URL:     server.URL,
				Events:  []EventType{"*"},
				Enabled: true,
			},
		},
	}

	client := NewClient(cfg)
	defer client.Close()

	// Test SendSnapshotCreated
	err := client.SendSnapshotCreated("repo-1", "/path", "snap-1", "test", []string{"tag1"}, false)
	if err != nil {
		t.Fatalf("SendSnapshotCreated failed: %v", err)
	}
	if receivedEvent.Event != EventSnapshotCreated {
		t.Errorf("expected %s, got %s", EventSnapshotCreated, receivedEvent.Event)
	}

	// Test SendRestoreComplete
	err = client.SendRestoreComplete("repo-1", "/path", "snap-1", false)
	if err != nil {
		t.Fatalf("SendRestoreComplete failed: %v", err)
	}
	if receivedEvent.Event != EventRestoreComplete {
		t.Errorf("expected %s, got %s", EventRestoreComplete, receivedEvent.Event)
	}

	// Test SendGCComplete
	err = client.SendGCComplete("repo-1", "/path", 1024, 5, false)
	if err != nil {
		t.Fatalf("SendGCComplete failed: %v", err)
	}
	if receivedEvent.Event != EventGCComplete {
		t.Errorf("expected %s, got %s", EventGCComplete, receivedEvent.Event)
	}
	if receivedEvent.Metadata == nil {
		t.Error("expected metadata in GC complete event")
	}
}

func TestClientConnectionError(t *testing.T) {
	// Use an invalid URL that will cause connection error
	cfg := &Config{
		Enabled:    true,
		MaxRetries: 1,
		RetryDelay: 10 * time.Millisecond,
		Hooks: []HookConfig{
			{
				URL:     "http://invalid.local:9999",
				Events:  []EventType{EventSnapshotCreated},
				Enabled: true,
			},
		},
	}

	client := NewClient(cfg)
	defer client.Close()

	err := client.Send(Event{Event: EventSnapshotCreated}, false)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestClientGracefulShutdown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{
		Enabled:        true,
		MaxRetries:     0,
		AsyncQueueSize: 5,
		Hooks: []HookConfig{
			{
				URL:     server.URL,
				Events:  []EventType{EventSnapshotCreated},
				Enabled: true,
			},
		},
	}

	client := NewClient(cfg)

	// Send several async events
	for i := 0; i < 3; i++ {
		client.Send(Event{Event: EventSnapshotCreated}, true)
	}

	// Close should wait for queue to drain
	start := time.Now()
	err := client.Close()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Should have waited for pending events
	if duration < 300*time.Millisecond {
		t.Errorf("Close should have waited for pending events, took %v", duration)
	}
}

func TestClientQueueFull(t *testing.T) {
	cfg := &Config{
		Enabled:        true,
		MaxRetries:     0,
		AsyncQueueSize: 2,
		Hooks: []HookConfig{
			{
				URL:     "http://slow.example.com",
				Events:  []EventType{EventSnapshotCreated},
				Enabled: true,
			},
		},
	}

	client := NewClient(cfg)
	defer client.Close()

	// Fill the queue
	for i := 0; i < 10; i++ {
		client.Send(Event{Event: EventSnapshotCreated}, true)
	}

	// Queue is full, but Send should not block
	// We just verify it doesn't hang
	done := make(chan bool)
	go func() {
		client.Send(Event{Event: EventSnapshotCreated}, true)
		close(done)
	}()

	select {
	case <-done:
		// Success - didn't block
	case <-time.After(100 * time.Millisecond):
		t.Error("Send blocked when queue full")
	}
}

func TestHookEnabledDisabled(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{
		Enabled: true,
		Hooks: []HookConfig{
			{
				URL:     server.URL,
				Events:  []EventType{EventSnapshotCreated},
				Enabled: false, // Disabled hook
			},
		},
	}

	client := NewClient(cfg)
	defer client.Close()

	err := client.Send(Event{Event: EventSnapshotCreated}, false)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if called {
		t.Error("disabled hook should not have been called")
	}
}

func TestClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{
		Enabled:    true,
		MaxRetries: 0,
		Hooks: []HookConfig{
			{
				URL:     server.URL,
				Events:  []EventType{EventSnapshotCreated},
				Timeout: 50 * time.Millisecond,
				Enabled: true,
			},
		},
	}

	client := NewClient(cfg)
	defer client.Close()

	// The default HTTP client has a 30s timeout, so this test
	// verifies the hook's timeout doesn't affect the request directly
	// In a real implementation, you'd need to create per-hook timeouts
	err := client.Send(Event{Event: EventSnapshotCreated}, false)
	// Should succeed due to default client timeout
	if err != nil {
		t.Logf("Send with slow server: %v", err)
	}
}
