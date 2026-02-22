package logging

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGlobal_Info tests the global Info function.
func TestGlobal_Info(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewLogger(LevelInfo)
	testLogger.SetOutput(&buf)
	SetGlobal(testLogger)

	Info("global info message")

	output := buf.String()
	assert.Contains(t, output, `"level":"info"`)
	assert.Contains(t, output, `"message":"global info message"`)
}

// TestGlobal_Warn tests the global Warn function.
func TestGlobal_Warn(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewLogger(LevelInfo)
	testLogger.SetOutput(&buf)
	SetGlobal(testLogger)

	Warn("global warn message")

	output := buf.String()
	assert.Contains(t, output, `"level":"warn"`)
	assert.Contains(t, output, `"message":"global warn message"`)
}

// TestGlobal_Error tests the global Error function.
func TestGlobal_Error(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewLogger(LevelError)
	testLogger.SetOutput(&buf)
	SetGlobal(testLogger)

	Error("global error message")

	output := buf.String()
	assert.Contains(t, output, `"level":"error"`)
	assert.Contains(t, output, `"message":"global error message"`)
}

// TestGlobal_ErrorErr tests the global ErrorErr function.
func TestGlobal_ErrorErr(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewLogger(LevelError)
	testLogger.SetOutput(&buf)
	SetGlobal(testLogger)

	testErr := errors.New("test error")
	ErrorErr("operation failed", testErr)

	output := buf.String()
	assert.Contains(t, output, `"level":"error"`)
	assert.Contains(t, output, `"message":"operation failed"`)
	assert.Contains(t, output, `"error":"test error"`)
}

// TestGlobal_Debug tests the global Debug function.
func TestGlobal_Debug(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewLogger(LevelDebug)
	testLogger.SetOutput(&buf)
	SetGlobal(testLogger)

	Debug("global debug message")

	output := buf.String()
	assert.Contains(t, output, `"level":"debug"`)
	assert.Contains(t, output, `"message":"global debug message"`)
}

// TestLogger_ErrorErr_WithAdditionalFields tests ErrorErr with extra fields.
func TestLogger_ErrorErr_WithAdditionalFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelError)
	logger.SetOutput(&buf)

	testErr := errors.New("database error")
	logger.ErrorErr("query failed", testErr, map[string]any{"query": "SELECT *", "retry": 3})

	output := buf.String()
	var entry LogEntry
	require.NoError(t, json.Unmarshal([]byte(output), &entry))

	assert.Equal(t, "query failed", entry.Message)
	assert.Equal(t, "database error", entry.Fields["error"])
	assert.Equal(t, "SELECT *", entry.Fields["query"])
	assert.Equal(t, float64(3), entry.Fields["retry"])
}

// TestLogger_ErrorErr_MultipleFieldMaps tests ErrorErr with multiple field maps.
func TestLogger_ErrorErr_MultipleFieldMaps(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelError)
	logger.SetOutput(&buf)

	testErr := errors.New("multi error")
	logger.ErrorErr("multi fields", testErr,
		map[string]any{"field1": "value1"},
		map[string]any{"field2": "value2"})

	output := buf.String()
	var entry LogEntry
	require.NoError(t, json.Unmarshal([]byte(output), &entry))

	assert.Equal(t, "multi error", entry.Fields["error"])
	assert.Equal(t, "value1", entry.Fields["field1"])
	assert.Equal(t, "value2", entry.Fields["field2"])
}

// TestLogger_WithFields_MultipleCalls tests chaining WithFields.
func TestLogger_WithFields_MultipleCalls(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf)

	logger1 := logger.WithFields(map[string]any{"base": "field"})
	logger2 := logger1.WithFields(map[string]any{"extra": "data"})
	logger2.Info("chained message")

	output := buf.String()
	var entry LogEntry
	require.NoError(t, json.Unmarshal([]byte(output), &entry))

	assert.Equal(t, "field", entry.Fields["base"])
	assert.Equal(t, "data", entry.Fields["extra"])
}

// TestLogger_WithFields_OriginalUnmodified tests that original logger is not modified.
func TestLogger_WithFields_OriginalUnmodified(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf)

	logger.WithFields(map[string]any{"temp": "value"}).Info("temp message")

	// Original logger should not have the temp field
	buf.Reset()
	logger.Info("original message")

	output := buf.String()
	assert.NotContains(t, output, `"temp":"value"`)
}

// TestLogger_WithFields_Overwrite tests field overwriting behavior.
func TestLogger_WithFields_Overwrite(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf)

	logger1 := logger.WithFields(map[string]any{"key": "original"})
	logger2 := logger1.WithFields(map[string]any{"key": "overwritten"})

	logger2.Info("test")

	output := buf.String()
	var entry LogEntry
	require.NoError(t, json.Unmarshal([]byte(output), &entry))

	assert.Equal(t, "overwritten", entry.Fields["key"])
}

