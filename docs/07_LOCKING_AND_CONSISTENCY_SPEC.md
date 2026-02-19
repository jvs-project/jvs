# Locking & Consistency Spec (v6.3)

## Goals
- Enforce SWMR for `exclusive` worktrees.
- Prevent stale-holder writes via fencing.
- Keep deterministic behavior under bounded clock skew.

## Lock scope
- Locks are repo-local files in `.jvs/locks/`.
- One active writer lock per exclusive worktree.
- `shared` mode does not provide SWMR guarantees.

## Lock record schema (MUST)
- `lock_id`
- `worktree_id`
- `holder_id` (`host:user:pid:start_time`)
- `holder_nonce`
- `session_id`
- `acquire_seq`
- `created_at`
- `last_renewed_at`
- `lease_duration_ms`
- `renew_interval_ms`
- `max_clock_skew_ms`
- `steal_grace_ms`
- `lease_expires_at`
- `fencing_token`

## Default policy values
- `lease_duration_ms = 30000`
- `renew_interval_ms = 10000`
- `max_clock_skew_ms = 2000`
- `steal_grace_ms = 1000`

## Protocol (MUST)
### Acquire
- create lock file atomically using `O_CREAT|O_EXCL` on `.jvs/locks/<worktree_id>.lock`
- if `open()` returns `EEXIST`, read existing lock and evaluate expiry:
  - if active non-expired lock exists, return `E_LOCK_CONFLICT`
  - if expired plus skew+grace, steal flow applies and increments fencing token
- after successful create, fsync lock file and parent directory
- on JuiceFS, `O_CREAT|O_EXCL` provides the required atomicity guarantee via the metadata engine

### Renew
- only holder with matching `holder_nonce` and `session_id` may renew
- renewal extends lease by `lease_duration_ms`
- if renewal commit fails, holder must stop writes immediately

### Steal
- allowed only when:
  `now > lease_expires_at + max_clock_skew_ms + steal_grace_ms`
- steal MUST use atomic file replacement:
  1. write new lock record to `.jvs/locks/<worktree_id>.lock.tmp`; fsync.
  2. `rename()` tmp over existing `.lock` file; fsync parent dir.
  3. `rename()` is atomic on POSIX and JuiceFS metadata engine.
- concurrent stealer race: if `rename()` succeeds, that stealer wins; the loser's tmp is orphaned and harmless (cleaned by `doctor`).
- new holder increments fencing token and `acquire_seq`
- audit event is mandatory

### Release
- only holder with matching nonce/session can release
- non-holder release fails with `E_LOCK_NOT_HELD`

## Fencing token rules (MUST)
- mutating writes in exclusive mode must validate current token before publish commit
- stale token fails with `E_FENCING_MISMATCH`
- no partial publish after fencing failure

## Snapshot consistency levels
### `quiesced` (default)
- requires quiesced source window
- in exclusive mode, holder ensures no concurrent payload writers

### `best_effort`
- snapshot allowed without strict quiesce
- descriptor carries risk label
- `history`/`info` JSON exposes risk flag

## READY visibility
Only READY snapshots are visible.
Incomplete snapshots/intents are hidden and recoverable by `doctor --strict`.

## Clock skew detection (MUST)
- On lock acquire and renew, JVS records `last_renewed_at` using local monotonic + wall clock.
- On steal evaluation, compute `observed_skew = abs(local_now - lock.lease_expires_at - lock.lease_duration_ms)`.
- If `observed_skew > max_clock_skew_ms`, fail with `E_CLOCK_SKEW_EXCEEDED` and refuse steal.
- Operator MUST resolve clock drift before retrying.
- `jvs doctor --strict` SHOULD warn if system clock offset exceeds `max_clock_skew_ms / 2`.

## Error classes
`E_LOCK_CONFLICT`, `E_LOCK_EXPIRED`, `E_LOCK_NOT_HELD`, `E_FENCING_MISMATCH`, `E_CLOCK_SKEW_EXCEEDED`, `E_CONSISTENCY_UNAVAILABLE`, `E_PARTIAL_SNAPSHOT`.
