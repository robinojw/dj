package agents

import (
	"testing"

	"github.com/robinojw/dj/internal/api"
)

func TestShouldCompactBelowThreshold(t *testing.T) {
	c := NewCompactor(nil, 0.60)
	usage := api.Usage{InputTokens: 100_000} // 25% of 400K
	if c.ShouldCompact(usage) {
		t.Error("Should NOT compact at 25%")
	}
}

func TestShouldCompactAboveThreshold(t *testing.T) {
	c := NewCompactor(nil, 0.60)
	usage := api.Usage{InputTokens: 280_000} // 70% of 400K
	if !c.ShouldCompact(usage) {
		t.Error("Should compact at 70%")
	}
}

func TestShouldCompactAtExactThreshold(t *testing.T) {
	c := NewCompactor(nil, 0.60)
	usage := api.Usage{InputTokens: 240_000} // exactly 60%
	if c.ShouldCompact(usage) {
		t.Error("Should NOT compact at exactly the threshold (needs to exceed)")
	}
}

func TestBuildCompactionPrompt(t *testing.T) {
	turns := []Turn{
		{Role: "user", Content: "Fix the login bug"},
		{Role: "assistant", Content: "I'll check auth.go..."},
		{Role: "user", Content: "Also update the tests"},
	}

	prompt := buildCompactionPrompt(turns)
	if prompt == "" {
		t.Error("Expected non-empty compaction prompt")
	}
}
