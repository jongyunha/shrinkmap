package shrinkmap

// BatchOperations provides batch operation capabilities
type BatchOperations[K comparable, V any] struct {
	Operations []BatchOperation[K, V]
}

type BatchOperation[K comparable, V any] struct {
	Type  BatchOpType
	Key   K
	Value V
}

type BatchOpType int

const (
	BatchSet BatchOpType = iota
	BatchDelete
)

// ApplyBatch applies multiple operations atomically.
// This method is more efficient than individual operations when
// processing multiple items as it acquires the lock only once.
// Returns an error if the map is stopped or capacity is exceeded.
//
// Example:
//
//	batch := shrinkmap.BatchOperations[string, int]{
//		Operations: []shrinkmap.BatchOperation[string, int]{
//			{Type: shrinkmap.BatchSet, Key: "key1", Value: 1},
//			{Type: shrinkmap.BatchSet, Key: "key2", Value: 2},
//			{Type: shrinkmap.BatchDelete, Key: "key3"},
//		},
//	}
//	if err := sm.ApplyBatch(batch); err != nil {
//		log.Printf("Batch operation failed: %v", err)
//	}
func (sm *ShrinkableMap[K, V]) ApplyBatch(batch BatchOperations[K, V]) error {
	if sm.stopped.Load() {
		return ErrMapStopped
	}

	if len(batch.Operations) == 0 {
		return nil
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Count new items to check capacity
	newItemCount := int64(0)
	for _, op := range batch.Operations {
		if op.Type == BatchSet {
			if _, exists := sm.data[op.Key]; !exists {
				newItemCount++
			}
		}
	}

	// Check capacity before applying any operations
	if sm.config.MaxMapSize > 0 && sm.itemCount.Load()+newItemCount > int64(sm.config.MaxMapSize) {
		return ErrCapacityExceeded.WithDetails("currentSize", sm.itemCount.Load()).WithDetails("maxSize", sm.config.MaxMapSize)
	}

	// Apply operations
	itemsAdded := int64(0)
	for _, op := range batch.Operations {
		switch op.Type {
		case BatchSet:
			_, exists := sm.data[op.Key]
			sm.data[op.Key] = op.Value
			if !exists {
				sm.itemCount.Add(1)
				itemsAdded++
			}
		case BatchDelete:
			if _, exists := sm.data[op.Key]; exists {
				delete(sm.data, op.Key)
				sm.deletedCount.Add(1)
			}
		}
	}

	if itemsAdded > 0 {
		sm.updateMetrics(itemsAdded)
	}

	if sm.config.AutoShrinkEnabled {
		go sm.TryShrink()
	}
	return nil
}
