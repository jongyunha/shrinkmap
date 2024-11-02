package shrinkmap

import (
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestBasicOperations tests the basic operations of ShrinkableMap
func TestBasicOperations(t *testing.T) {
	t.Run("Set and Get operations", func(t *testing.T) {
		sm := New[string, int](DefaultConfig())

		// Test Set
		sm.Set("test1", 100)
		if val, exists := sm.Get("test1"); !exists || val != 100 {
			t.Errorf("Expected 100, got %v, exists: %v", val, exists)
		}

		// Test overwrite
		sm.Set("test1", 200)
		if val, exists := sm.Get("test1"); !exists || val != 200 {
			t.Errorf("Expected 200, got %v, exists: %v", val, exists)
		}

		// Test non-existent key
		if _, exists := sm.Get("nonexistent"); exists {
			t.Error("Expected false for non-existent key")
		}
	})

	t.Run("Delete operation", func(t *testing.T) {
		sm := New[string, int](DefaultConfig())

		sm.Set("test1", 100)
		if deleted := sm.Delete("test1"); !deleted {
			t.Error("Delete should return true for existing key")
		}

		if _, exists := sm.Get("test1"); exists {
			t.Error("Item should be deleted")
		}

		if deleted := sm.Delete("nonexistent"); deleted {
			t.Error("Delete should return false for non-existent key")
		}
	})

	t.Run("Len operation", func(t *testing.T) {
		sm := New[string, int](DefaultConfig())

		if l := sm.Len(); l != 0 {
			t.Errorf("Expected length 0, got %d", l)
		}

		sm.Set("test1", 100)
		sm.Set("test2", 200)

		// Wait for potential async operations
		time.Sleep(10 * time.Millisecond)

		if l := sm.Len(); l != 2 {
			t.Errorf("Expected length 2, got %d", l)
		}

		sm.Delete("test1")

		// Wait for potential async operations
		time.Sleep(10 * time.Millisecond)

		if l := sm.Len(); l != 1 {
			t.Errorf("Expected length 1, got %d", l)
		}
	})
}

// TestShrinking tests the shrinking functionality
func TestShrinking(t *testing.T) {
	t.Run("Force shrink", func(t *testing.T) {
		config := DefaultConfig()
		config.AutoShrinkEnabled = false
		sm := New[int, string](config)

		// Add items
		for i := 0; i < 100; i++ {
			sm.Set(i, fmt.Sprintf("value%d", i))
		}

		// Wait for operations to complete
		time.Sleep(10 * time.Millisecond)

		initialLen := sm.Len()
		if initialLen != 100 {
			t.Errorf("Expected initial length 100, got %d", initialLen)
		}

		// Delete items
		for i := 0; i < 40; i++ {
			sm.Delete(i)
		}

		// Wait for operations to complete
		time.Sleep(10 * time.Millisecond)

		afterDeleteLen := sm.Len()
		if afterDeleteLen != 60 {
			t.Errorf("Expected length after delete 60, got %d", afterDeleteLen)
		}

		// Force shrink
		if !sm.ForceShrink() {
			t.Error("Force shrink should return true")
		}

		// Wait for shrink to complete
		time.Sleep(50 * time.Millisecond)

		metrics := sm.GetMetrics()
		if metrics.totalShrinks != 1 {
			t.Errorf("Expected 1 shrink operation, got %d", metrics.totalShrinks)
		}

		afterShrinkLen := sm.Len()
		if afterShrinkLen != 60 {
			t.Errorf("Expected length after shrink 60, got %d", afterShrinkLen)
		}
	})
}

// TestConcurrency tests concurrent access to the map
func TestConcurrency(t *testing.T) {
	sm := New[int, int](DefaultConfig())
	const goroutines = 4
	const operationsPerGoroutine = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines * 2) // For both readers and writers

	// Writers
	for i := 0; i < goroutines; i++ {
		go func(base int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := base*operationsPerGoroutine + j
				sm.Set(key, j)
				if j%2 == 0 {
					sm.Delete(key)
				}
				// Small sleep to reduce contention
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	// Readers
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				sm.Get(rand.Intn(goroutines * operationsPerGoroutine))
				// Small sleep to reduce contention
				time.Sleep(time.Microsecond)
			}
		}()
	}

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(30 * time.Second):
		t.Fatal("Test timed out")
	}
}

