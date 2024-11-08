package shrinkmap

// Iterator provides a safe way to iterate over map entries
type Iterator[K comparable, V any] struct {
	sm       *ShrinkableMap[K, V]
	snapshot []KeyValue[K, V]
	index    int
}

// NewIterator creates a new iterator for the map
func (sm *ShrinkableMap[K, V]) NewIterator() *Iterator[K, V] {
	return &Iterator[K, V]{
		sm:       sm,
		snapshot: sm.Snapshot(),
		index:    0,
	}
}

func (it *Iterator[K, V]) Next() bool {
	return it.index < len(it.snapshot)
}

func (it *Iterator[K, V]) Get() (K, V) {
	item := it.snapshot[it.index]
	it.index++
	return item.Key, item.Value
}
