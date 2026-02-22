// Package metrics provides Prometheus metrics export for JVS.
// This is a stub implementation while the full metrics implementation is being developed.
package metrics

import (
	"sync"
	"time"
)

var (
	enabled     bool
	enabledMutex sync.RWMutex
	defaultRegistry *Registry
)

// Init initializes the metrics system.
func Init() {
	enabledMutex.Lock()
	defer enabledMutex.Unlock()
	enabled = true
	defaultRegistry = NewRegistry()
}

// Enabled returns true if metrics are enabled.
func Enabled() bool {
	enabledMutex.RLock()
	defer enabledMutex.RUnlock()
	return enabled
}

// Default returns the default metrics registry.
func Default() *Registry {
	if defaultRegistry == nil {
		Init()
	}
	return defaultRegistry
}

// Registry holds all JVS metrics.
type Registry struct{}

// NewRegistry creates a new metrics registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// RecordSnapshot records a snapshot operation.
func (r *Registry) RecordSnapshot(success bool, duration time.Duration, sizeBytes int64, engine string) {
	// Stub implementation
}

// RecordRestore records a restore operation.
func (r *Registry) RecordRestore(success bool, duration time.Duration) {
	// Stub implementation
}

// RecordGC records a GC operation.
func (r *Registry) RecordGC(deletedCount int, duration time.Duration, bytesReclaimed int64) {
	// Stub implementation
}
