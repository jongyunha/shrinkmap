package shrinkmap

import (
	"fmt"
	"math/rand"
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
