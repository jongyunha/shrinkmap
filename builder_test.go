package shrinkmap

import (
	"testing"
	"time"
)

func TestBuilder(t *testing.T) {
	t.Run("Basic builder usage", func(t *testing.T) {
		sm := NewBuilder[string, int]().
			WithShrinkRatio(0.5).
			WithInitialCapacity(100).
			WithShrinkInterval(10 * time.Second).
			WithMinShrinkInterval(5 * time.Second).
			WithMaxMapSize(1000).
			WithCapacityGrowthFactor(1.5).
			WithAutoShrink(false).
			Build()
		defer sm.Stop()

		if sm.config.ShrinkRatio != 0.5 {
			t.Errorf("Expected ShrinkRatio 0.5, got %v", sm.config.ShrinkRatio)
		}
		if sm.config.InitialCapacity != 100 {
			t.Errorf("Expected InitialCapacity 100, got %v", sm.config.InitialCapacity)
		}
		if sm.config.ShrinkInterval != 10*time.Second {
			t.Errorf("Expected ShrinkInterval 10s, got %v", sm.config.ShrinkInterval)
		}
		if sm.config.MinShrinkInterval != 5*time.Second {
			t.Errorf("Expected MinShrinkInterval 5s, got %v", sm.config.MinShrinkInterval)
		}
		if sm.config.MaxMapSize != 1000 {
			t.Errorf("Expected MaxMapSize 1000, got %v", sm.config.MaxMapSize)
		}
		if sm.config.CapacityGrowthFactor != 1.5 {
			t.Errorf("Expected CapacityGrowthFactor 1.5, got %v", sm.config.CapacityGrowthFactor)
		}
		if sm.config.AutoShrinkEnabled {
			t.Error("Expected AutoShrinkEnabled false")
		}
	})

	t.Run("Builder with validation", func(t *testing.T) {
		// Valid configuration
		sm, err := NewBuilder[string, int]().
			WithShrinkRatio(0.5).
			WithInitialCapacity(100).
			BuildWithValidation()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if sm == nil {
			t.Error("Expected non-nil ShrinkableMap")
		}
		if sm != nil {
			sm.Stop()
		}

		// Invalid configuration
		_, err = NewBuilder[string, int]().
			WithShrinkRatio(0).
			BuildWithValidation()
		if err == nil {
			t.Error("Expected validation error for invalid shrink ratio")
		}
	})

	t.Run("Default builder configuration", func(t *testing.T) {
		sm := NewBuilder[string, int]().Build()
		defer sm.Stop()

		defaultConfig := DefaultConfig()
		if sm.config.ShrinkRatio != defaultConfig.ShrinkRatio {
			t.Errorf("Expected default ShrinkRatio %v, got %v", defaultConfig.ShrinkRatio, sm.config.ShrinkRatio)
		}
		if sm.config.InitialCapacity != defaultConfig.InitialCapacity {
			t.Errorf("Expected default InitialCapacity %v, got %v", defaultConfig.InitialCapacity, sm.config.InitialCapacity)
		}
		if sm.config.AutoShrinkEnabled != defaultConfig.AutoShrinkEnabled {
			t.Errorf("Expected default AutoShrinkEnabled %v, got %v", defaultConfig.AutoShrinkEnabled, sm.config.AutoShrinkEnabled)
		}
	})
}

