// Package webhook provides HTTP webhook notification support for JVS events.
package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// EventType represents the type of JVS event that can trigger webhooks.
type EventType string

const (
	EventSnapshotCreated EventType = "snapshot.created"
	EventSnapshotDeleted EventType = "snapshot.deleted"
	EventRestoreStart    EventType = "restore.start"
	EventRestoreComplete EventType = "restore.complete"
	EventRestoreFailed   EventType = "restore.failed"
	EventGCStart         EventType = "gc.start"
	EventGCComplete      EventType = "gc.complete"
	EventVerifyStart     EventType = "verify.start"
	EventVerifyComplete  EventType = "verify.complete"
	EventVerifyFailed    EventType = "verify.failed"
)

// Event represents a JVS event payload sent to webhooks.
type Event struct {
	Event      EventType            `json:"event"`
	Timestamp  string               `json:"timestamp"`
	RepoID     string               `json:"repo_id,omitempty"`
	RepoRoot   string               `json:"repo_root,omitempty"`
	SnapshotID string               `json:"snapshot_id,omitempty"`
	Note       string               `json:"note,omitempty"`
	Tags       []string             `json:"tags,omitempty"`
	Error      string               `json:"error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// HookConfig represents a single webhook configuration.
type HookConfig struct {
	URL      string      `json:"url" toml:"url"`
	Secret   string      `json:"secret,omitempty" toml:"secret,omitempty"`
	Events   []EventType `json:"events" toml:"events"`
	Timeout  time.Duration `json:"timeout" toml:"timeout"`
	Enabled  bool        `json:"enabled" toml:"enabled"`
}

// Config represents the webhook configuration.
type Config struct {
	Hooks           []HookConfig `json:"hooks" toml:"hooks"`
	Enabled         bool         `json:"enabled" toml:"enabled"`
	MaxRetries      int          `json:"max_retries" toml:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay" toml:"retry_delay"`
	AsyncQueueSize  int          `json:"async_queue_size" toml:"async_queue_size"`
}

// DefaultConfig returns the default webhook configuration.
func DefaultConfig() *Config {
	return &Config{
		Enabled:        true,
		MaxRetries:     3,
		RetryDelay:     5 * time.Second,
		AsyncQueueSize: 100,
	}
}

// Client handles sending webhook notifications.
type Client struct {
	config    *Config
	http      *http.Client
	queue     chan *job
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	once      sync.Once
	mu        sync.RWMutex
	eventChan chan Event
}

type job struct {
	event Event
	hook  HookConfig
}

// NewClient creates a new webhook client.
func NewClient(cfg *Config) *Client {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := &Client{
		config:    cfg,
		http:      &http.Client{Timeout: 30 * time.Second},
		queue:     make(chan *job, cfg.AsyncQueueSize),
		ctx:       ctx,
		cancel:    cancel,
		eventChan: make(chan Event, cfg.AsyncQueueSize),
	}

	// Start background worker if enabled
	if cfg.Enabled {
		c.start()
	}

	return c
}

// Start starts the background webhook worker.
func (c *Client) start() {
	c.once.Do(func() {
		c.wg.Add(1)
		go c.worker()
	})
}

// worker processes webhook notifications in the background.
func (c *Client) worker() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			// Drain remaining jobs
			for len(c.queue) > 0 {
				job := <-c.queue
				c.send(job)
			}
			return
		case job := <-c.queue:
			c.send(job)
		}
	}
}

// Send sends an event to all matching webhooks.
// If async is true, the event is queued for background sending.
// If async is false, the event is sent synchronously.
func (c *Client) Send(event Event, async bool) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.config.Enabled {
		return nil
	}

	// Find matching hooks
	var hooks []HookConfig
	for _, hook := range c.config.Hooks {
		if !hook.Enabled {
			continue
		}
		if c.matchesEvent(hook, event.Event) {
			hooks = append(hooks, hook)
		}
	}

	if len(hooks) == 0 {
		return nil
	}

	// Set timestamp if not set
	if event.Timestamp == "" {
		event.Timestamp = time.Now().Format(time.RFC3339)
	}

	if async {
		// Queue all webhook calls
		for _, hook := range hooks {
			job := &job{
				event: event,
				hook:  hook,
			}
			select {
			case c.queue <- job:
			default:
				// Queue full, log warning but don't block
				fmt.Printf("Warning: webhook queue full, dropping event: %s\n", event.Event)
			}
		}
		return nil
	}

	// Send synchronously
	var lastErr error
	for _, hook := range hooks {
		if err := c.sendSync(&job{event: event, hook: hook}); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// send sends a single webhook job.
func (c *Client) send(job *job) {
	if err := c.sendSync(job); err != nil {
		fmt.Printf("Webhook error: %v\n", err)
	}
}

// sendSync sends a webhook synchronously with retries.
func (c *Client) sendSync(job *job) error {
	payload, err := json.Marshal(job.event)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-c.ctx.Done():
				return c.ctx.Err()
			case <-time.After(c.config.RetryDelay):
			}
		}

		req, err := c.createRequest(job.hook, payload)
		if err != nil {
			return err
		}

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("http request: %w", err)
			continue
		}

		// Read and close body
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		lastErr = fmt.Errorf("http %d: %s", resp.StatusCode, string(body))
	}

	return lastErr
}

