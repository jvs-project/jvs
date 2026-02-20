package uuidutil

import (
	"crypto/rand"
	"fmt"
)

// NewV4 generates a random UUID v4 string.
// Panics if crypto/rand fails (system-level error, should never happen on a healthy system).
func NewV4() string {
	var u [16]byte
	if _, err := rand.Read(u[:]); err != nil {
		// This is a system-level error (entropy exhaustion, broken random source).
		// There's no recovery, so we panic rather than return an error that would
		// need to be handled everywhere.
		panic("jvs: crypto/rand failed (system error): " + err.Error())
	}
	u[6] = (u[6] & 0x0f) | 0x40 // version 4
	u[8] = (u[8] & 0x3f) | 0x80 // variant RFC 4122
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
}