func TestMapBuilder(t *testing.T) {
	t.Run("Basic map builder usage", func(t *testing.T) {
		sm := NewBuilder[string, int]().
			WithMaxMapSize(10).
			Build()
		defer sm.Stop()

		mb := NewMapBuilder(sm)
		finalMap := mb.
			Set("key1", 1).
			Set("key2", 2).
			Set("key3", 3).
			SetIfAbsent("key4", 4).
			SetIfAbsent("key1", 100). // Should not overwrite
			Delete("key2").
			Done()

		if finalMap != sm {
			t.Error("Expected Done() to return the original map")
		}

		// Verify operations
		if val, exists := sm.Get("key1"); !exists || val != 1 {
			t.Errorf("Expected key1=1, got %v (exists: %v)", val, exists)
		}
		if val, exists := sm.Get("key2"); exists {
			t.Errorf("Expected key2 to be deleted, but got %v", val)
		}
		if val, exists := sm.Get("key3"); !exists || val != 3 {
			t.Errorf("Expected key3=3, got %v (exists: %v)", val, exists)
		}
		if val, exists := sm.Get("key4"); !exists || val != 4 {
			t.Errorf("Expected key4=4, got %v (exists: %v)", val, exists)
		}
	})

	t.Run("Map builder with Map() method", func(t *testing.T) {
		sm := NewBuilder[string, int]().Build()
		defer sm.Stop()

		mb := NewMapBuilder(sm)
		retrievedMap := mb.Set("key", 42).Map()

		if retrievedMap != sm {
			t.Error("Expected Map() to return the original map")
		}

		if val, exists := retrievedMap.Get("key"); !exists || val != 42 {
			t.Errorf("Expected key=42, got %v (exists: %v)", val, exists)
		}
	})
}

func TestBuilderChaining(t *testing.T) {
	t.Run("Complex chaining example", func(t *testing.T) {
		sm := NewBuilder[string, int]().
			WithShrinkRatio(0.3).
			WithInitialCapacity(50).
			WithAutoShrink(true).
			WithMaxMapSize(100).
			Build()
		defer sm.Stop()

		// Use map builder to populate data
		NewMapBuilder(sm).
			Set("a", 1).
			Set("b", 2).
			Set("c", 3).
			SetIfAbsent("d", 4).
			Delete("b")

		// Verify final state
		if sm.Len() != 3 {
			t.Errorf("Expected length 3, got %d", sm.Len())
		}

		expectedKeys := []string{"a", "c", "d"}
		for _, key := range expectedKeys {
			if !sm.Contains(key) {
				t.Errorf("Expected key %s to exist", key)
			}
		}

		if sm.Contains("b") {
			t.Error("Expected key b to be deleted")
		}
	})

	t.Run("Builder method chaining order independence", func(t *testing.T) {
		// Build in one order
		sm1 := NewBuilder[string, int]().
			WithShrinkRatio(0.4).
			WithInitialCapacity(64).
			WithAutoShrink(false).
			Build()
		defer sm1.Stop()

		// Build in different order
		sm2 := NewBuilder[string, int]().
			WithAutoShrink(false).
			WithInitialCapacity(64).
			WithShrinkRatio(0.4).
			Build()
		defer sm2.Stop()

		// Both should have same configuration
		if sm1.config.ShrinkRatio != sm2.config.ShrinkRatio {
			t.Errorf("Expected same ShrinkRatio, got %v and %v", sm1.config.ShrinkRatio, sm2.config.ShrinkRatio)
		}
		if sm1.config.InitialCapacity != sm2.config.InitialCapacity {
			t.Errorf("Expected same InitialCapacity, got %v and %v", sm1.config.InitialCapacity, sm2.config.InitialCapacity)
		}
		if sm1.config.AutoShrinkEnabled != sm2.config.AutoShrinkEnabled {
			t.Errorf("Expected same AutoShrinkEnabled, got %v and %v", sm1.config.AutoShrinkEnabled, sm2.config.AutoShrinkEnabled)
		}
	})
}

func BenchmarkBuilder(b *testing.B) {
	b.Run("Builder creation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sm := NewBuilder[int, int]().
				WithShrinkRatio(0.5).
				WithInitialCapacity(100).
				WithAutoShrink(false).
				Build()
			sm.Stop()
		}
	})

	b.Run("MapBuilder operations", func(b *testing.B) {
		sm := NewBuilder[int, int]().
			WithAutoShrink(false).
			Build()
		defer sm.Stop()

		mb := NewMapBuilder(sm)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			mb.Set(i, i).SetIfAbsent(i+1000, i+1000)
		}
	})
}
