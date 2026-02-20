package errclass_test

import (
	"errors"
	"testing"

	"github.com/jvs-project/jvs/pkg/errclass"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJVSError_Error(t *testing.T) {
	err := errclass.ErrLockConflict.WithMessage("worktree main is locked")
	assert.Equal(t, "E_LOCK_CONFLICT: worktree main is locked", err.Error())
}

func TestJVSError_Is(t *testing.T) {
	err := errclass.ErrLockConflict.WithMessage("specific message")
	require.True(t, errors.Is(err, errclass.ErrLockConflict))
	require.False(t, errors.Is(err, errclass.ErrLockExpired))
}

func TestJVSError_Code(t *testing.T) {
	assert.Equal(t, "E_LOCK_CONFLICT", errclass.ErrLockConflict.Code)
	assert.Equal(t, "E_FENCING_MISMATCH", errclass.ErrFencingMismatch.Code)
}

func TestJVSError_AllErrorsDefined(t *testing.T) {
	// All 15 v0.x error classes must exist
	all := []error{
		errclass.ErrNameInvalid,
		errclass.ErrPathEscape,
		errclass.ErrLockConflict,
		errclass.ErrLockExpired,
		errclass.ErrLockNotHeld,
		errclass.ErrFencingMismatch,
		errclass.ErrClockSkewExceeded,
		errclass.ErrConsistencyUnavailable,
		errclass.ErrDescriptorCorrupt,
		errclass.ErrPayloadHashMismatch,
		errclass.ErrLineageBroken,
		errclass.ErrPartialSnapshot,
		errclass.ErrGCPlanMismatch,
		errclass.ErrFormatUnsupported,
		errclass.ErrAuditChainBroken,
	}
	assert.Len(t, all, 15)
}
