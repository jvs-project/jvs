package progress

import (
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	p := New("test-op", 100, nil)
	if p.Op != "test-op" {
		t.Errorf("expected Op 'test-op', got %s", p.Op)
	}
	if p.Total != 100 {
		t.Errorf("expected Total 100, got %d", p.Total)
	}
	if p.cb == nil {
		t.Error("expected callback to be set to Noop, got nil")
	}
}

func TestNewWithCallback(t *testing.T) {
	var called bool
	cb := func(op string, current, total int, message string) {
		called = true
	}
	p := New("test-op", 100, cb)
	if p.cb == nil {
		t.Error("expected callback to be set")
	}
	// Trigger the callback
	p.Increment("test")
	if !called {
		t.Error("expected callback to be called")
	}
}

func TestIncrement(t *testing.T) {
	var mu sync.Mutex
	var calls []struct {
		op      string
		current int
		total   int
		message string
	}
	cb := func(op string, current, total int, message string) {
		mu.Lock()
		calls = append(calls, struct {
			op      string
			current int
			total   int
			message string
		}{op, current, total, message})
		mu.Unlock()
	}
	p := New("increment-test", 10, cb)

	p.Increment("step1")
	p.Increment("step2")
	p.Increment("step3")

	mu.Lock()
	if len(calls) != 3 {
		t.Errorf("expected 3 calls, got %d", len(calls))
	}
	if calls[0].current != 1 {
		t.Errorf("expected first call current=1, got %d", calls[0].current)
	}
	if calls[1].current != 2 {
		t.Errorf("expected second call current=2, got %d", calls[1].current)
	}
	if calls[2].current != 3 {
		t.Errorf("expected third call current=3, got %d", calls[2].current)
	}
	if calls[0].message != "step1" {
		t.Errorf("expected message 'step1', got %s", calls[0].message)
	}
	mu.Unlock()
}

func TestSet(t *testing.T) {
	var mu sync.Mutex
	var lastCurrent int
	var lastMessage string
	cb := func(op string, current, total int, message string) {
		mu.Lock()
		lastCurrent = current
		lastMessage = message
		mu.Unlock()
	}
	p := New("set-test", 100, cb)

	p.Set(50, "halfway")

	mu.Lock()
	if lastCurrent != 50 {
		t.Errorf("expected current 50, got %d", lastCurrent)
	}
	if lastMessage != "halfway" {
		t.Errorf("expected message 'halfway', got %s", lastMessage)
	}
	mu.Unlock()

	p.Set(75, "almost done")

	mu.Lock()
	if lastCurrent != 75 {
		t.Errorf("expected current 75, got %d", lastCurrent)
	}
	mu.Unlock()
}

func TestDone(t *testing.T) {
	var mu sync.Mutex
	var lastCurrent int
	var lastTotal int
	var lastMessage string
	cb := func(op string, current, total int, message string) {
		mu.Lock()
		lastCurrent = current
		lastTotal = total
		lastMessage = message
		mu.Unlock()
	}
	p := New("done-test", 100, cb)

	p.Done("complete")

	mu.Lock()
	if lastCurrent != 100 {
		t.Errorf("expected current 100, got %d", lastCurrent)
	}
	if lastTotal != 100 {
		t.Errorf("expected total 100, got %d", lastTotal)
	}
	if lastMessage != "complete" {
		t.Errorf("expected message 'complete', got %s", lastMessage)
	}
	mu.Unlock()
}

func TestCurrent(t *testing.T) {
	p := New("current-test", 100, nil)

	if p.Current() != 0 {
		t.Errorf("expected initial current 0, got %d", p.Current())
	}

	p.Increment("")
	if p.Current() != 1 {
		t.Errorf("expected current 1 after increment, got %d", p.Current())
	}

	p.Set(50, "")
	if p.Current() != 50 {
		t.Errorf("expected current 50 after set, got %d", p.Current())
	}

	p.Done("")
	if p.Current() != 100 {
		t.Errorf("expected current 100 after done, got %d", p.Current())
	}
}

func TestNoop(t *testing.T) {
	// Just verify Noop doesn't panic
	Noop("test", 1, 100, "message")
}

func TestProgressSequence(t *testing.T) {
	var mu sync.Mutex
	var calls []int
	cb := func(op string, current, total int, message string) {
		mu.Lock()
		calls = append(calls, current)
		mu.Unlock()
	}
	p := New("sequence-test", 10, cb)

	p.Increment("")
	p.Increment("")
	p.Set(5, "")
	p.Increment("")
	p.Done("")

	mu.Lock()
	expected := []int{1, 2, 5, 6, 10}
	if len(calls) != len(expected) {
		t.Errorf("expected %d calls, got %d", len(expected), len(calls))
	}
	for i, v := range expected {
		if i < len(calls) && calls[i] != v {
			t.Errorf("expected call %d to be %d, got %d", i, v, calls[i])
		}
	}
	mu.Unlock()
}
