// Package pathutil provides path and name validation utilities for JVS.
package pathutil

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"

	"github.com/jvs-project/jvs/pkg/errclass"
)

var nameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// ValidateTag validates a tag string (same rules as worktree names).
func ValidateTag(tag string) error {
	if tag == "" {
		return errclass.ErrNameInvalid.WithMessage("tag must not be empty")
	}
	if !nameRegex.MatchString(tag) {
		return errclass.ErrNameInvalid.WithMessagef("tag must match [a-zA-Z0-9._-]+: %s", tag)
	}
	return nil
}

// ValidateName checks worktree/ref name safety per spec 02/03.
func ValidateName(name string) error {
	if name == "" {
		return errclass.ErrNameInvalid.WithMessage("name must not be empty")
	}

	// NFC normalize
	name = norm.NFC.String(name)

	if name == ".." || strings.Contains(name, "..") {
		return errclass.ErrNameInvalid.WithMessagef("name must not contain '..': %s", name)
	}

	if strings.ContainsAny(name, "/\\") {
		return errclass.ErrNameInvalid.WithMessagef("name must not contain separators: %s", name)
	}

	// Check for control characters
	for _, r := range name {
		if unicode.IsControl(r) {
			return errclass.ErrNameInvalid.WithMessagef("name must not contain control characters: %q", name)
		}
	}

	if !nameRegex.MatchString(name) {
		return errclass.ErrNameInvalid.WithMessagef("name must match [a-zA-Z0-9._-]+: %s", name)
	}

	return nil
}

// ValidatePathSafety verifies target path does not escape repo root.
func ValidatePathSafety(repoRoot, targetPath string) error {
	// Resolve repo root symlinks
	resolvedRoot, err := filepath.EvalSymlinks(repoRoot)
	if err != nil {
		return errclass.ErrPathEscape.WithMessagef("cannot resolve repo root: %v", err)
	}

	// Try resolving target; if it doesn't exist, resolve closest ancestor
	resolvedTarget, err := filepath.EvalSymlinks(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			resolvedTarget = resolveClosestAncestor(targetPath)
		} else {
			return errclass.ErrPathEscape.WithMessagef("cannot resolve target: %v", err)
		}
	}

	// Ensure resolved target is under resolved root
	if !strings.HasPrefix(resolvedTarget+"/", resolvedRoot+"/") &&
		resolvedTarget != resolvedRoot {
		return errclass.ErrPathEscape.WithMessagef("path escapes repo root: %s", targetPath)
	}

	return nil
}

// resolveClosestAncestor walks up from path to find the closest existing
// ancestor, resolves it, then appends the remaining components.
func resolveClosestAncestor(path string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		if os.IsNotExist(err) {
			// Recurse up
			resolved = resolveClosestAncestor(dir)
		} else {
			return filepath.Clean(path)
		}
	}
	return filepath.Join(resolved, base)
}
