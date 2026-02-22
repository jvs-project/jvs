// Package progress provides progress reporting for long-running operations.
package progress

// Callback receives progress updates during long operations.
type Callback func(op string, current, total int, message string)

// Noop is a no-op callback for default behavior.
func Noop(op string, current, total int, message string) {}

// Progress tracks operation progress.
type Progress struct {
	Op      string
	Total   int
	current int
	cb      Callback
}

// New creates a new Progress tracker.
func New(op string, total int, cb Callback) *Progress {
	if cb == nil {
		cb = Noop
	}
	return &Progress{Op: op, Total: total, cb: cb}
}

// Increment advances the progress and calls the callback.
func (p *Progress) Increment(message string) {
	p.current++
	p.cb(p.Op, p.current, p.Total, message)
}

// Set sets the current progress value.
func (p *Progress) Set(current int, message string) {
	p.current = current
	p.cb(p.Op, p.current, p.Total, message)
}

// Done marks the operation as complete.
func (p *Progress) Done(message string) {
	p.current = p.Total
	p.cb(p.Op, p.current, p.Total, message)
}

// Current returns the current progress value.
func (p *Progress) Current() int {
	return p.current
}
