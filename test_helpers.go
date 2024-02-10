package migration

import (
	"bytes"
	"log"
	"testing"
)

// LogHelper encapsulates logging functionality
type LogHelper struct {
	buffer *bytes.Buffer
}

// Logger interface defines the Fatal method
type Logger interface {
	Fatal(v ...interface{})
}

// NewLogHelper initializes a LogHelper instance
func NewLogHelper() *LogHelper {
	return &LogHelper{
		buffer: &bytes.Buffer{},
	}
}

// Fatal is a method to log a fatal message
func (lh *LogHelper) Fatal(v ...interface{}) {
	log.Print(v...)
}

// TestLogFatalError is a helper function to test a function that might call log.Fatal
func TestLogFatalError(t *testing.T, expectedErrorMsg string, f func()) bool {
	// Initialize log helper
	logHelper := NewLogHelper()

	// Replace log output to capture the message
	origOutput := log.Writer()
	defer log.SetOutput(origOutput)
	log.SetOutput(logHelper.buffer)

	// Call the function that may lead to log.Fatal
	f()

	println("here")

	// Check if the logged message matches the expected error message
	return bytes.Contains(logHelper.buffer.Bytes(), []byte(expectedErrorMsg))
}
