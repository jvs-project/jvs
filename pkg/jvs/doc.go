// Package jvs provides a high-level library API for JVS (Juicy Versioned Workspaces).
//
// This package is the primary integration point for external consumers such as
// sandbox-manager. It wraps internal packages into a clean, stable public API.
//
// # Concurrency Safety
//
// JVS operations are filesystem-based and follow these concurrency rules:
//
//   - Snapshot() is safe when no concurrent writes to the payload directory.
//     Always snapshot AFTER the agent pod has been deleted (process stopped).
//
//   - Restore() / RestoreLatest() is safe when no concurrent reads from the
//     payload directory. Always restore BEFORE the agent pod is created.
//
//   - Multiple Client instances for DIFFERENT repositories are fully independent
//     and safe to use concurrently.
//
//   - Multiple Client instances for the SAME repository must NOT call
//     mutating operations (Snapshot, Restore, GC) concurrently.
//
// # Recommended Usage Pattern (sandbox-manager)
//
//	// Pod startup: restore workspace before creating pod
//	client, err := jvs.OpenOrInit(repoPath, jvs.InitOptions{Name: "agent-ws"})
//	payloadPath := client.WorktreePayloadPath("main")
//	if has, _ := client.HasSnapshots(ctx, "main"); has {
//	    client.RestoreLatest(ctx, "main")
//	}
//	// Mount payloadPath as /workspace in pod via JuiceFS subPath
//
//	// Pod shutdown: snapshot after pod is deleted
//	client.Snapshot(ctx, jvs.SnapshotOptions{
//	    Note: "auto: pod shutdown",
//	    Tags: []string{"auto", "shutdown"},
//	})
package jvs
