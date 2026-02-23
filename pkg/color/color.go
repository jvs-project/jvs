// Package color provides terminal color output support for JVS.
// It respects the NO_COLOR environment variable (https://no-color.org/).
package color

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// colorState holds the global color configuration.
var (
	state struct {
		enabled  bool
		once     sync.Once
		disabled bool
	}
)

// Init initializes the color system based on environment and flags.
// It respects the NO_COLOR environment variable (https://no-color.org/)
// and can be disabled programmatically.
func Init(noColorFlag bool) {
	state.once.Do(func() {
		// Check NO_COLOR environment variable
		if _, exists := os.LookupEnv("NO_COLOR"); exists {
			state.disabled = true
		}
		// Check if we're in a dumb terminal
		if term := os.Getenv("TERM"); term == "dumb" {
			state.disabled = true
		}
		// Check explicit flag
		if noColorFlag {
			state.disabled = true
		}
		// Enable colors if not explicitly disabled
		state.enabled = !state.disabled
	})
}

// Enabled returns true if color output is enabled.
func Enabled() bool {
	Init(false) // Ensure initialized
	return state.enabled
}

// Disable turns off color output.
func Disable() {
	state.disabled = true
	state.enabled = false
}

// Enable turns on color output.
func Enable() {
	state.disabled = false
	state.enabled = true
}

// ANSI color codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	DimCode   = "\033[2m"
	Underline = "\033[4m"

	// Foreground colors
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Gray    = "\033[90m"

	// Background colors
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
)

// colorFunc is a function that wraps text with color codes.
type colorFunc func(string) string

// makeColorFunc creates a color function that applies the given color codes.
func makeColorFunc(codes ...string) colorFunc {
	return func(s string) string {
		if !Enabled() {
			return s
		}
		code := strings.Join(codes, "")
		return code + s + Reset
	}
}

// Pre-defined color functions
var (
	Redf     = makeColorFunc(Red)
	Greenf   = makeColorFunc(Green)
	Yellowf  = makeColorFunc(Yellow)
	Bluef    = makeColorFunc(Blue)
	Magentaf = makeColorFunc(Magenta)
	Cyanf    = makeColorFunc(Cyan)
	Whitef   = makeColorFunc(White)
	Grayf    = makeColorFunc(Gray)
	Boldf    = makeColorFunc(Bold)
	Dimf     = makeColorFunc(DimCode)
)

// Specialized formatting functions for common JVS elements

// Success formats a success message in green.
func Success(s string) string {
	return Greenf(s)
}

// Successf formats a success message with printf-style arguments.
func Successf(format string, args ...any) string {
	return Greenf(fmt.Sprintf(format, args...))
}

// Error formats an error message in red.
func Error(s string) string {
	return Redf(s)
}

// Errorf formats an error message with printf-style arguments.
func Errorf(format string, args ...any) string {
	return Redf(fmt.Sprintf(format, args...))
}

// Warning formats a warning message in yellow.
func Warning(s string) string {
	return Yellowf(s)
}

// Warningf formats a warning message with printf-style arguments.
func Warningf(format string, args ...any) string {
	return Yellowf(fmt.Sprintf(format, args...))
}

// Info formats an informational message in cyan.
func Info(s string) string {
	return Cyanf(s)
}

// Infof formats an informational message with printf-style arguments.
func Infof(format string, args ...any) string {
	return Cyanf(fmt.Sprintf(format, args...))
}

// SnapshotID formats a snapshot ID in cyan (for visibility).
func SnapshotID(s string) string {
	return Cyanf(s)
}

// Tag formats a tag name in blue.
func Tag(s string) string {
	return Bluef(s)
}

// Header formats a header in bold.
func Header(s string) string {
	return Boldf(s)
}

// Dim formats dimmed text (for secondary information).
func Dim(s string) string {
	return Dimf(s)
}

// Highlight highlights important text in yellow.
func Highlight(s string) string {
	return Yellowf(s)
}

// Code formats code/command strings in a distinct style (bold + dim).
func Code(s string) string {
	if !Enabled() {
		return s
	}
	return Bold + DimCode + s + Reset
}
