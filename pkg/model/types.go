package model

// EngineType identifies the snapshot engine used.
type EngineType string

const (
	EngineJuiceFSClone EngineType = "juicefs-clone"
	EngineReflinkCopy  EngineType = "reflink-copy"
	EngineCopy         EngineType = "copy"
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
