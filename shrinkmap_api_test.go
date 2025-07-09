package shrinkmap

import (
	"errors"
	"testing"
)

func TestNewAPIMethods(t *testing.T) {
	t.Run("Contains method", func(t *testing.T) {
		sm := New[string, int](DefaultConfig())
		defer sm.Stop()

		if sm.Contains("nonexistent") {
			t.Error("Contains should return false for non-existent key")
		}

		if err := sm.Set("test", 100); err != nil {
			t.Errorf("Set should not error: %v", err)
		}
		if !sm.Contains("test") {
			t.Error("Contains should return true for existing key")
		}

		sm.Delete("test")
		if sm.Contains("test") {
			t.Error("Contains should return false after deletion")
		}
	})

	t.Run("SetIf method", func(t *testing.T) {
		sm := New[string, int](DefaultConfig())
		defer sm.Stop()

		// Test setting when key doesn't exist
		if success, err := sm.SetIf("key1", 100, func(oldValue int, exists bool) bool {
			return !exists
		}); err != nil || !success {
			t.Errorf("SetIf should succeed when condition is met: success=%v, err=%v", success, err)
		}

		// Test not setting when key exists
		if success, err := sm.SetIf("key1", 200, func(oldValue int, exists bool) bool {
			return !exists
		}); err != nil || success {
			t.Errorf("SetIf should not succeed when condition is not met: success=%v, err=%v", success, err)
		}

		// Verify original value is preserved
		if val, exists := sm.Get("key1"); !exists || val != 100 {
			t.Errorf("Expected 100, got %v", val)
		}
	})

	t.Run("GetOrSet method", func(t *testing.T) {
		sm := New[string, int](DefaultConfig())
		defer sm.Stop()

		// Test getting when key doesn't exist (should set)
		val, existed, err := sm.GetOrSet("key1", 100)
		if err != nil || val != 100 || existed {
			t.Errorf("Expected (100, false, nil), got (%v, %v, %v)", val, existed, err)
		}

		// Test getting when key exists (should not set)
		val, existed, err = sm.GetOrSet("key1", 200)
		if err != nil || val != 100 || !existed {
			t.Errorf("Expected (100, true, nil), got (%v, %v, %v)", val, existed, err)
		}
	})

	t.Run("SetIfAbsent method", func(t *testing.T) {
		sm := New[string, int](DefaultConfig())
		defer sm.Stop()

		// Test setting when key doesn't exist
		if success, err := sm.SetIfAbsent("key1", 100); err != nil || !success {
			t.Errorf("SetIfAbsent should succeed when key doesn't exist: success=%v, err=%v", success, err)
		}

		// Test not setting when key exists
		if success, err := sm.SetIfAbsent("key1", 200); err != nil || success {
			t.Errorf("SetIfAbsent should not succeed when key exists: success=%v, err=%v", success, err)
		}

		// Verify original value is preserved
		if val, exists := sm.Get("key1"); !exists || val != 100 {
			t.Errorf("Expected 100, got %v", val)
		}
	})

	t.Run("Error handling", func(t *testing.T) {
		// Test capacity exceeded error
		config := DefaultConfig()
		config.MaxMapSize = 2
		sm := New[string, int](config)
		defer sm.Stop()

		// Fill to capacity
		if err := sm.Set("key1", 1); err != nil {
			t.Errorf("First set should not error: %v", err)
		}
		if err := sm.Set("key2", 2); err != nil {
			t.Errorf("Second set should not error: %v", err)
		}

		// Try to exceed capacity
		err := sm.Set("key3", 3)
		if err == nil {
			t.Error("Set should return error when capacity exceeded")
		}
		if !IsCapacityExceeded(err) {
			t.Errorf("Expected capacity exceeded error, got %v", err)
		}

		// Test stopped map error
		sm.Stop()
		err = sm.Set("key4", 4)
		if err == nil {
			t.Error("Set should return error when map is stopped")
		}
		if !IsMapStopped(err) {
			t.Errorf("Expected map stopped error, got %v", err)
		}
	})
}

func TestAPIErrorTypes(t *testing.T) {
	t.Run("ShrinkMapError", func(t *testing.T) {
		err := NewShrinkMapError(ErrCodeInvalidConfig, "test", "test message")
		if err.Code != ErrCodeInvalidConfig {
			t.Errorf("Expected code %v, got %v", ErrCodeInvalidConfig, err.Code)
		}
		if err.Message != "test message" {
			t.Errorf("Expected message 'test message', got %v", err.Message)
		}
	})

	t.Run("Error comparison", func(t *testing.T) {
		err1 := NewShrinkMapError(ErrCodeMapStopped, "test", "test message")
		err2 := NewShrinkMapError(ErrCodeMapStopped, "test2", "test message 2")

		if !errors.Is(err1, err2) {
			t.Error("Errors with same code should be equal")
		}

		if !IsMapStopped(err1) {
			t.Error("Should recognize map stopped error")
		}
	})
}

func BenchmarkNewAPIMethods(b *testing.B) {
	sm := New[int, int](DefaultConfig())
	defer sm.Stop()

	// Pre-populate some data
	for i := 0; i < 1000; i++ {
		_ = sm.Set(i, i)
	}

	b.Run("Contains", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				sm.Contains(counter % 1000)
				counter++
			}
		})
	})

	b.Run("GetOrSet", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				_, _, err := sm.GetOrSet(counter, counter)
				if err != nil {
					return
				}
				counter++
			}
		})
	})

	b.Run("SetIfAbsent", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				_, _ = sm.SetIfAbsent(counter+10000, counter)
				counter++
			}
		})
	})
}
