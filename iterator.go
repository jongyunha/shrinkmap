package shrinkmap

// Iterator provides a safe way to iterate over map entries.
// The iterator works with a snapshot of the map taken at creation time,
// making it safe to use even if the map is modified during iteration.
type Iterator[K comparable, V any] struct {
	sm       *ShrinkableMap[K, V]
	snapshot []KeyValue[K, V]
	index    int
}

// NewIterator creates a new iterator for the map.
// The iterator takes a snapshot of the current map state, making it
// safe to use even if the map is modified during iteration.
// The snapshot is taken with a read lock for consistency.
//
// Example:
//
//	sm := shrinkmap.New[string, int](shrinkmap.DefaultConfig())
//	sm.Set("key1", 1)
//	sm.Set("key2", 2)
//
//	iter := sm.NewIterator()
//	for iter.Next() {
//		key, value := iter.Get()
//		fmt.Printf("Key: %s, Value: %d\n", key, value)
//	}
func (sm *ShrinkableMap[K, V]) NewIterator() *Iterator[K, V] {
	return &Iterator[K, V]{
		sm:       sm,
		snapshot: sm.Snapshot(),
		index:    0,
	}
}

// Next advances the iterator to the next item.
// Returns true if there are more items to iterate over, false otherwise.
// Must be called before each Get() call.
func (it *Iterator[K, V]) Next() bool {
	return it.index < len(it.snapshot)
}

// Get returns the current key-value pair and advances the iterator.
// Must be called only after Next() returns true.
// Panics if called when there are no more items.
func (it *Iterator[K, V]) Get() (K, V) {
	if it.index >= len(it.snapshot) {
		panic("iterator: Get() called after Next() returned false")
	}
	item := it.snapshot[it.index]
	it.index++
	return item.Key, item.Value
}
