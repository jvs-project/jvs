// Package template provides snapshot template expansion functionality.
package template

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"
	"time"
)

// Expand expands template placeholders in the input string.
//
// Supported placeholders:
//   {date}      - Current date in YYYY-MM-DD format
//   {time}      - Current time in HH:MM:SS format
//   {datetime}  - Current date and time in YYYY-MM-DD HH:MM:SS format
//   {iso8601}   - Current time in ISO 8601 format
//   {unix}      - Current Unix timestamp
//   {user}      - Current username
//   {hostname}  - System hostname
//   {arch}      - System architecture (e.g., amd64, arm64)
//
// Custom values can be provided via the vars map, which will override
// built-in placeholders.
func Expand(text string, vars map[string]string) string {
	now := time.Now()

	// Built-in placeholders
	placeholders := map[string]string{
		"date":     now.Format("2006-01-02"),
		"time":     now.Format("15:04:05"),
		"datetime": now.Format("2006-01-02 15:04:05"),
		"iso8601":  now.Format(time.RFC3339),
		"unix":     fmt.Sprintf("%d", now.Unix()),
	}

	// Add user info (cached to avoid repeated lookups)
	if u, err := user.Current(); err == nil {
		placeholders["user"] = u.Username
	} else {
		placeholders["user"] = "unknown"
	}

	// Add hostname (cached to avoid repeated lookups)
	if h, err := os.Hostname(); err == nil {
		// Remove domain part if present
		placeholders["hostname"] = strings.Split(h, ".")[0]
	} else {
		placeholders["hostname"] = "unknown"
	}

	// Add architecture
	placeholders["arch"] = runtime.GOARCH

	// Override with custom vars
	for k, v := range vars {
		placeholders[k] = v
	}

	// Replace placeholders
	result := text
	for key, value := range placeholders {
		result = strings.ReplaceAll(result, "{"+key+"}", value)
	}

	return result
}

// ExpandNote is a convenience function for expanding snapshot notes.
func ExpandNote(note string) string {
	return Expand(note, nil)
}
