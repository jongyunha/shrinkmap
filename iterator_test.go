package shrinkmap

import (
	"sync"
	"testing"
	"time"
)

func TestIterator(t *testing.T) {
	config := Config{
		InitialCapacity:      10,
		AutoShrinkEnabled:    true,
		ShrinkInterval:       time.Second,
		MinShrinkInterval:    time.Second,
		ShrinkRatio:          0.5,
		CapacityGrowthFactor: 1.5,
	}

	t.Run("Basic Iterator Usage", func(t *testing.T) {
		sm := New[string, int](config)
		defer sm.Stop()

		expected := map[string]int{
			"a": 1,
			"b": 2,
			"c": 3,
		}

		for k, v := range expected {
			sm.Set(k, v)
		}

		iter := sm.NewIterator()
		found := make(map[string]int)

		for iter.Next() {
			k, v := iter.Get()
			found[k] = v
		}

		if len(found) != len(expected) {
			t.Errorf("Expected %d items, found %d", len(expected), len(found))
		}

		for k, v := range expected {
			if foundVal, exists := found[k]; !exists || foundVal != v {
				t.Errorf("Key %s: expected %d, got %v, exists=%v", k, v, foundVal, exists)
			}
		}
	})

	t.Run("Iterator With Concurrent Modifications", func(t *testing.T) {
		sm := New[string, int](config)
		defer sm.Stop()

		initial := map[string]int{
			"a": 1,
			"b": 2,
			"c": 3,
		}
		for k, v := range initial {
			sm.Set(k, v)
		}

		iter := sm.NewIterator()

		sm.Set("d", 4)
		sm.Delete("a")

		found := make(map[string]int)
		for iter.Next() {
			k, v := iter.Get()
			found[k] = v
		}

		if len(found) != len(initial) {
			t.Errorf("Expected %d items, found %d", len(initial), len(found))
		}

		for k, v := range initial {
			if foundVal, exists := found[k]; !exists || foundVal != v {
				t.Errorf("Key %s: expected %d, got %v, exists=%v", k, v, foundVal, exists)
			}
		}

		newIter := sm.NewIterator()
		newFound := make(map[string]int)
		for newIter.Next() {
			k, v := newIter.Get()
			newFound[k] = v
		}

		if _, exists := newFound["d"]; !exists {
			t.Error("New iterator should contain newly added key 'd'")
		}
		if _, exists := newFound["a"]; exists {
			t.Error("New iterator should not contain deleted key 'a'")
		}
	})

	t.Run("Multiple Iterators", func(t *testing.T) {
		sm := New[string, int](config)
		defer sm.Stop()

		for i := 0; i < 5; i++ {
			sm.Set(string(rune('a'+i)), i)
		}

		iter1 := sm.NewIterator()
		iter2 := sm.NewIterator()

		found1 := make(map[string]int)
		found2 := make(map[string]int)

		for iter1.Next() {
			k, v := iter1.Get()
			found1[k] = v
		}

		for iter2.Next() {
			k, v := iter2.Get()
			found2[k] = v
		}

		if len(found1) != len(found2) {
			t.Errorf("Iterators returned different number of items: %d vs %d",
				len(found1), len(found2))
		}

		for k, v := range found1 {
			if v2, exists := found2[k]; !exists || v != v2 {
				t.Errorf("Inconsistency between iterators for key %s: %d vs %d",
					k, v, v2)
			}
		}
	})

	t.Run("Iterator With Large Dataset", func(t *testing.T) {
		sm := New[int, int](config)
		defer sm.Stop()

		itemCount := 10000
		for i := 0; i < itemCount; i++ {
			sm.Set(i, i*10)
		}

		iter := sm.NewIterator()
		count := 0
		for iter.Next() {
			k, v := iter.Get()
			if v != k*10 {
				t.Errorf("Invalid value for key %d: expected %d, got %d",
					k, k*10, v)
			}
			count++
		}

		if count != itemCount {
			t.Errorf("Iterator visited %d items, expected %d", count, itemCount)
		}
	})

	t.Run("Concurrent Iterator Usage", func(t *testing.T) {
		sm := New[int, int](config)
		defer sm.Stop()

		for i := 0; i < 1000; i++ {
			sm.Set(i, i)
		}

		var wg sync.WaitGroup
		numGoroutines := 10

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(routineID int) {
				defer wg.Done()

				iter := sm.NewIterator()
				localSum := 0
				for iter.Next() {
					_, v := iter.Get()
					localSum += v
				}

				expectedSum := (999 * 1000) / 2
				if localSum != expectedSum {
					t.Errorf("Goroutine %d: Invalid sum: got %d, expected %d",
						routineID, localSum, expectedSum)
				}
			}(i)
		}

		go func() {
			for i := 0; i < 100; i++ {
				sm.Set(1000+i, i)
				time.Sleep(time.Millisecond)
			}
		}()

		wg.Wait()
	})

	t.Run("Iterator After Shrink", func(t *testing.T) {
		sm := New[string, int](config)
		defer sm.Stop()

		for i := 0; i < 100; i++ {
			sm.Set(string(rune('a'+i%26)), i)
		}

		for i := 0; i < 50; i++ {
			sm.Delete(string(rune('a' + i%26)))
		}

		sm.ForceShrink()

		iter := sm.NewIterator()
		count := 0
		for iter.Next() {
			k, _ := iter.Get()
			if len(k) != 1 {
				t.Errorf("Invalid key format: %s", k)
			}
			count++
		}

		if int64(count) != sm.Len() {
			t.Errorf("Iterator visited %d items, but map length is %d",
				count, sm.Len())
		}
	})
}
