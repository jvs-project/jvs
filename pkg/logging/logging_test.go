package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger(LevelInfo)
	if logger.level != LevelInfo {
		t.Errorf("expected level %s, got %s", LevelInfo, logger.level)
	}
}

func TestLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelDebug)
	logger.SetOutput(&buf)

	logger.Debug("test message", map[string]any{"key": "value"})

	output := buf.String()
	if !strings.Contains(output, `"level":"debug"`) {
		t.Errorf("expected debug level in output, got: %s", output)
	}
	if !strings.Contains(output, `"message":"test message"`) {
		t.Errorf("expected message in output, got: %s", output)
	}
}

func TestLogger_DebugFiltered(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf)

	logger.Debug("test message")

	if buf.Len() > 0 {
		t.Errorf("expected no output for debug when level is info, got: %s", buf.String())
	}
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf)

	logger.Info("info message")

	output := buf.String()
	if !strings.Contains(output, `"level":"info"`) {
		t.Errorf("expected info level in output, got: %s", output)
	}
}

func TestLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf)

	logger.Warn("warn message")

	output := buf.String()
	if !strings.Contains(output, `"level":"warn"`) {
		t.Errorf("expected warn level in output, got: %s", output)
	}
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelError)
	logger.SetOutput(&buf)

	logger.Error("error message")

	output := buf.String()
	if !strings.Contains(output, `"level":"error"`) {
		t.Errorf("expected error level in output, got: %s", output)
	}
}

func TestLogger_ErrorErr(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelError)
	logger.SetOutput(&buf)

	logger.ErrorErr("operation failed", assertError("test error"))

	output := buf.String()
	if !strings.Contains(output, `"error":"test error"`) {
		t.Errorf("expected error field in output, got: %s", output)
	}
}

func TestLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf)

	loggerWithFields := logger.WithFields(map[string]any{"request_id": "123"})
	loggerWithFields.Info("test message")

	output := buf.String()
	if !strings.Contains(output, `"request_id":"123"`) {
		t.Errorf("expected request_id field in output, got: %s", output)
	}
}

func TestLogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf)

	logger.Info("test message", map[string]any{"count": 42})

	output := buf.String()

	// Verify it's valid JSON
	var entry LogEntry
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Errorf("output is not valid JSON: %v, got: %s", err, output)
	}

	if entry.Message != "test message" {
		t.Errorf("expected message 'test message', got: %s", entry.Message)
	}
	if entry.Fields["count"].(float64) != 42 {
		t.Errorf("expected count 42, got: %v", entry.Fields["count"])
	}
}

func TestLogger_SetLevel(t *testing.T) {
	logger := NewLogger(LevelError)
	logger.SetLevel(LevelDebug)

	if logger.level != LevelDebug {
		t.Errorf("expected level %s, got %s", LevelDebug, logger.level)
	}
}

func TestGlobalLogger(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewLogger(LevelDebug)
	testLogger.SetOutput(&buf)
	SetGlobal(testLogger)

	Debug("global debug message")

	output := buf.String()
	if !strings.Contains(output, `"message":"global debug message"`) {
		t.Errorf("expected global message in output, got: %s", output)
	}
}

func TestWithFields_Global(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewLogger(LevelInfo)
	testLogger.SetOutput(&buf)
	SetGlobal(testLogger)

	logger := WithFields(map[string]any{"component": "test"})
	logger.Info("component message")

	output := buf.String()
	if !strings.Contains(output, `"component":"test"`) {
		t.Errorf("expected component field in output, got: %s", output)
	}
}

func assertError(msg string) error {
	return &testError{msg: msg}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
