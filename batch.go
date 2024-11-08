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

// ApplyBatch applies multiple operations atomically
func (sm *ShrinkableMap[K, V]) ApplyBatch(batch BatchOperations[K, V]) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, op := range batch.Operations {
		switch op.Type {
		case BatchSet:
			_, exists := sm.data[op.Key]
			sm.data[op.Key] = op.Value
			if !exists {
				sm.itemCount.Add(1)
				sm.updateMetrics(1)
			}
		case BatchDelete:
			if _, exists := sm.data[op.Key]; exists {
				delete(sm.data, op.Key)
				sm.deletedCount.Add(1)
			}
		}
	}

	if sm.config.AutoShrinkEnabled {
		go sm.TryShrink()
	}
	return nil
}