// createRequest creates an HTTP request for the webhook.
func (c *Client) createRequest(hook HookConfig, payload []byte) (*http.Request, error) {
	req, err := http.NewRequest("POST", hook.URL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "JVS-Webhook/1.0")
	req.Header.Set("X-JVS-Event", string(payload))

	// Add HMAC signature if secret is configured
	if hook.Secret != "" {
		signature := c.sign(payload, hook.Secret)
		req.Header.Set("X-JVS-Signature", signature)
	}

	return req, nil
}

// sign creates an HMAC-SHA256 signature for the payload.
func (c *Client) sign(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// matchesEvent checks if a hook is configured for the given event.
func (c *Client) matchesEvent(hook HookConfig, event EventType) bool {
	for _, e := range hook.Events {
		if e == event || e == "*" {
			return true
		}
	}
	return false
}

// Close gracefully shuts down the webhook client.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.config.Enabled {
		return nil
	}

	c.cancel()
	c.wg.Wait()
	return nil
}

// SendSnapshotCreated sends a snapshot.created event.
func (c *Client) SendSnapshotCreated(repoID, repoRoot, snapshotID, note string, tags []string, async bool) error {
	return c.Send(Event{
		Event:      EventSnapshotCreated,
		RepoID:     repoID,
		RepoRoot:   repoRoot,
		SnapshotID: snapshotID,
		Note:       note,
		Tags:       tags,
	}, async)
}

// SendSnapshotDeleted sends a snapshot.deleted event.
func (c *Client) SendSnapshotDeleted(repoID, repoRoot, snapshotID string, async bool) error {
	return c.Send(Event{
		Event:      EventSnapshotDeleted,
		RepoID:     repoID,
		RepoRoot:   repoRoot,
		SnapshotID: snapshotID,
	}, async)
}

// SendRestoreStart sends a restore.start event.
func (c *Client) SendRestoreStart(repoID, repoRoot, snapshotID string, async bool) error {
	return c.Send(Event{
		Event:      EventRestoreStart,
		RepoID:     repoID,
		RepoRoot:   repoRoot,
		SnapshotID: snapshotID,
	}, async)
}

// SendRestoreComplete sends a restore.complete event.
func (c *Client) SendRestoreComplete(repoID, repoRoot, snapshotID string, async bool) error {
	return c.Send(Event{
		Event:      EventRestoreComplete,
		RepoID:     repoID,
		RepoRoot:   repoRoot,
		SnapshotID: snapshotID,
	}, async)
}

// SendRestoreFailed sends a restore.failed event.
func (c *Client) SendRestoreFailed(repoID, repoRoot, snapshotID, errMsg string, async bool) error {
	return c.Send(Event{
		Event:      EventRestoreFailed,
		RepoID:     repoID,
		RepoRoot:   repoRoot,
		SnapshotID: snapshotID,
		Error:      errMsg,
	}, async)
}

// SendGCStart sends a gc.start event.
func (c *Client) SendGCStart(repoID, repoRoot string, async bool) error {
	return c.Send(Event{
		Event:    EventGCStart,
		RepoID:   repoID,
		RepoRoot: repoRoot,
	}, async)
}

// SendGCComplete sends a gc.complete event.
func (c *Client) SendGCComplete(repoID, repoRoot string, freedBytes int64, snapshotsDeleted int, async bool) error {
	return c.Send(Event{
		Event:    EventGCComplete,
		RepoID:   repoID,
		RepoRoot: repoRoot,
		Metadata: map[string]interface{}{
			"freed_bytes":       freedBytes,
			"snapshots_deleted": snapshotsDeleted,
		},
	}, async)
}

// SendVerifyStart sends a verify.start event.
func (c *Client) SendVerifyStart(repoID, repoRoot string, async bool) error {
	return c.Send(Event{
		Event:    EventVerifyStart,
		RepoID:   repoID,
		RepoRoot: repoRoot,
	}, async)
}

// SendVerifyComplete sends a verify.complete event.
func (c *Client) SendVerifyComplete(repoID, repoRoot string, snapshotsVerified int, async bool) error {
	return c.Send(Event{
		Event:    EventVerifyComplete,
		RepoID:   repoID,
		RepoRoot: repoRoot,
		Metadata: map[string]interface{}{
			"snapshots_verified": snapshotsVerified,
		},
	}, async)
}

// SendVerifyFailed sends a verify.failed event.
func (c *Client) SendVerifyFailed(repoID, repoRoot, snapshotID, errMsg string, async bool) error {
	return c.Send(Event{
		Event:      EventVerifyFailed,
		RepoID:     repoID,
		RepoRoot:   repoRoot,
		SnapshotID: snapshotID,
		Error:      errMsg,
	}, async)
}
