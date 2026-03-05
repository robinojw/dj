package api

import (
	"sync"
)

// Model pricing per 1M tokens (USD).
var modelPricing = map[string][2]float64{
	"gpt-5.1-codex-mini": {1.50, 6.00},   // {input, output}
	"gpt-5.1-codex":      {3.00, 12.00},
	"o3-pro":             {20.00, 80.00},
}

// Tracker accumulates token counts and cost across a session.
type Tracker struct {
	mu           sync.Mutex
	model        string
	inputTokens  int
	outputTokens int
	cost         float64
}

func NewTracker(model string) *Tracker {
	return &Tracker{model: model}
}

func (t *Tracker) Record(usage Usage) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.inputTokens += usage.InputTokens
	t.outputTokens += usage.OutputTokens

	pricing, ok := modelPricing[t.model]
	if !ok {
		pricing = [2]float64{2.00, 8.00} // fallback
	}
	t.cost += float64(usage.InputTokens) / 1_000_000 * pricing[0]
	t.cost += float64(usage.OutputTokens) / 1_000_000 * pricing[1]
}

func (t *Tracker) InputTokens() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.inputTokens
}

func (t *Tracker) OutputTokens() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.outputTokens
}

func (t *Tracker) Cost() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.cost
}

func (t *Tracker) SetModel(model string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.model = model
}
