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

		if l := sm.Len(); l != 2 {
			t.Errorf("Expected length 2, got %d", l)
		}

		sm.Delete("test1")
		if l := sm.Len(); l != 1 {
			t.Errorf("Expected length 1, got %d", l)
		}
	})
}

// TestShrinking tests the shrinking functionality
func TestShrinking(t *testing.T) {
	t.Run("Auto shrinking", func(t *testing.T) {
		config := DefaultConfig()
		config.ShrinkInterval = 100 * time.Millisecond
		config.ShrinkRatio = 0.3
		config.MinShrinkInterval = 50 * time.Millisecond

		sm := New[int, string](config)

		// Add items
		for i := 0; i < 100; i++ {
			sm.Set(i, fmt.Sprintf("value%d", i))
		}

		// Delete some items
		for i := 0; i < 40; i++ {
			sm.Delete(i)
		}

		// Wait for auto-shrink
		time.Sleep(200 * time.Millisecond)

		metrics := sm.GetMetrics()
		if metrics.TotalShrinks == 0 {
			t.Error("Expected at least one shrink operation")
		}
	})

	t.Run("Force shrink", func(t *testing.T) {
		sm := New[int, string](DefaultConfig())

		// Add items
		for i := 0; i < 100; i++ {
			sm.Set(i, fmt.Sprintf("value%d", i))
		}

		// Delete items
		for i := 0; i < 40; i++ {
			sm.Delete(i)
		}

		if !sm.ForceShrink() {
			t.Error("Force shrink should return true")
		}

		metrics := sm.GetMetrics()
		if metrics.TotalShrinks != 1 {
			t.Errorf("Expected 1 shrink operation, got %d", metrics.TotalShrinks)
		}
	})
}

// TestConcurrency tests concurrent access to the map
func TestConcurrency(t *testing.T) {
	sm := New[int, int](DefaultConfig())
	const goroutines = 10
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
			}
		}(i)
	}

	// Readers
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				sm.Get(rand.Intn(goroutines * operationsPerGoroutine))
			}
		}()
	}

	wg.Wait()
}

// BenchmarkShrinkableMap provides performance benchmarks
func BenchmarkShrinkableMap(b *testing.B) {
	b.Run("Set", func(b *testing.B) {
		sm := New[int, int](DefaultConfig())
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			sm.Set(i, i)
		}
	})

	b.Run("Get", func(b *testing.B) {
		sm := New[int, int](DefaultConfig())
		// Reduce initial dataset size
		itemCount := 100
		for i := 0; i < itemCount; i++ {
			sm.Set(i, i)
		}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			sm.Get(i % itemCount)
		}
	})

	b.Run("Delete", func(b *testing.B) {
		sm := New[int, int](DefaultConfig())
		// Pre-populate with reasonable amount
		itemCount := b.N
		if itemCount > 10000 {
			itemCount = 10000
		}
		for i := 0; i < itemCount; i++ {
			sm.Set(i, i)
		}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			sm.Delete(i % itemCount)
		}
	})

	b.Run("Mixed Operations", func(b *testing.B) {
		sm := New[int, int](DefaultConfig())
		itemCount := 1000 // Fixed size for mixed operations
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			key := i % itemCount
			switch i % 3 {
			case 0:
				sm.Set(key, i)
			case 1:
				sm.Get(key)
			case 2:
				sm.Delete(key)
			}
		}
	})
}

// BenchmarkConcurrent tests concurrent performance
func BenchmarkConcurrent(b *testing.B) {
	// Reduce number of goroutines for testing
	for _, goroutines := range []int{1, 4, 8} {
		b.Run(fmt.Sprintf("Goroutines-%d", goroutines), func(b *testing.B) {
			sm := New[int, int](DefaultConfig())
			var wg sync.WaitGroup
			opsPerGoroutine := b.N / goroutines

			b.ResetTimer()

			for i := 0; i < goroutines; i++ {
				wg.Add(1)
				go func(base int) {
					defer wg.Done()
					for j := 0; j < opsPerGoroutine; j++ {
						key := (base*opsPerGoroutine + j) % 1000 // Limit key range
						sm.Set(key, j)
						sm.Get(key)
						if j%2 == 0 {
							sm.Delete(key)
						}
					}
				}(i)
			}

			wg.Wait()
		})
	}
}

// BenchmarkShrinking tests shrinking performance
func BenchmarkShrinking(b *testing.B) {
	b.Run("Shrink Performance", func(b *testing.B) {
		config := DefaultConfig()
		config.ShrinkRatio = 0.3
		sm := New[int, int](config)

		// Reduce dataset size
		itemCount := 10000

		// Prepare data
		for i := 0; i < itemCount; i++ {
			sm.Set(i, i)
		}

		// Delete 40% of items
		for i := 0; i < itemCount*4/10; i++ {
			sm.Delete(i)
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			sm.ForceShrink()
		}
	})
}

// BenchmarkMapComparison compares ShrinkableMap with built-in map
func BenchmarkMapComparison(b *testing.B) {
	itemCount := 1000 // Fixed size for comparison

	b.Run("Built-in Map", func(b *testing.B) {
		m := make(map[int]int)
		var mu sync.RWMutex

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			key := i % itemCount
			switch i % 3 {
			case 0:
				mu.Lock()
				m[key] = i
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
		}
	})

	b.Run("ShrinkableMap", func(b *testing.B) {
		sm := New[int, int](DefaultConfig())

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			key := i % itemCount
			switch i % 3 {
			case 0:
				sm.Set(key, i)
			case 1:
				sm.Get(key)
			case 2:
				sm.Delete(key)
			}
		}
	})
}
