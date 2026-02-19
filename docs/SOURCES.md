# Sources referenced during design

This doc set references upstream docs for key behaviors:

- JuiceFS clone command and options: see JuiceFS command reference and clone guide.
- JuiceFS clone is metadata-only and fast for directories (O(1) w.r.t data size).
- Linux reflink mechanism (FICLONE ioctl) and `cp --reflink` usage on CoW filesystems.
- Btrfs reflink docs.
- JuiceFS sync supports exclude/include filters.

(Links are embedded as citations in the design review response, not repeated here.)
