package model_test

import (
	"regexp"
	"testing"

	"github.com/jvs-project/jvs/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var snapshotIDPattern = regexp.MustCompile(`^\d{13}-[0-9a-f]{8}$`)

func TestNewSnapshotID_Format(t *testing.T) {
	id := model.NewSnapshotID()
	require.Regexp(t, snapshotIDPattern, string(id))
}

func TestSnapshotID_ShortID(t *testing.T) {
	id := model.SnapshotID("1708300800000-a3f7c1b2")
	assert.Equal(t, "17083008", id.ShortID())
}

func TestNewSnapshotID_Uniqueness(t *testing.T) {
	seen := make(map[model.SnapshotID]bool)
	for i := 0; i < 100; i++ {
		id := model.NewSnapshotID()
		assert.False(t, seen[id], "duplicate: %s", id)
		seen[id] = true
	}
}