// BenchmarkShrinkableMap provides performance benchmarks
func BenchmarkShrinkableMap(b *testing.B) {
	b.Run("Set", func(b *testing.B) {
		sm := New[int, int](DefaultConfig())
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				sm.Set(counter, counter)
				counter++
			}
		})
	})

	b.Run("Get", func(b *testing.B) {
		sm := New[int, int](DefaultConfig())
		itemCount := 1000
		for i := 0; i < itemCount; i++ {
			sm.Set(i, i)
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				sm.Get(counter % itemCount)
				counter++
			}
		})
	})

	b.Run("Delete", func(b *testing.B) {
		sm := New[int, int](DefaultConfig())
		itemCount := 1000
		for i := 0; i < itemCount; i++ {
			sm.Set(i, i)
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				sm.Delete(counter % itemCount)
				counter++
			}
		})
	})

	b.Run("Mixed Operations", func(b *testing.B) {
		sm := New[int, int](DefaultConfig())
		itemCount := 1000
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				key := counter % itemCount
				switch counter % 3 {
				case 0:
					sm.Set(key, counter)
				case 1:
					sm.Get(key)
				case 2:
					sm.Delete(key)
				}
				counter++
			}
		})
	})
}

// BenchmarkMapComparison compares ShrinkableMap with built-in map
func BenchmarkMapComparison(b *testing.B) {
	itemCount := 1000

	b.Run("Built-in Map", func(b *testing.B) {
		m := make(map[int]int)
		var mu sync.RWMutex
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				key := counter % itemCount
				switch counter % 3 {
				case 0:
					mu.Lock()
					m[key] = counter
					mu.Unlock()
				case 1:
					mu.RLock()
					_ = m[key]
					mu.RUnlock()
				case 2:
					mu.Lock()
					delete(m, key)
					mu.Unlock()
				}
				counter++
			}
		})
	})

	b.Run("ShrinkableMap", func(b *testing.B) {
		sm := New[int, int](DefaultConfig())
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				key := counter % itemCount
				switch counter % 3 {
				case 0:
					sm.Set(key, counter)
				case 1:
					sm.Get(key)
				case 2:
					sm.Delete(key)
				}
				counter++
			}
		})
	})
}

// TestErrorMetrics tests the error tracking functionality
func TestErrorMetrics(t *testing.T) {
	t.Run("Error Recording", func(t *testing.T) {
		sm := New[string, int](DefaultConfig())

		err := fmt.Errorf("test error")
		stack := getStackTrace()
		sm.metrics.RecordError(err, stack)

		metrics := sm.GetMetrics()
		if metrics.TotalErrors() != 1 {
			t.Errorf("Expected 1 error, got %d", metrics.TotalErrors())
		}

		lastErr := metrics.LastError()
		if lastErr == nil {
			t.Fatal("Expected last error to be recorded")
		}
		if lastErr.Error.(error).Error() != "test error" {
			t.Errorf("Expected 'test error', got %v", lastErr.Error)
		}
		if !strings.Contains(lastErr.Stack, "TestErrorMetrics") {
			t.Error("Stack trace should contain test function name")
		}
	})

	t.Run("Panic Recording", func(t *testing.T) {
		sm := New[string, int](DefaultConfig())

		// panic 발생 시뮬레이션
		panicVal := "test panic"
		stack := getStackTrace()
		sm.metrics.RecordPanic(panicVal, stack)

		metrics := sm.GetMetrics()
		if metrics.TotalPanics() != 1 {
			t.Errorf("Expected 1 panic, got %d", metrics.TotalPanics())
		}

		if metrics.LastPanicTime().IsZero() {
			t.Error("Last panic time should be set")
		}
	})

	t.Run("Error History", func(t *testing.T) {
		sm := New[string, int](DefaultConfig())

		for i := 0; i < 15; i++ {
			err := fmt.Errorf("error %d", i)
			sm.metrics.RecordError(err, getStackTrace())
		}

		history := sm.metrics.ErrorHistory()
		if len(history) > 10 {
			t.Errorf("Error history should be limited to 10 entries, got %d", len(history))
		}

		lastErr := history[len(history)-1]
		if lastErr.Error.(error).Error() != "error 14" {
			t.Errorf("Expected last error to be 'error 14', got %v", lastErr.Error)
		}
	})

	t.Run("Metrics Reset", func(t *testing.T) {
		sm := New[string, int](DefaultConfig())

		sm.metrics.RecordError(fmt.Errorf("test error"), getStackTrace())
		sm.metrics.RecordPanic("test panic", getStackTrace())

		sm.metrics.Reset()

		if sm.metrics.TotalErrors() != 0 {
			t.Error("Total errors should be reset to 0")
		}
		if sm.metrics.TotalPanics() != 0 {
			t.Error("Total panics should be reset to 0")
		}
		if sm.metrics.LastError() != nil {
			t.Error("Last error should be nil after reset")
		}
		if !sm.metrics.LastPanicTime().IsZero() {
			t.Error("Last panic time should be reset")
		}
	})
}

