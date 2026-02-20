package model

// EngineType identifies the snapshot engine used.
type EngineType string

const (
	EngineJuiceFSClone EngineType = "juicefs-clone"
	EngineReflinkCopy  EngineType = "reflink-copy"
	EngineCopy         EngineType = "copy"
)

// ConsistencyLevel specifies the snapshot consistency guarantee.
type ConsistencyLevel string

const (
	ConsistencyQuiesced    ConsistencyLevel = "quiesced"
	ConsistencyBestEffort  ConsistencyLevel = "best_effort"
)

// IntegrityState represents the verification status of a snapshot.
type IntegrityState string

const (
	IntegrityVerified IntegrityState = "verified"
	IntegrityTampered IntegrityState = "tampered"
	IntegrityUnknown  IntegrityState = "unknown"
)

// HashValue is a SHA-256 hash stored as hex string.
type HashValue string

// Isolation mode constants (v0.x exclusive only).
const (
	IsolationExclusive = "exclusive"
)

// LockState represents the current state of a lock.
type LockState string

const (
	LockStateHeld    LockState = "held"
	LockStateExpired LockState = "expired"
	LockStateFree    LockState = "free"
)
