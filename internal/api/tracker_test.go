package api

import (
	"sync"
	"testing"
)

func TestNewTracker(t *testing.T) {
	tracker := NewTracker("gpt-5.4")

	if tracker.Model() != "gpt-5.4" {
		t.Errorf("expected model 'gpt-5.4', got %q", tracker.Model())
	}

	if tracker.InputTokens() != 0 {
		t.Errorf("expected 0 input tokens, got %d", tracker.InputTokens())
	}

	if tracker.OutputTokens() != 0 {
		t.Errorf("expected 0 output tokens, got %d", tracker.OutputTokens())
	}

	if tracker.Cost() != 0 {
		t.Errorf("expected 0 cost, got %f", tracker.Cost())
	}
}

func TestTrackerRecord(t *testing.T) {
	tests := []struct {
		name           string
		model          string
		usage          Usage
		wantInputCost  float64
		wantOutputCost float64
	}{
		{
			name:  "gpt-5.4",
			model: "gpt-5.4",
			usage: Usage{
				InputTokens:  1_000_000,
				OutputTokens: 1_000_000,
			},
			wantInputCost:  2.50,  // 1M tokens * $2.50 per 1M
			wantOutputCost: 10.00, // 1M tokens * $10.00 per 1M
		},
		{
			name:  "gpt-5.1-codex-mini",
			model: "gpt-5.1-codex-mini",
			usage: Usage{
				InputTokens:  500_000,
				OutputTokens: 250_000,
			},
			wantInputCost:  0.75, // 500k tokens * $1.50 per 1M
			wantOutputCost: 1.50, // 250k tokens * $6.00 per 1M
		},
		{
			name:  "o3-pro",
			model: "o3-pro",
			usage: Usage{
				InputTokens:  100_000,
				OutputTokens: 100_000,
			},
			wantInputCost:  2.00, // 100k tokens * $20.00 per 1M
			wantOutputCost: 8.00, // 100k tokens * $80.00 per 1M
		},
		{
			name:  "unknown model fallback",
			model: "unknown-model",
			usage: Usage{
				InputTokens:  1_000_000,
				OutputTokens: 1_000_000,
			},
			wantInputCost:  2.00, // fallback: $2.00 per 1M
			wantOutputCost: 8.00, // fallback: $8.00 per 1M
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewTracker(tt.model)
			tracker.Record(tt.usage)

			if tracker.InputTokens() != tt.usage.InputTokens {
				t.Errorf("expected %d input tokens, got %d", tt.usage.InputTokens, tracker.InputTokens())
			}

			if tracker.OutputTokens() != tt.usage.OutputTokens {
				t.Errorf("expected %d output tokens, got %d", tt.usage.OutputTokens, tracker.OutputTokens())
			}

			expectedCost := tt.wantInputCost + tt.wantOutputCost
			if diff := tracker.Cost() - expectedCost; diff > 0.001 || diff < -0.001 {
				t.Errorf("expected cost %.2f, got %.2f", expectedCost, tracker.Cost())
			}
		})
	}
}

func TestTrackerRecordMultiple(t *testing.T) {
	tracker := NewTracker("gpt-5.4")

	// Record multiple usages
	tracker.Record(Usage{InputTokens: 1000, OutputTokens: 500})
	tracker.Record(Usage{InputTokens: 2000, OutputTokens: 1000})
	tracker.Record(Usage{InputTokens: 1500, OutputTokens: 750})

	expectedInput := 4500
	expectedOutput := 2250

	if tracker.InputTokens() != expectedInput {
		t.Errorf("expected %d input tokens, got %d", expectedInput, tracker.InputTokens())
	}

	if tracker.OutputTokens() != expectedOutput {
		t.Errorf("expected %d output tokens, got %d", expectedOutput, tracker.OutputTokens())
	}

	// Cost calculation: 4500 input * $2.50/1M + 2250 output * $10.00/1M
	expectedCost := (4500.0 / 1_000_000 * 2.50) + (2250.0 / 1_000_000 * 10.00)
	if diff := tracker.Cost() - expectedCost; diff > 0.001 || diff < -0.001 {
		t.Errorf("expected cost %.6f, got %.6f", expectedCost, tracker.Cost())
	}
}

func TestTrackerSetModel(t *testing.T) {
	tracker := NewTracker("gpt-5.4")

	tracker.SetModel("o4-mini")

	if tracker.Model() != "o4-mini" {
		t.Errorf("expected model 'o4-mini', got %q", tracker.Model())
	}
}

func TestTrackerConcurrency(t *testing.T) {
	tracker := NewTracker("gpt-5.4")

	// Run concurrent operations to test mutex protection
	var wg sync.WaitGroup
	iterations := 100

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tracker.Record(Usage{InputTokens: 100, OutputTokens: 50})
		}()
	}

	wg.Wait()

	expectedInput := iterations * 100
	expectedOutput := iterations * 50

	if tracker.InputTokens() != expectedInput {
		t.Errorf("expected %d input tokens, got %d", expectedInput, tracker.InputTokens())
	}

	if tracker.OutputTokens() != expectedOutput {
		t.Errorf("expected %d output tokens, got %d", expectedOutput, tracker.OutputTokens())
	}
}

func TestCycleModels(t *testing.T) {
	if len(CycleModels) == 0 {
		t.Error("CycleModels should not be empty")
	}

	// Verify all cycle models have pricing
	for _, model := range CycleModels {
		if _, ok := modelPricing[model]; !ok {
			t.Errorf("cycle model %q is missing pricing information", model)
		}
	}
}

func TestModelPricing(t *testing.T) {
	// Verify all models have valid pricing
	for model, pricing := range modelPricing {
		if pricing[0] <= 0 {
			t.Errorf("model %q has invalid input pricing: %f", model, pricing[0])
		}
		if pricing[1] <= 0 {
			t.Errorf("model %q has invalid output pricing: %f", model, pricing[1])
		}
		if pricing[1] <= pricing[0] {
			t.Errorf("model %q output pricing (%f) should be higher than input pricing (%f)",
				model, pricing[1], pricing[0])
		}
	}
}