// TestConcurrentMetricsAccess tests concurrent access to metrics
func TestConcurrentMetricsAccess(t *testing.T) {
	sm := New[int, int](DefaultConfig())
	const goroutines = 10
	const operations = 100

	for i := 0; i < goroutines; i++ {
		for j := 0; j < operations; j++ {
			switch j % 3 {
			case 0:
				sm.metrics.RecordError(fmt.Errorf("error %d-%d", i, j), getStackTrace())
			case 1:
				sm.metrics.RecordPanic(fmt.Sprintf("panic %d-%d", i, j), getStackTrace())
			case 2:
				_ = sm.metrics.ErrorHistory()
			}
		}
	}

	metrics := sm.GetMetrics()
	history := metrics.ErrorHistory()
	if len(history) != 10 {
		t.Errorf("Expected error history to be limited to 10 entries, got %d", len(history))
	}
}

// TestStopFunctionality tests the Stop method
func TestStopFunctionality(t *testing.T) {
	t.Run("Stop Auto-shrink", func(t *testing.T) {
		config := DefaultConfig()
		config.AutoShrinkEnabled = true
		config.ShrinkInterval = 10 * time.Millisecond
		sm := New[string, int](config)

		for i := 0; i < 100; i++ {
			sm.Set(fmt.Sprintf("key%d", i), i)
		}

		sm.Stop()

		time.Sleep(50 * time.Millisecond)

		sm.Set("new-key", 1000)
		sm.Delete("new-key")

		if !sm.stopped.Load() {
			t.Error("Map should be marked as stopped")
		}
	})
}

// TestSnapshot tests the snapshot functionality
func TestSnapshot(t *testing.T) {
	t.Run("Basic Snapshot", func(t *testing.T) {
		sm := New[string, int](DefaultConfig())

		expectedData := map[string]int{
			"key1": 1,
			"key2": 2,
			"key3": 3,
		}

		for k, v := range expectedData {
			sm.Set(k, v)
		}

		snapshot := sm.Snapshot()

		if len(snapshot) != len(expectedData) {
			t.Errorf("Expected %d items in snapshot, got %d", len(expectedData), len(snapshot))
		}

		snapshotMap := make(map[string]int)
		for _, kv := range snapshot {
			snapshotMap[kv.Key] = kv.Value
		}

		for k, v := range expectedData {
			if sv, exists := snapshotMap[k]; !exists || sv != v {
				t.Errorf("Snapshot mismatch for key %s: expected %d, got %d", k, v, sv)
			}
		}
	})

	t.Run("Concurrent Snapshot", func(t *testing.T) {
		sm := New[int, int](DefaultConfig())
		const goroutines = 4
		const operations = 1000

		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < operations; j++ {
					key := id*operations + j
					switch j % 3 {
					case 0:
						sm.Set(key, j)
					case 2:
						sm.Delete(key)
					}
				}
			}(i)
		}

		wg.Wait()
	})
}

func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}
