package cli

import (
	"fmt"
	"os"

	"github.com/jvs-project/jvs/internal/repo"
	"github.com/jvs-project/jvs/pkg/color"
)

// requireRepo discovers the repo from CWD and returns it, or exits with error.
func requireRepo() *repo.Repo {
	cwd, err := os.Getwd()
	if err != nil {
		fmtErr("cannot get current directory: %v", err)
		os.Exit(1)
	}
	r, err := repo.Discover(cwd)
	if err != nil {
		// Enhanced error message with suggestion
		fmt.Fprintln(os.Stderr, formatNotInRepositoryError())
		os.Exit(1)
	}
	return r
}

// requireWorktree discovers the repo and worktree from CWD, or exits with error.
func requireWorktree() (*repo.Repo, string) {
	cwd, err := os.Getwd()
	if err != nil {
		fmtErr("cannot get current directory: %v", err)
		os.Exit(1)
	}
	r, wtName, err := repo.DiscoverWorktree(cwd)
	if err != nil {
		fmtErr("not a JVS repository: %v", err)
		os.Exit(1)
	}
	if wtName == "" {
		fmtErr("not inside a worktree (current directory is not under main/ or worktrees/)")
		os.Exit(1)
	}
	return r, wtName
}

func fmtErr(format string, args ...any) {
	// Colorize the error prefix
	prefix := "jvs: "
	if color.Enabled() {
		prefix = color.Error("jvs:") + " "
	}
	fmt.Fprintf(os.Stderr, prefix+format+"\n", args...)
}

