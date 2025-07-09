package shrinkmap

import (
	"errors"
	"testing"
)

func TestErrorTypes(t *testing.T) {
	t.Run("ErrCode String representation", func(t *testing.T) {
		tests := []struct {
			code     ErrCode
			expected string
		}{
			{ErrCodeInvalidConfig, "INVALID_CONFIG"},
			{ErrCodeMapStopped, "MAP_STOPPED"},
			{ErrCodeShrinkFailed, "SHRINK_FAILED"},
			{ErrCodeBatchFailed, "BATCH_FAILED"},
			{ErrCodeCapacityExceeded, "CAPACITY_EXCEEDED"},
			{ErrCode(999), "UNKNOWN"},
		}

		for _, tt := range tests {
			if got := tt.code.String(); got != tt.expected {
				t.Errorf("ErrCode(%d).String() = %s, want %s", tt.code, got, tt.expected)
			}
		}
	})

	t.Run("ShrinkMapError creation and methods", func(t *testing.T) {
		err := NewShrinkMapError(ErrCodeMapStopped, "test_operation", "test message")
		
		if err.Code != ErrCodeMapStopped {
			t.Errorf("Expected code %v, got %v", ErrCodeMapStopped, err.Code)
		}
		if err.Operation != "test_operation" {
			t.Errorf("Expected operation 'test_operation', got %v", err.Operation)
		}
		if err.Message != "test message" {
			t.Errorf("Expected message 'test message', got %v", err.Message)
		}
		if err.Timestamp.IsZero() {
			t.Error("Expected timestamp to be set")
		}
		if err.Details == nil {
			t.Error("Expected details map to be initialized")
		}

		expectedError := "[MAP_STOPPED] test_operation: test message"
		if err.Error() != expectedError {
			t.Errorf("Expected error string %q, got %q", expectedError, err.Error())
		}
	})

	t.Run("ShrinkMapError with details", func(t *testing.T) {
		err := NewShrinkMapError(ErrCodeCapacityExceeded, "set", "capacity exceeded").
			WithDetails("currentSize", int64(100)).
			WithDetails("maxSize", int64(50))
		
		if len(err.Details) != 2 {
			t.Errorf("Expected 2 details, got %d", len(err.Details))
		}
		if err.Details["currentSize"] != int64(100) {
			t.Errorf("Expected currentSize to be 100, got %v", err.Details["currentSize"])
		}
		if err.Details["maxSize"] != int64(50) {
			t.Errorf("Expected maxSize to be 50, got %v", err.Details["maxSize"])
		}
	})

	t.Run("Error comparison with errors.Is", func(t *testing.T) {
		err1 := NewShrinkMapError(ErrCodeMapStopped, "op1", "message1")
		err2 := NewShrinkMapError(ErrCodeMapStopped, "op2", "message2")
		err3 := NewShrinkMapError(ErrCodeCapacityExceeded, "op3", "message3")
		
		if !errors.Is(err1, err2) {
			t.Error("Errors with same code should be equal")
		}
		if errors.Is(err1, err3) {
			t.Error("Errors with different codes should not be equal")
		}
	})

	t.Run("Predefined error instances", func(t *testing.T) {
		if ErrMapStopped.Code != ErrCodeMapStopped {
			t.Errorf("Expected ErrMapStopped code to be %v, got %v", ErrCodeMapStopped, ErrMapStopped.Code)
		}
		if ErrCapacityExceeded.Code != ErrCodeCapacityExceeded {
			t.Errorf("Expected ErrCapacityExceeded code to be %v, got %v", ErrCodeCapacityExceeded, ErrCapacityExceeded.Code)
		}
		if ErrInvalidConfig.Code != ErrCodeInvalidConfig {
			t.Errorf("Expected ErrInvalidConfig code to be %v, got %v", ErrCodeInvalidConfig, ErrInvalidConfig.Code)
		}
		if ErrShrinkFailed.Code != ErrCodeShrinkFailed {
			t.Errorf("Expected ErrShrinkFailed code to be %v, got %v", ErrCodeShrinkFailed, ErrShrinkFailed.Code)
		}
		if ErrBatchFailed.Code != ErrCodeBatchFailed {
			t.Errorf("Expected ErrBatchFailed code to be %v, got %v", ErrCodeBatchFailed, ErrBatchFailed.Code)
		}
	})

	t.Run("Error helper functions", func(t *testing.T) {
		mapStoppedErr := NewShrinkMapError(ErrCodeMapStopped, "test", "test")
		capacityErr := NewShrinkMapError(ErrCodeCapacityExceeded, "test", "test")
		otherErr := NewShrinkMapError(ErrCodeInvalidConfig, "test", "test")
		regularErr := errors.New("regular error")

		if !IsMapStopped(mapStoppedErr) {
			t.Error("IsMapStopped should return true for map stopped error")
		}
		if IsMapStopped(capacityErr) {
			t.Error("IsMapStopped should return false for capacity exceeded error")
		}
		if IsMapStopped(regularErr) {
			t.Error("IsMapStopped should return false for regular error")
		}

		if !IsCapacityExceeded(capacityErr) {
			t.Error("IsCapacityExceeded should return true for capacity exceeded error")
		}
		if IsCapacityExceeded(otherErr) {
			t.Error("IsCapacityExceeded should return false for other error")
		}
		if IsCapacityExceeded(regularErr) {
			t.Error("IsCapacityExceeded should return false for regular error")
		}
	})
}

