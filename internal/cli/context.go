package cli

import (
	"fmt"
	"os"

	"github.com/jvs-project/jvs/internal/repo"
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
		fmtErr("not a JVS repository (or any parent): %v", err)
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
	fmt.Fprintf(os.Stderr, "jvs: "+format+"\n", args...)
}
