package doctor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jvs-project/jvs/internal/repo"
	"github.com/jvs-project/jvs/internal/verify"
	"github.com/jvs-project/jvs/internal/worktree"
	"github.com/jvs-project/jvs/pkg/model"
)

// Finding represents a detected issue.
type Finding struct {
	Category    string `json:"category"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	ErrorCode   string `json:"error_code,omitempty"`
	Path        string `json:"path,omitempty"`
}

// Result contains doctor check results.
type Result struct {
	Healthy  bool      `json:"healthy"`
	Findings []Finding `json:"findings"`
}

// Doctor performs repository health checks.
type Doctor struct {
	repoRoot string
}

// NewDoctor creates a new doctor.
func NewDoctor(repoRoot string) *Doctor {
	return &Doctor{repoRoot: repoRoot}
}

// Check runs all diagnostic checks.
func (d *Doctor) Check(strict bool) (*Result, error) {
	result := &Result{Healthy: true}

	// 1. Check format version
	d.checkFormatVersion(result)

	// 2. Check worktrees
	d.checkWorktrees(result)

	// 3. Check for orphan intents
	d.checkOrphanIntents(result)

	// 4. Check snapshot integrity (if strict)
	if strict {
		d.checkSnapshotIntegrity(result)
		// 5. Check audit chain (if strict)
		d.checkAuditChain(result)
	}

	// 6. Check for orphan tmp files
	d.checkOrphanTmp(result)

	return result, nil
}

func (d *Doctor) checkFormatVersion(result *Result) {
	versionPath := filepath.Join(d.repoRoot, ".jvs", "format_version")
	data, err := os.ReadFile(versionPath)
	if err != nil {
		result.Findings = append(result.Findings, Finding{
			Category:    "format",
			Description: "format_version file missing or unreadable",
			Severity:    "critical",
			Path:        versionPath,
		})
		result.Healthy = false
		return
	}

	var version int
	fmt.Sscanf(string(data), "%d", &version)
	if version > repo.FormatVersion {
		result.Findings = append(result.Findings, Finding{
			Category:    "format",
			Description: fmt.Sprintf("format version %d > supported %d", version, repo.FormatVersion),
			Severity:    "critical",
		})
		result.Healthy = false
	}
}

func (d *Doctor) checkWorktrees(result *Result) {
	wtMgr := worktree.NewManager(d.repoRoot)
	list, err := wtMgr.List()
	if err != nil {
		result.Findings = append(result.Findings, Finding{
			Category:    "worktree",
			Description: fmt.Sprintf("cannot list worktrees: %v", err),
			Severity:    "error",
		})
		return
	}

	for _, cfg := range list {
		// Check payload directory exists
		payloadPath := wtMgr.Path(cfg.Name)
		if _, err := os.Stat(payloadPath); os.IsNotExist(err) {
			result.Findings = append(result.Findings, Finding{
				Category:    "worktree",
				Description: fmt.Sprintf("worktree '%s' payload directory missing", cfg.Name),
				Severity:    "error",
				Path:        payloadPath,
			})
		}

		// Check head snapshot exists
		if cfg.HeadSnapshotID != "" {
			descPath := filepath.Join(d.repoRoot, ".jvs", "descriptors", string(cfg.HeadSnapshotID)+".json")
			if _, err := os.Stat(descPath); os.IsNotExist(err) {
				result.Findings = append(result.Findings, Finding{
					Category:    "worktree",
					Description: fmt.Sprintf("worktree '%s' head snapshot %s not found", cfg.Name, cfg.HeadSnapshotID),
					Severity:    "warning",
				})
			}
		}
	}
}

func (d *Doctor) checkOrphanIntents(result *Result) {
	intentsDir := filepath.Join(d.repoRoot, ".jvs", "intents")
	entries, err := os.ReadDir(intentsDir)
	if err != nil {
		return // directory doesn't exist, that's fine
	}

	for _, entry := range entries {
		result.Findings = append(result.Findings, Finding{
			Category:    "intent",
			Description: fmt.Sprintf("orphan intent file: %s", entry.Name()),
			Severity:    "warning",
			Path:        filepath.Join(intentsDir, entry.Name()),
		})
	}
}

func (d *Doctor) checkSnapshotIntegrity(result *Result) {
	verifier := verify.NewVerifier(d.repoRoot)
	results, err := verifier.VerifyAll(true)
	if err != nil {
		result.Findings = append(result.Findings, Finding{
			Category:    "integrity",
			Description: fmt.Sprintf("verification failed: %v", err),
			Severity:    "error",
		})
		return
	}

	for _, r := range results {
		if r.TamperDetected {
			result.Findings = append(result.Findings, Finding{
				Category:    "integrity",
				Description: fmt.Sprintf("snapshot %s: %s", r.SnapshotID, r.Error),
				Severity:    "critical",
			})
			result.Healthy = false
		}
	}
}

func (d *Doctor) checkOrphanTmp(result *Result) {
	// Check for orphan .jvs-tmp-* files
	filepath.Walk(d.repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		name := info.Name()
		if len(name) > 9 && name[:9] == ".jvs-tmp-" {
			result.Findings = append(result.Findings, Finding{
				Category:    "tmp",
				Description: fmt.Sprintf("orphan temp file: %s", name),
				Severity:    "info",
				Path:        path,
			})
		}
		return nil
	})

	// Check for orphan snapshot .tmp directories
	snapshotsDir := filepath.Join(d.repoRoot, ".jvs", "snapshots")
	entries, err := os.ReadDir(snapshotsDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() && filepath.Ext(entry.Name()) == ".tmp" {
				result.Findings = append(result.Findings, Finding{
					Category:    "tmp",
					Description: fmt.Sprintf("orphan snapshot tmp directory: %s", entry.Name()),
					Severity:    "warning",
					Path:        filepath.Join(snapshotsDir, entry.Name()),
				})
			}
		}
	}
}

// checkAuditChain verifies the audit log hash chain integrity.
func (d *Doctor) checkAuditChain(result *Result) {
	auditPath := filepath.Join(d.repoRoot, ".jvs", "audit", "audit.jsonl")
	file, err := os.Open(auditPath)
	if os.IsNotExist(err) {
		return // No audit log yet is OK
	}
	if err != nil {
		result.Findings = append(result.Findings, Finding{
			Category:    "audit",
			Description: fmt.Sprintf("cannot open audit log: %v", err),
			Severity:    "warning",
			Path:        auditPath,
		})
		return
	}
	defer file.Close()

	var prevHash model.HashValue
	var lineNum int
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineNum++
		var record model.AuditRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			result.Findings = append(result.Findings, Finding{
				Category:    "audit",
				Description: fmt.Sprintf("malformed record at line %d", lineNum),
				Severity:    "warning",
				Path:        auditPath,
			})
			continue
		}

		// Verify chain linkage (skip first record which has empty prevHash)
		if prevHash != "" && record.PrevHash != prevHash {
			result.Findings = append(result.Findings, Finding{
				Category:    "audit",
				Description: fmt.Sprintf("audit hash chain broken at line %d", lineNum),
				Severity:    "critical",
				ErrorCode:   "E_AUDIT_CHAIN_BROKEN",
				Path:        auditPath,
			})
			result.Healthy = false
			return
		}
		prevHash = record.RecordHash
	}

	if err := scanner.Err(); err != nil {
		result.Findings = append(result.Findings, Finding{
			Category:    "audit",
			Description: fmt.Sprintf("error reading audit log: %v", err),
			Severity:    "warning",
			Path:        auditPath,
		})
	}
}