// TestLogger_Info_WithFields tests Info with additional fields.
func TestLogger_Info_WithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf)

	logger.Info("info with fields", map[string]any{"user": "alice", "action": "login"})

	output := buf.String()
	var entry LogEntry
	require.NoError(t, json.Unmarshal([]byte(output), &entry))

	assert.Equal(t, "info with fields", entry.Message)
	assert.Equal(t, "alice", entry.Fields["user"])
	assert.Equal(t, "login", entry.Fields["action"])
}

// TestLogger_Debug_WithFields tests Debug with additional fields.
func TestLogger_Debug_WithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelDebug)
	logger.SetOutput(&buf)

	logger.Debug("debug with fields", map[string]any{"trace_id": "abc123"})

	output := buf.String()
	var entry LogEntry
	require.NoError(t, json.Unmarshal([]byte(output), &entry))

	assert.Equal(t, "debug with fields", entry.Message)
	assert.Equal(t, "abc123", entry.Fields["trace_id"])
}

// TestLogger_Warn_WithFields tests Warn with additional fields.
func TestLogger_Warn_WithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf)

	logger.Warn("warn with fields", map[string]any{"warning": "deprecated"})

	output := buf.String()
	var entry LogEntry
	require.NoError(t, json.Unmarshal([]byte(output), &entry))

	assert.Equal(t, "warn with fields", entry.Message)
	assert.Equal(t, "deprecated", entry.Fields["warning"])
}

// TestLogger_Error_WithFields tests Error with additional fields.
func TestLogger_Error_WithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelError)
	logger.SetOutput(&buf)

	logger.Error("error with fields", map[string]any{"code": 500, "path": "/api"})

	output := buf.String()
	var entry LogEntry
	require.NoError(t, json.Unmarshal([]byte(output), &entry))

	assert.Equal(t, "error with fields", entry.Message)
	assert.Equal(t, float64(500), entry.Fields["code"])
	assert.Equal(t, "/api", entry.Fields["path"])
}

// TestLogger_WithFields_ThenInfo tests WithFields followed by Info.
func TestLogger_WithFields_ThenInfo(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf)

	logger.WithFields(map[string]any{"service": "api"}).Info("service started")

	output := buf.String()
	var entry LogEntry
	require.NoError(t, json.Unmarshal([]byte(output), &entry))

	assert.Equal(t, "service started", entry.Message)
	assert.Equal(t, "api", entry.Fields["service"])
}

// TestLogger_WithFields_ThenDebug tests WithFields followed by Debug.
func TestLogger_WithFields_ThenDebug(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelDebug)
	logger.SetOutput(&buf)

	logger.WithFields(map[string]any{"component": "cache"}).Debug("cache miss")

	output := buf.String()
	var entry LogEntry
	require.NoError(t, json.Unmarshal([]byte(output), &entry))

	assert.Equal(t, "cache miss", entry.Message)
	assert.Equal(t, "cache", entry.Fields["component"])
}

// TestLogger_WithFields_ThenWarn tests WithFields followed by Warn.
func TestLogger_WithFields_ThenWarn(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelWarn)
	logger.SetOutput(&buf)

	logger.WithFields(map[string]any{"resource": "memory"}).Warn("high usage")

	output := buf.String()
	var entry LogEntry
	require.NoError(t, json.Unmarshal([]byte(output), &entry))

	assert.Equal(t, "high usage", entry.Message)
	assert.Equal(t, "memory", entry.Fields["resource"])
}

// TestLogger_WithFields_ThenError tests WithFields followed by Error.
func TestLogger_WithFields_ThenError(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelError)
	logger.SetOutput(&buf)

	logger.WithFields(map[string]any{"module": "database"}).Error("connection failed")

	output := buf.String()
	var entry LogEntry
	require.NoError(t, json.Unmarshal([]byte(output), &entry))

	assert.Equal(t, "connection failed", entry.Message)
	assert.Equal(t, "database", entry.Fields["module"])
}

// TestLogger_Log_EmptyFields tests log output when no fields are provided.
func TestLogger_Log_EmptyFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf)

	logger.Info("message without fields")

	output := buf.String()
	var entry LogEntry
	require.NoError(t, json.Unmarshal([]byte(output), &entry))

	// Fields should be null/empty when no fields provided
	assert.Nil(t, entry.Fields)
}

// TestLogger_Log_BaseFields tests that base fields from WithFields are included.
func TestLogger_Log_BaseFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf)

	loggerWithBase := logger.WithFields(map[string]any{"env": "prod", "region": "us-east"})
	loggerWithBase.Info("operational message")

	output := buf.String()
	var entry LogEntry
	require.NoError(t, json.Unmarshal([]byte(output), &entry))

	assert.Equal(t, "prod", entry.Fields["env"])
	assert.Equal(t, "us-east", entry.Fields["region"])
}

