package shrinkmap

import (
	"errors"
	"fmt"
	"time"
)

// ErrCode represents different types of errors that can occur in ShrinkableMap
type ErrCode int

const (
	// ErrCodeInvalidConfig indicates configuration validation failed
	ErrCodeInvalidConfig ErrCode = iota
	// ErrCodeMapStopped indicates operation attempted on stopped map
	ErrCodeMapStopped
	// ErrCodeShrinkFailed indicates shrink operation failed
	ErrCodeShrinkFailed
	// ErrCodeBatchFailed indicates batch operation failed
	ErrCodeBatchFailed
	// ErrCodeCapacityExceeded indicates maximum capacity exceeded
	ErrCodeCapacityExceeded
)

// String returns the string representation of the error code
func (e ErrCode) String() string {
	switch e {
	case ErrCodeInvalidConfig:
		return "INVALID_CONFIG"
	case ErrCodeMapStopped:
		return "MAP_STOPPED"
	case ErrCodeShrinkFailed:
		return "SHRINK_FAILED"
	case ErrCodeBatchFailed:
		return "BATCH_FAILED"
	case ErrCodeCapacityExceeded:
		return "CAPACITY_EXCEEDED"
	default:
		return "UNKNOWN"
	}
}

// ShrinkMapError represents a structured error from ShrinkableMap operations
type ShrinkMapError struct {
	Code      ErrCode
	Message   string
	Operation string
	Timestamp time.Time
	Details   map[string]interface{}
}

// Error implements the error interface
func (e *ShrinkMapError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Operation, e.Message)
}

// Is allows error comparison using errors.Is
func (e *ShrinkMapError) Is(target error) bool {
	var shrinkMapErr *ShrinkMapError
	if errors.As(target, &shrinkMapErr) {
		return e.Code == shrinkMapErr.Code
	}
	return false
}

// NewShrinkMapError creates a new ShrinkMapError
func NewShrinkMapError(code ErrCode, operation, message string) *ShrinkMapError {
	return &ShrinkMapError{
		Code:      code,
		Message:   message,
		Operation: operation,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}
}

// WithDetails adds details to the error
func (e *ShrinkMapError) WithDetails(key string, value interface{}) *ShrinkMapError {
	e.Details[key] = value
	return e
}

// Common error instances
var (
	ErrMapStopped        = NewShrinkMapError(ErrCodeMapStopped, "operation", "map has been stopped")
	ErrCapacityExceeded  = NewShrinkMapError(ErrCodeCapacityExceeded, "set", "maximum capacity exceeded")
	ErrInvalidConfig     = NewShrinkMapError(ErrCodeInvalidConfig, "config", "invalid configuration")
	ErrShrinkFailed      = NewShrinkMapError(ErrCodeShrinkFailed, "shrink", "shrink operation failed")
	ErrBatchFailed       = NewShrinkMapError(ErrCodeBatchFailed, "batch", "batch operation failed")
)

// IsMapStopped checks if the error indicates a stopped map
func IsMapStopped(err error) bool {
	var shrinkMapErr *ShrinkMapError
	return errors.As(err, &shrinkMapErr) && shrinkMapErr.Code == ErrCodeMapStopped
}

// IsCapacityExceeded checks if the error indicates capacity exceeded
func IsCapacityExceeded(err error) bool {
	var shrinkMapErr *ShrinkMapError
	return errors.As(err, &shrinkMapErr) && shrinkMapErr.Code == ErrCodeCapacityExceeded
}