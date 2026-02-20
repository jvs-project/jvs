package pathutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jvs-project/jvs/pkg/errclass"
	"github.com/jvs-project/jvs/pkg/pathutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateName_Valid(t *testing.T) {
	valid := []string{"main", "feature-1", "v1.0", "my_branch", "A-Z.test"}
	for _, name := range valid {
		assert.NoError(t, pathutil.ValidateName(name), "should accept: %s", name)
	}
}

func TestValidateName_Empty(t *testing.T) {
	err := pathutil.ValidateName("")
	require.ErrorIs(t, err, errclass.ErrNameInvalid)
}

func TestValidateName_DotDot(t *testing.T) {
	err := pathutil.ValidateName("..")
	require.ErrorIs(t, err, errclass.ErrNameInvalid)
}

func TestValidateName_Separators(t *testing.T) {
	for _, name := range []string{"a/b", "a\\b"} {
		err := pathutil.ValidateName(name)
		require.ErrorIs(t, err, errclass.ErrNameInvalid, "should reject: %s", name)
	}
}

func TestValidateName_ControlChars(t *testing.T) {
	err := pathutil.ValidateName("hello\x00world")
	require.ErrorIs(t, err, errclass.ErrNameInvalid)
}

func TestValidatePathSafety_UnderRoot(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "worktrees", "test")
	require.NoError(t, os.MkdirAll(target, 0755))
	assert.NoError(t, pathutil.ValidatePathSafety(root, target))
}

func TestValidatePathSafety_Escape(t *testing.T) {
	root := t.TempDir()
	err := pathutil.ValidatePathSafety(root, "/tmp/evil")
	require.ErrorIs(t, err, errclass.ErrPathEscape)
}

func TestValidatePathSafety_SymlinkEscape(t *testing.T) {
	root := t.TempDir()
	link := filepath.Join(root, "escape")
	os.Symlink("/tmp", link)
	err := pathutil.ValidatePathSafety(root, link)
	require.ErrorIs(t, err, errclass.ErrPathEscape)
}

func TestValidatePathSafety_NonExistentTarget(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "worktrees", "new-wt")
	// parent exists, target does not
	require.NoError(t, os.MkdirAll(filepath.Join(root, "worktrees"), 0755))
	assert.NoError(t, pathutil.ValidatePathSafety(root, target))
}