// TestLogger_Log_TimestampFormat tests that timestamps are in RFC3339Nano format.
func TestLogger_Log_TimestampFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf)

	logger.Info("timestamp test")

	output := buf.String()
	var entry LogEntry
	require.NoError(t, json.Unmarshal([]byte(output), &entry))

	// RFC3339Nano timestamps contain 'T' and end with 'Z' or have timezone
	assert.Contains(t, entry.Timestamp, "T")
	assert.True(t, strings.HasSuffix(entry.Timestamp, "Z") || strings.Contains(entry.Timestamp, "+"))
}

// TestLogger_LevelFiltering_DebugLevel tests all log levels at Debug level.
func TestLogger_LevelFiltering_DebugLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelDebug)
	logger.SetOutput(&buf)

	logger.Debug("debug msg")
	logger.Info("info msg")
	logger.Warn("warn msg")
	logger.Error("error msg")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Len(t, lines, 4, "all 4 messages should be logged at debug level")
}

// TestLogger_LevelFiltering_InfoLevel tests log filtering at Info level.
func TestLogger_LevelFiltering_InfoLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf)

	logger.Debug("debug msg")
	logger.Info("info msg")
	logger.Warn("warn msg")
	logger.Error("error msg")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Debug should be filtered, others should log
	assert.Len(t, lines, 3, "info, warn, error should be logged at info level")
	assert.NotContains(t, output, "debug msg")
	assert.Contains(t, output, "info msg")
	assert.Contains(t, output, "warn msg")
	assert.Contains(t, output, "error msg")
}

// TestLogger_LevelFiltering_WarnLevel tests log filtering at Warn level.
func TestLogger_LevelFiltering_WarnLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelWarn)
	logger.SetOutput(&buf)

	logger.Debug("debug msg")
	logger.Info("info msg")
	logger.Warn("warn msg")
	logger.Error("error msg")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	assert.Len(t, lines, 2, "warn and error should be logged at warn level")
	assert.NotContains(t, output, "debug msg")
	assert.NotContains(t, output, "info msg")
	assert.Contains(t, output, "warn msg")
	assert.Contains(t, output, "error msg")
}

// TestLogger_LevelFiltering_ErrorLevel tests log filtering at Error level.
func TestLogger_LevelFiltering_ErrorLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelError)
	logger.SetOutput(&buf)

	logger.Debug("debug msg")
	logger.Info("info msg")
	logger.Warn("warn msg")
	logger.Error("error msg")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	assert.Len(t, lines, 1, "only error should be logged at error level")
	assert.Contains(t, output, "error msg")
	assert.NotContains(t, output, "debug msg")
	assert.NotContains(t, output, "info msg")
	assert.NotContains(t, output, "warn msg")
}

// TestLogger_SetOutput_DuringLogging tests changing output mid-stream.
func TestLogger_SetOutput_DuringLogging(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf1)

	logger.Info("first message")

	logger.SetOutput(&buf2)
	logger.Info("second message")

	assert.Contains(t, buf1.String(), "first message")
	assert.NotContains(t, buf1.String(), "second message")
	assert.Contains(t, buf2.String(), "second message")
	assert.NotContains(t, buf2.String(), "first message")
}

// TestLogger_SetLevel_DuringLogging tests changing level mid-stream.
func TestLogger_SetLevel_DuringLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelError)
	logger.SetOutput(&buf)

	logger.Debug("debug1")
	logger.Info("info1")
	logger.Warn("warn1")

	logger.SetLevel(LevelDebug)

	logger.Debug("debug2")
	logger.Info("info2")
	logger.Warn("warn2")

	output := buf.String()
	assert.NotContains(t, output, "debug1")
	assert.NotContains(t, output, "info1")
	assert.NotContains(t, output, "warn1")
	assert.Contains(t, output, "debug2")
	assert.Contains(t, output, "info2")
	assert.Contains(t, output, "warn2")
}

// TestGlobal_AllLevels tests all global log functions.
func TestGlobal_AllLevels(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewLogger(LevelDebug)
	testLogger.SetOutput(&buf)
	SetGlobal(testLogger)

	Debug("debug")
	Info("info")
	Warn("warn")
	Error("error")

	output := buf.String()
	assert.Contains(t, output, `"level":"debug"`)
	assert.Contains(t, output, `"level":"info"`)
	assert.Contains(t, output, `"level":"warn"`)
	assert.Contains(t, output, `"level":"error"`)
}

