package shrinkmap

import (
	"testing"
	"time"
)

func TestBatchOperations(t *testing.T) {
	config := Config{
		InitialCapacity:      10,
		AutoShrinkEnabled:    true,
		ShrinkInterval:       time.Second,
		MinShrinkInterval:    time.Second,
		ShrinkRatio:          0.5,
		CapacityGrowthFactor: 1.5,
	}

	t.Run("Basic Batch Set Operations", func(t *testing.T) {
		sm := New[string, int](config)
		defer sm.Stop()

		batch := BatchOperations[string, int]{
			Operations: []BatchOperation[string, int]{
				{Type: BatchSet, Key: "a", Value: 1},
				{Type: BatchSet, Key: "b", Value: 2},
				{Type: BatchSet, Key: "c", Value: 3},
			},
		}

		err := sm.ApplyBatch(batch)
		if err != nil {
			t.Errorf("ApplyBatch failed: %v", err)
		}

		if val, exists := sm.Get("a"); !exists || val != 1 {
			t.Errorf("Expected a=1, got %v, exists=%v", val, exists)
		}
		if val, exists := sm.Get("b"); !exists || val != 2 {
			t.Errorf("Expected b=2, got %v, exists=%v", val, exists)
		}
		if val, exists := sm.Get("c"); !exists || val != 3 {
			t.Errorf("Expected c=3, got %v, exists=%v", val, exists)
		}

		if sm.Len() != 3 {
			t.Errorf("Expected length 3, got %d", sm.Len())
		}
	})

	t.Run("Mixed Batch Operations", func(t *testing.T) {
		sm := New[string, int](config)
		defer sm.Stop()

		sm.Set("existing1", 100)
		sm.Set("existing2", 200)
		sm.Set("toDelete", 300)

		batch := BatchOperations[string, int]{
			Operations: []BatchOperation[string, int]{
				{Type: BatchSet, Key: "new1", Value: 1},
				{Type: BatchDelete, Key: "toDelete"},
				{Type: BatchSet, Key: "existing1", Value: 101},
				{Type: BatchSet, Key: "new2", Value: 2},
			},
		}

		err := sm.ApplyBatch(batch)
		if err != nil {
			t.Errorf("ApplyBatch failed: %v", err)
		}

		expectedValues := map[string]int{
			"existing1": 101,
			"existing2": 200,
			"new1":      1,
			"new2":      2,
		}

		for k, expected := range expectedValues {
			if val, exists := sm.Get(k); !exists || val != expected {
				t.Errorf("Key %s: expected %d, got %v, exists=%v", k, expected, val, exists)
			}
		}

		if _, exists := sm.Get("toDelete"); exists {
			t.Error("Key 'toDelete' should have been deleted")
		}

		if sm.Len() != 4 {
			t.Errorf("Expected length 4, got %d", sm.Len())
		}
	})

	t.Run("Empty Batch Operations", func(t *testing.T) {
		sm := New[string, int](config)
		defer sm.Stop()

		batch := BatchOperations[string, int]{
			Operations: []BatchOperation[string, int]{},
		}

		err := sm.ApplyBatch(batch)
		if err != nil {
			t.Errorf("ApplyBatch failed for empty batch: %v", err)
		}

		if sm.Len() != 0 {
			t.Errorf("Expected length 0, got %d", sm.Len())
		}
	})

	t.Run("Large Batch Operations", func(t *testing.T) {
		sm := New[int, int](config)
		defer sm.Stop()

		batchSize := 1000
		batch := BatchOperations[int, int]{
			Operations: make([]BatchOperation[int, int], batchSize),
		}

		for i := 0; i < batchSize; i++ {
			batch.Operations[i] = BatchOperation[int, int]{
				Type:  BatchSet,
				Key:   i,
				Value: i * 10,
			}
		}

		start := time.Now()
		err := sm.ApplyBatch(batch)
		duration := time.Since(start)

		if err != nil {
			t.Errorf("ApplyBatch failed for large batch: %v", err)
		}

		if sm.Len() != int64(batchSize) {
			t.Errorf("Expected length %d, got %d", batchSize, sm.Len())
		}

		for i := 0; i < batchSize; i += 100 {
			if val, exists := sm.Get(i); !exists || val != i*10 {
				t.Errorf("Key %d: expected %d, got %v, exists=%v", i, i*10, val, exists)
			}
		}

		t.Logf("Large batch operation (%d items) took: %v", batchSize, duration)
	})

	t.Run("Concurrent Batch Operations", func(t *testing.T) {
		sm := New[int, int](config)
		defer sm.Stop()

		numGoroutines := 10
		batchesPerGoroutine := 100
		doneCh := make(chan bool, numGoroutines)

		for g := 0; g < numGoroutines; g++ {
			go func(routine int) {
				base := routine * batchesPerGoroutine
				for i := 0; i < batchesPerGoroutine; i++ {
					batch := BatchOperations[int, int]{
						Operations: []BatchOperation[int, int]{
							{Type: BatchSet, Key: base + i, Value: (base + i) * 10},
						},
					}
					err := sm.ApplyBatch(batch)
					if err != nil {
						t.Errorf("Goroutine %d: ApplyBatch failed: %v", routine, err)
					}
				}
				doneCh <- true
			}(g)
		}

		for i := 0; i < numGoroutines; i++ {
			<-doneCh
		}

		expectedTotal := numGoroutines * batchesPerGoroutine
		if sm.Len() != int64(expectedTotal) {
			t.Errorf("Expected length %d, got %d", expectedTotal, sm.Len())
		}
	})

	t.Run("Batch Operation Metrics", func(t *testing.T) {
		sm := New[string, int](config)
		defer sm.Stop()

		initialMetrics := sm.GetMetrics()

		batch := BatchOperations[string, int]{
			Operations: []BatchOperation[string, int]{
				{Type: BatchSet, Key: "a", Value: 1},
				{Type: BatchSet, Key: "b", Value: 2},
				{Type: BatchSet, Key: "c", Value: 3},
			},
		}

		err := sm.ApplyBatch(batch)
		if err != nil {
			t.Errorf("ApplyBatch failed: %v", err)
		}

		finalMetrics := sm.GetMetrics()

		if finalMetrics.totalItemsProcessed <= initialMetrics.totalItemsProcessed {
			t.Error("Metrics should show increased items processed")
		}

		if finalMetrics.peakSize < 3 {
			t.Errorf("Peak size should be at least 3, got %d", finalMetrics.peakSize)
		}
	})
}
