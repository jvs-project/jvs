package errclass

import "fmt"

// JVSError is a stable, machine-readable error class.
type JVSError struct {
	Code    string
	Message string
}

func (e *JVSError) Error() string {
	if e.Message == "" {
		return e.Code
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *JVSError) Is(target error) bool {
	t, ok := target.(*JVSError)
	return ok && e.Code == t.Code
}

// WithMessage returns a new JVSError with the same Code but a specific message.
func (e *JVSError) WithMessage(msg string) *JVSError {
	return &JVSError{Code: e.Code, Message: msg}
}

// WithMessagef returns a new JVSError with a formatted message.
func (e *JVSError) WithMessagef(format string, args ...any) *JVSError {
	return &JVSError{Code: e.Code, Message: fmt.Sprintf(format, args...)}
}

// All stable error classes for v0.x (15 total).
var (
	ErrNameInvalid            = &JVSError{Code: "E_NAME_INVALID"}
	ErrPathEscape             = &JVSError{Code: "E_PATH_ESCAPE"}
	ErrLockConflict           = &JVSError{Code: "E_LOCK_CONFLICT"}
	ErrLockExpired            = &JVSError{Code: "E_LOCK_EXPIRED"}
	ErrLockNotHeld            = &JVSError{Code: "E_LOCK_NOT_HELD"}
	ErrFencingMismatch        = &JVSError{Code: "E_FENCING_MISMATCH"}
	ErrClockSkewExceeded      = &JVSError{Code: "E_CLOCK_SKEW_EXCEEDED"}
	ErrConsistencyUnavailable = &JVSError{Code: "E_CONSISTENCY_UNAVAILABLE"}
	ErrDescriptorCorrupt      = &JVSError{Code: "E_DESCRIPTOR_CORRUPT"}
	ErrPayloadHashMismatch    = &JVSError{Code: "E_PAYLOAD_HASH_MISMATCH"}
	ErrLineageBroken          = &JVSError{Code: "E_LINEAGE_BROKEN"}
	ErrPartialSnapshot        = &JVSError{Code: "E_PARTIAL_SNAPSHOT"}
	ErrGCPlanMismatch         = &JVSError{Code: "E_GC_PLAN_MISMATCH"}
	ErrFormatUnsupported      = &JVSError{Code: "E_FORMAT_UNSUPPORTED"}
	ErrAuditChainBroken       = &JVSError{Code: "E_AUDIT_CHAIN_BROKEN"}
)