// TestGlobal_WithFields tests global WithFields function.
func TestGlobal_WithFields(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewLogger(LevelDebug)
	testLogger.SetOutput(&buf)
	SetGlobal(testLogger)

	logger := WithFields(map[string]any{"global": "field"})
	logger.Debug("message with global field")

	output := buf.String()
	var entry LogEntry
	require.NoError(t, json.Unmarshal([]byte(output), &entry))

	assert.Equal(t, "field", entry.Fields["global"])
	assert.Equal(t, "message with global field", entry.Message)
}

// TestLogger_ErrorErr_NilError_Panics tests that ErrorErr panics with nil error.
// Note: This documents current behavior - ErrorErr assumes non-nil error.
func TestLogger_ErrorErr_NilError_Panics(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelError)
	logger.SetOutput(&buf)

	assert.Panics(t, func() {
		logger.ErrorErr("no error", nil)
	})
}

// TestLogger_JSONOutput_ContainsTimestamp tests that JSON output contains timestamp.
func TestLogger_JSONOutput_ContainsTimestamp(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf)

	logger.Info("test")

	output := buf.String()
	var entry LogEntry
	require.NoError(t, json.Unmarshal([]byte(output), &entry))

	assert.NotEmpty(t, entry.Timestamp)
}

// TestLogger_MultipleFieldMaps tests logging with multiple field maps.
func TestLogger_MultipleFieldMaps(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf)

	logger.Info("multi map",
		map[string]any{"first": 1},
		map[string]any{"second": 2})

	output := buf.String()
	var entry LogEntry
	require.NoError(t, json.Unmarshal([]byte(output), &entry))

	assert.Equal(t, float64(1), entry.Fields["first"])
	assert.Equal(t, float64(2), entry.Fields["second"])
}

// TestLogger_ConcurrentLogging tests thread-safe concurrent logging.
func TestLogger_ConcurrentLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf)

	done := make(chan bool)
	for i := 0; i < 50; i++ {
		go func(n int) {
			logger.Info("message", map[string]any{"id": n})
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 50; i++ {
		<-done
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Len(t, lines, 50)
}

// TestNewLogger_Defaults tests NewLogger creates proper defaults.
func TestNewLogger_Defaults(t *testing.T) {
	logger := NewLogger(LevelWarn)

	assert.Equal(t, LevelWarn, logger.level)
	assert.NotNil(t, logger.fields)
	assert.NotNil(t, logger.output)
}

// TestLogger_AllLevelConstants tests all level constants.
func TestLogger_AllLevelConstants(t *testing.T) {
	tests := []struct {
		name        string
		loggerLevel Level
		logFunc     func(*Logger, string)
		expectedLog Level
	}{
		{"debug", LevelDebug, func(l *Logger, msg string) { l.Debug(msg) }, LevelDebug},
		{"info", LevelInfo, func(l *Logger, msg string) { l.Info(msg) }, LevelInfo},
		{"warn", LevelWarn, func(l *Logger, msg string) { l.Warn(msg) }, LevelWarn},
		{"error", LevelError, func(l *Logger, msg string) { l.Error(msg) }, LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(tt.loggerLevel)
			logger.SetOutput(&buf)
			tt.logFunc(logger, "test")

			var entry LogEntry
			require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
			assert.Equal(t, tt.expectedLog, entry.Level)
		})
	}
}

// TestLogger_SetOutput_ReturnsToDefault tests that SetOutput persists.
func TestLogger_SetOutput_Persists(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)

	logger.SetOutput(&buf)
	logger.Info("test1")

	// Output should still be buf
	logger.Info("test2")

	output := buf.String()
	assert.Contains(t, output, "test1")
	assert.Contains(t, output, "test2")
}

// TestLogger_WithFields_EmptyMap tests WithFields with empty map.
func TestLogger_WithFields_EmptyMap(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf)

	loggerWithFields := logger.WithFields(map[string]any{})
	loggerWithFields.Info("test")

	output := buf.String()
	var entry LogEntry
	require.NoError(t, json.Unmarshal([]byte(output), &entry))

	assert.Nil(t, entry.Fields)
}

// TestLogger_JSONMarshalError tests the error path when JSON marshaling fails.
// This is difficult to trigger in practice since json.Marshal handles most types,
// but we document that the error path outputs a specific error message.
func TestLogger_JSONMarshalError(t *testing.T) {
	// Create a type that cannot be marshaled by JSON
	type unmarshalableType struct {
		Chan chan int
	}

	var buf bytes.Buffer
	logger := NewLogger(LevelInfo)
	logger.SetOutput(&buf)

	// This will fail to marshal because channels cannot be JSON encoded
	logger.Info("test", map[string]any{"data": unmarshalableType{make(chan int)}})

	output := buf.String()
	// Should contain the fallback error message
	assert.Contains(t, output, `{"level":"error","message":"failed to marshal log entry"}`)
}
