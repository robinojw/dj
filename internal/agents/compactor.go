package agents

import (
	"context"
	"fmt"
	"strings"

	"github.com/robinojw/dj/internal/api"
)

const contextWindowSize = 400_000

// Turn represents a single conversation turn for compaction.
type Turn struct {
	Role    string
	Content string
}

// Compactor summarises conversation history to free up context window.
type Compactor struct {
	client    api.Client
	threshold float64 // fraction of context window (e.g. 0.60)
}

func NewCompactor(client api.Client, threshold float64) *Compactor {
	if threshold <= 0 {
		threshold = 0.60
	}
	return &Compactor{client: client, threshold: threshold}
}

// ShouldCompact returns true when input tokens exceed the threshold.
func (c *Compactor) ShouldCompact(usage api.Usage) bool {
	return float64(usage.InputTokens)/float64(contextWindowSize) > c.threshold
}

// Compact summarises the conversation history into a compressed memory block.
func (c *Compactor) Compact(ctx context.Context, history []Turn) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("no API client for compaction")
	}

	prompt := buildCompactionPrompt(history)

	req := api.CreateResponseRequest{
		Model:        "o4-mini",
		Input:        api.MakeStringInput(prompt),
		Instructions: compactionInstructions,
		Reasoning: &api.Reasoning{
			Effort: "low",
		},
	}

	resp, err := c.client.Send(ctx, req)
	if err != nil {
		return "", fmt.Errorf("compaction call: %w", err)
	}

	// Extract text from response
	var text string
	for _, item := range resp.Output {
		if item.Content != nil {
			text += string(item.Content)
		}
	}

	return text, nil
}

const compactionInstructions = `Summarise this conversation concisely, preserving:
1. Decisions made and their rationale
2. Files that were read, created, or modified
3. Current task state and next steps
4. Any errors encountered and how they were resolved
5. User preferences expressed during the conversation

Output a structured summary, not a transcript. Be concise but complete.`

func buildCompactionPrompt(turns []Turn) string {
	var sb strings.Builder
	sb.WriteString("Conversation history to summarise:\n\n")
	for _, t := range turns {
		sb.WriteString(fmt.Sprintf("[%s]: %s\n\n", t.Role, t.Content))
	}
	return sb.String()
}
