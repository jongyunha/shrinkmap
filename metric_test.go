package shrinkmap

import (
	"testing"
)

func TestMetrics(t *testing.T) {
	t.Run("Metrics accessor methods", func(t *testing.T) {
		sm := New[string, int](DefaultConfig())
		defer sm.Stop()
		
		// Initial state
		metrics := sm.GetMetrics()
		if metrics.TotalShrinks() != 0 {
			t.Errorf("Expected 0 total shrinks, got %d", metrics.TotalShrinks())
		}
		if metrics.TotalItemsProcessed() != 0 {
			t.Errorf("Expected 0 total items processed, got %d", metrics.TotalItemsProcessed())
		}
		if metrics.PeakSize() != 0 {
			t.Errorf("Expected 0 peak size, got %d", metrics.PeakSize())
		}
		if metrics.TotalErrors() != 0 {
			t.Errorf("Expected 0 total errors, got %d", metrics.TotalErrors())
		}
		if metrics.TotalPanics() != 0 {
			t.Errorf("Expected 0 total panics, got %d", metrics.TotalPanics())
		}
		if metrics.LastError() != nil {
			t.Errorf("Expected nil last error, got %v", metrics.LastError())
		}
		if !metrics.LastPanicTime().IsZero() {
			t.Errorf("Expected zero last panic time, got %v", metrics.LastPanicTime())
		}
		if metrics.LastShrinkDuration() != 0 {
			t.Errorf("Expected 0 last shrink duration, got %v", metrics.LastShrinkDuration())
		}
		
		// Add some items to trigger metrics
		sm.Set("key1", 1)
		sm.Set("key2", 2)
		sm.Set("key3", 3)
		
		metrics = sm.GetMetrics()
		if metrics.TotalItemsProcessed() != 3 {
			t.Errorf("Expected 3 total items processed, got %d", metrics.TotalItemsProcessed())
		}
		if metrics.PeakSize() != 3 {
			t.Errorf("Expected 3 peak size, got %d", metrics.PeakSize())
		}
	})

	t.Run("Metrics error recording", func(t *testing.T) {
		sm := New[string, int](DefaultConfig())
		defer sm.Stop()
		
		// Record some errors
		sm.metrics.RecordError(ErrInvalidConfig, "test stack trace")
		sm.metrics.RecordError(ErrMapStopped, "test stack trace 2")
		
		metrics := sm.GetMetrics()
		if metrics.TotalErrors() != 2 {
			t.Errorf("Expected 2 total errors, got %d", metrics.TotalErrors())
		}
		
		lastError := metrics.LastError()
		if lastError == nil {
			t.Fatal("Expected last error to be recorded")
		}
		if lastError.Error != ErrMapStopped {
			t.Errorf("Expected last error to be ErrMapStopped, got %v", lastError.Error)
		}
		if lastError.Stack != "test stack trace 2" {
			t.Errorf("Expected stack trace 'test stack trace 2', got %s", lastError.Stack)
		}
		
		history := metrics.ErrorHistory()
		if len(history) != 2 {
			t.Errorf("Expected 2 errors in history, got %d", len(history))
		}
	})

	t.Run("Metrics panic recording", func(t *testing.T) {
		sm := New[string, int](DefaultConfig())
		defer sm.Stop()
		
		// Record some panics
		sm.metrics.RecordPanic("test panic", "test stack trace")
		
		metrics := sm.GetMetrics()
		if metrics.TotalPanics() != 1 {
			t.Errorf("Expected 1 total panic, got %d", metrics.TotalPanics())
		}
		
		if metrics.LastPanicTime().IsZero() {
			t.Error("Expected last panic time to be set")
		}
		
		// Panic should also be recorded in error history
		lastError := metrics.LastError()
		if lastError == nil {
			t.Fatal("Expected last error to be recorded for panic")
		}
		if lastError.Error != "test panic" {
			t.Errorf("Expected panic value 'test panic', got %v", lastError.Error)
		}
	})

	t.Run("Error history limit", func(t *testing.T) {
		sm := New[string, int](DefaultConfig())
		defer sm.Stop()
		
		// Record more than 10 errors
		for i := 0; i < 15; i++ {
			sm.metrics.RecordError(ErrInvalidConfig, "test stack trace")
		}
		
		metrics := sm.GetMetrics()
		if metrics.TotalErrors() != 15 {
			t.Errorf("Expected 15 total errors, got %d", metrics.TotalErrors())
		}
		
		history := metrics.ErrorHistory()
		if len(history) != 10 {
			t.Errorf("Expected error history to be limited to 10 entries, got %d", len(history))
		}
	})

	t.Run("Metrics reset", func(t *testing.T) {
		sm := New[string, int](DefaultConfig())
		defer sm.Stop()
		
		// Add some data and errors
		sm.Set("key", 1)
		sm.metrics.RecordError(ErrInvalidConfig, "test stack trace")
		sm.metrics.RecordPanic("test panic", "test stack trace")
		
		// Force a shrink to generate shrink metrics
		sm.ForceShrink()
		
		// Verify metrics are recorded
		metrics := sm.GetMetrics()
		if metrics.TotalErrors() == 0 {
			t.Error("Expected some errors before reset")
		}
		if metrics.TotalPanics() == 0 {
			t.Error("Expected some panics before reset")
		}
		
		// Reset metrics
		sm.metrics.Reset()
		
		// Verify everything is reset
		metrics = sm.GetMetrics()
		if metrics.TotalShrinks() != 0 {
			t.Errorf("Expected 0 total shrinks after reset, got %d", metrics.TotalShrinks())
		}
		if metrics.TotalItemsProcessed() != 0 {
			t.Errorf("Expected 0 total items processed after reset, got %d", metrics.TotalItemsProcessed())
		}
		if metrics.PeakSize() != 0 {
			t.Errorf("Expected 0 peak size after reset, got %d", metrics.PeakSize())
		}
		if metrics.TotalErrors() != 0 {
			t.Errorf("Expected 0 total errors after reset, got %d", metrics.TotalErrors())
		}
		if metrics.TotalPanics() != 0 {
			t.Errorf("Expected 0 total panics after reset, got %d", metrics.TotalPanics())
		}
		if metrics.LastError() != nil {
			t.Errorf("Expected nil last error after reset, got %v", metrics.LastError())
		}
		if !metrics.LastPanicTime().IsZero() {
			t.Errorf("Expected zero last panic time after reset, got %v", metrics.LastPanicTime())
		}
		if metrics.LastShrinkDuration() != 0 {
			t.Errorf("Expected 0 last shrink duration after reset, got %v", metrics.LastShrinkDuration())
		}
		
		history := metrics.ErrorHistory()
		if len(history) != 0 {
			t.Errorf("Expected empty error history after reset, got %d entries", len(history))
		}
	})

	t.Run("Thread-safe metrics access", func(t *testing.T) {
		sm := New[string, int](DefaultConfig())
		defer sm.Stop()
		
		done := make(chan bool, 1)
		go func() {
			for i := 0; i < 1000; i++ {
				sm.metrics.RecordError(ErrInvalidConfig, "test stack trace")
				sm.metrics.RecordPanic("test panic", "test stack trace")
				sm.GetMetrics() // Access metrics concurrently
			}
			done <- true
		}()
		
		go func() {
			for i := 0; i < 1000; i++ {
				sm.GetMetrics() // Access metrics concurrently
			}
			done <- true
		}()
		
		// Wait for both goroutines to complete
		<-done
		<-done
		
		// Verify final state
		metrics := sm.GetMetrics()
		if metrics.TotalErrors() != 1000 {
			t.Errorf("Expected 1000 total errors, got %d", metrics.TotalErrors())
		}
		if metrics.TotalPanics() != 1000 {
			t.Errorf("Expected 1000 total panics, got %d", metrics.TotalPanics())
		}
	})
}

func TestShrinkMetrics(t *testing.T) {
	t.Run("Shrink metrics recording", func(t *testing.T) {
		config := DefaultConfig()
		config.AutoShrinkEnabled = false // Disable auto-shrink for manual control
		sm := New[string, int](config)
		defer sm.Stop()
		
		// Add and delete items to trigger shrink
		for i := 0; i < 100; i++ {
			sm.Set(string(rune('a'+i)), i)
		}
		for i := 0; i < 50; i++ {
			sm.Delete(string(rune('a'+i)))
		}
		
		// Force shrink
		if !sm.ForceShrink() {
			t.Error("Expected force shrink to succeed")
		}
		
		metrics := sm.GetMetrics()
		if metrics.TotalShrinks() != 1 {
			t.Errorf("Expected 1 total shrink, got %d", metrics.TotalShrinks())
		}
		if metrics.LastShrinkDuration() <= 0 {
			t.Errorf("Expected positive last shrink duration, got %v", metrics.LastShrinkDuration())
		}
	})
}