func TestErrorScenarios(t *testing.T) {
	t.Run("Stopped map operations", func(t *testing.T) {
		sm := New[string, int](DefaultConfig())
		sm.Stop()
		
		// Test Set
		if err := sm.Set("key", 1); !IsMapStopped(err) {
			t.Errorf("Expected map stopped error, got %v", err)
		}
		
		// Test SetIf
		if _, err := sm.SetIf("key", 1, func(oldValue int, exists bool) bool { return true }); !IsMapStopped(err) {
			t.Errorf("Expected map stopped error, got %v", err)
		}
		
		// Test GetOrSet
		if _, _, err := sm.GetOrSet("key", 1); !IsMapStopped(err) {
			t.Errorf("Expected map stopped error, got %v", err)
		}
		
		// Test SetIfAbsent
		if _, err := sm.SetIfAbsent("key", 1); !IsMapStopped(err) {
			t.Errorf("Expected map stopped error, got %v", err)
		}
		
		// Test ApplyBatch
		batch := BatchOperations[string, int]{
			Operations: []BatchOperation[string, int]{
				{Type: BatchSet, Key: "key", Value: 1},
			},
		}
		if err := sm.ApplyBatch(batch); !IsMapStopped(err) {
			t.Errorf("Expected map stopped error, got %v", err)
		}
	})

	t.Run("Capacity exceeded scenarios", func(t *testing.T) {
		config := DefaultConfig()
		config.MaxMapSize = 2
		sm := New[string, int](config)
		defer sm.Stop()
		
		// Fill to capacity
		sm.Set("key1", 1)
		sm.Set("key2", 2)
		
		// Test Set exceeding capacity
		if err := sm.Set("key3", 3); !IsCapacityExceeded(err) {
			t.Errorf("Expected capacity exceeded error, got %v", err)
		}
		
		// Test SetIf exceeding capacity
		if _, err := sm.SetIf("key4", 4, func(oldValue int, exists bool) bool { return true }); !IsCapacityExceeded(err) {
			t.Errorf("Expected capacity exceeded error, got %v", err)
		}
		
		// Test GetOrSet exceeding capacity
		if _, _, err := sm.GetOrSet("key5", 5); !IsCapacityExceeded(err) {
			t.Errorf("Expected capacity exceeded error, got %v", err)
		}
		
		// Test batch operation exceeding capacity
		batch := BatchOperations[string, int]{
			Operations: []BatchOperation[string, int]{
				{Type: BatchSet, Key: "key6", Value: 6},
				{Type: BatchSet, Key: "key7", Value: 7},
			},
		}
		if err := sm.ApplyBatch(batch); !IsCapacityExceeded(err) {
			t.Errorf("Expected capacity exceeded error, got %v", err)
		}
	})

	t.Run("Empty batch operations", func(t *testing.T) {
		sm := New[string, int](DefaultConfig())
		defer sm.Stop()
		
		batch := BatchOperations[string, int]{
			Operations: []BatchOperation[string, int]{},
		}
		if err := sm.ApplyBatch(batch); err != nil {
			t.Errorf("Expected no error for empty batch, got %v", err)
		}
	})
}