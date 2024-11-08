package shrinkmap

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

const (
	smallDataset  = 1000
	mediumDataset = 10000
	largeDataset  = 100000
)

var benchConfig = Config{
	InitialCapacity:      1000,
	AutoShrinkEnabled:    true,
	ShrinkInterval:       time.Second,
	MinShrinkInterval:    time.Second,
	ShrinkRatio:          0.5,
	CapacityGrowthFactor: 1.5,
}

func BenchmarkBasicOperations(b *testing.B) {
	sm := New[string, int](benchConfig)
	defer sm.Stop()

	b.Run("Sequential Set", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sm.Set(strconv.Itoa(i), i)
		}
	})

	b.Run("Sequential Get", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = sm.Get(strconv.Itoa(i))
		}
	})

	b.Run("Mixed Set/Get", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				sm.Set(strconv.Itoa(i), i)
			} else {
				_, _ = sm.Get(strconv.Itoa(i - 1))
			}
		}
	})
}

func BenchmarkDatasetSizes(b *testing.B) {
	datasets := []struct {
		name string
		size int
	}{
		{"Small", smallDataset},
		{"Medium", mediumDataset},
		{"Large", largeDataset},
	}

	for _, ds := range datasets {
		b.Run(fmt.Sprintf("Dataset_%s", ds.name), func(b *testing.B) {
			sm := New[int, int](benchConfig)
			defer sm.Stop()

			for i := 0; i < ds.size; i++ {
				sm.Set(i, i)
			}

			b.ResetTimer()

			b.Run("Random_Access", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					key := rand.Intn(ds.size)
					_, _ = sm.Get(key)
				}
			})

			b.Run("Random_Update", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					key := rand.Intn(ds.size)
					sm.Set(key, i)
				}
			})

			b.Run("Random_Delete", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					key := rand.Intn(ds.size)
					sm.Delete(key)
				}
			})
		})
	}
}

func BenchmarkBatchOperations(b *testing.B) {
	batchSizes := []int{10, 100, 1000}

	for _, size := range batchSizes {
		b.Run(fmt.Sprintf("BatchSize_%d", size), func(b *testing.B) {
			sm := New[int, int](benchConfig)
			defer sm.Stop()

			batch := BatchOperations[int, int]{
				Operations: make([]BatchOperation[int, int], size),
			}

			b.ResetTimer()

			b.Run("BatchSet", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					for j := 0; j < size; j++ {
						batch.Operations[j] = BatchOperation[int, int]{
							Type:  BatchSet,
							Key:   i*size + j,
							Value: j,
						}
					}
					_ = sm.ApplyBatch(batch)
				}
			})

			b.Run("BatchMixed", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					for j := 0; j < size; j++ {
						if j%2 == 0 {
							batch.Operations[j] = BatchOperation[int, int]{
								Type:  BatchSet,
								Key:   j,
								Value: j,
							}
						} else {
							batch.Operations[j] = BatchOperation[int, int]{
								Type: BatchDelete,
								Key:  j - 1,
							}
						}
					}
					_ = sm.ApplyBatch(batch)
				}
			})
		})
	}
}

func BenchmarkConcurrency(b *testing.B) {
	sm := New[int, int](benchConfig)
	defer sm.Stop()

	parallelCount := []int{2, 4, 8, 16}

	for _, count := range parallelCount {
		b.Run(fmt.Sprintf("Parallel_%d", count), func(b *testing.B) {
			b.SetParallelism(count)

			b.RunParallel(func(pb *testing.PB) {
				localCounter := 0
				for pb.Next() {
					key := localCounter % 1000
					switch localCounter % 3 {
					case 0:
						sm.Set(key, localCounter)
					case 1:
						_, _ = sm.Get(key)
					case 2:
						sm.Delete(key)
					}
					localCounter++
				}
			})
		})
	}
}

func BenchmarkShrinking(b *testing.B) {
	sm := New[int, int](benchConfig)
	defer sm.Stop()

	for i := 0; i < largeDataset; i++ {
		sm.Set(i, i)
	}

	for i := 0; i < largeDataset/2; i++ {
		sm.Delete(i)
	}

	b.Run("ForceShrink", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sm.ForceShrink()
		}
	})

	b.Run("TryShrink", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sm.TryShrink()
		}
	})
}
