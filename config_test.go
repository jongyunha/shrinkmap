package shrinkmap

import (
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultConfig()

		if config.ShrinkInterval != 5*time.Minute {
			t.Errorf("Expected ShrinkInterval to be 5 minutes, got %v", config.ShrinkInterval)
		}
		if config.ShrinkRatio != 0.25 {
			t.Errorf("Expected ShrinkRatio to be 0.25, got %v", config.ShrinkRatio)
		}
		if config.InitialCapacity != 16 {
			t.Errorf("Expected InitialCapacity to be 16, got %v", config.InitialCapacity)
		}
		if !config.AutoShrinkEnabled {
			t.Error("Expected AutoShrinkEnabled to be true")
		}
		if config.MinShrinkInterval != 30*time.Second {
			t.Errorf("Expected MinShrinkInterval to be 30 seconds, got %v", config.MinShrinkInterval)
		}
		if config.MaxMapSize != 1_000_000 {
			t.Errorf("Expected MaxMapSize to be 1,000,000, got %v", config.MaxMapSize)
		}
		if config.CapacityGrowthFactor != 1.2 {
			t.Errorf("Expected CapacityGrowthFactor to be 1.2, got %v", config.CapacityGrowthFactor)
		}
	})

	t.Run("Config validation", func(t *testing.T) {
		tests := []struct {
			name        string
			config      Config
			expectError bool
		}{
			{
				name:        "valid config",
				config:      DefaultConfig(),
				expectError: false,
			},
			{
				name: "invalid shrink interval",
				config: Config{
					ShrinkInterval:       0,
					ShrinkRatio:          0.5,
					InitialCapacity:      16,
					AutoShrinkEnabled:    true,
					MinShrinkInterval:    30 * time.Second,
					MaxMapSize:           1000,
					CapacityGrowthFactor: 1.2,
				},
				expectError: true,
			},
			{
				name: "invalid shrink ratio - zero",
				config: Config{
					ShrinkInterval:       5 * time.Minute,
					ShrinkRatio:          0,
					InitialCapacity:      16,
					AutoShrinkEnabled:    true,
					MinShrinkInterval:    30 * time.Second,
					MaxMapSize:           1000,
					CapacityGrowthFactor: 1.2,
				},
				expectError: true,
			},
			{
				name: "invalid shrink ratio - one",
				config: Config{
					ShrinkInterval:       5 * time.Minute,
					ShrinkRatio:          1.0,
					InitialCapacity:      16,
					AutoShrinkEnabled:    true,
					MinShrinkInterval:    30 * time.Second,
					MaxMapSize:           1000,
					CapacityGrowthFactor: 1.2,
				},
				expectError: true,
			},
			{
				name: "invalid initial capacity",
				config: Config{
					ShrinkInterval:       5 * time.Minute,
					ShrinkRatio:          0.5,
					InitialCapacity:      -1,
					AutoShrinkEnabled:    true,
					MinShrinkInterval:    30 * time.Second,
					MaxMapSize:           1000,
					CapacityGrowthFactor: 1.2,
				},
				expectError: true,
			},
			{
				name: "invalid min shrink interval",
				config: Config{
					ShrinkInterval:       5 * time.Minute,
					ShrinkRatio:          0.5,
					InitialCapacity:      16,
					AutoShrinkEnabled:    true,
					MinShrinkInterval:    0,
					MaxMapSize:           1000,
					CapacityGrowthFactor: 1.2,
				},
				expectError: true,
			},
			{
				name: "invalid max map size",
				config: Config{
					ShrinkInterval:       5 * time.Minute,
					ShrinkRatio:          0.5,
					InitialCapacity:      16,
					AutoShrinkEnabled:    true,
					MinShrinkInterval:    30 * time.Second,
					MaxMapSize:           -1,
					CapacityGrowthFactor: 1.2,
				},
				expectError: true,
			},
			{
				name: "invalid capacity growth factor",
				config: Config{
					ShrinkInterval:       5 * time.Minute,
					ShrinkRatio:          0.5,
					InitialCapacity:      16,
					AutoShrinkEnabled:    true,
					MinShrinkInterval:    30 * time.Second,
					MaxMapSize:           1000,
					CapacityGrowthFactor: 1.0,
				},
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.config.Validate()
				if tt.expectError && err == nil {
					t.Error("Expected validation error but got none")
				}
				if !tt.expectError && err != nil {
					t.Errorf("Expected no validation error but got: %v", err)
				}
			})
		}
	})

	t.Run("Config builder methods", func(t *testing.T) {
		config := DefaultConfig().
			WithShrinkInterval(10 * time.Minute).
			WithShrinkRatio(0.5).
			WithInitialCapacity(32).
			WithAutoShrinkEnabled(false).
			WithMinShrinkInterval(1 * time.Minute).
			WithMaxMapSize(2000).
			WithCapacityGrowthFactor(1.5)

		if config.ShrinkInterval != 10*time.Minute {
			t.Errorf("Expected ShrinkInterval to be 10 minutes, got %v", config.ShrinkInterval)
		}
		if config.ShrinkRatio != 0.5 {
			t.Errorf("Expected ShrinkRatio to be 0.5, got %v", config.ShrinkRatio)
		}
		if config.InitialCapacity != 32 {
			t.Errorf("Expected InitialCapacity to be 32, got %v", config.InitialCapacity)
		}
		if config.AutoShrinkEnabled {
			t.Error("Expected AutoShrinkEnabled to be false")
		}
		if config.MinShrinkInterval != 1*time.Minute {
			t.Errorf("Expected MinShrinkInterval to be 1 minute, got %v", config.MinShrinkInterval)
		}
		if config.MaxMapSize != 2000 {
			t.Errorf("Expected MaxMapSize to be 2000, got %v", config.MaxMapSize)
		}
		if config.CapacityGrowthFactor != 1.5 {
			t.Errorf("Expected CapacityGrowthFactor to be 1.5, got %v", config.CapacityGrowthFactor)
		}
	})
}
