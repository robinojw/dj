package skills

import (
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"hello world", []string{"hello", "world"}},
		{"hello-world", []string{"hello", "world"}},
		{"hello_world", []string{"hello", "world"}},
		{"Hello, World!", []string{"Hello", "World"}},
		{"test123", []string{"test123"}},
		{"", []string{}},
		{"   spaces   ", []string{"spaces"}},
		{"multi-word-test", []string{"multi", "word", "test"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := tokenize(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("tokenize(%q) got %d words, want %d", tt.input, len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("tokenize(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestIntersect(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want int
	}{
		{"no overlap", []string{"a", "b"}, []string{"c", "d"}, 0},
		{"partial overlap", []string{"a", "b", "c"}, []string{"b", "c", "d"}, 2},
		{"full overlap", []string{"a", "b"}, []string{"a", "b"}, 2},
		{"empty a", []string{}, []string{"a", "b"}, 0},
		{"empty b", []string{"a", "b"}, []string{}, 0},
		{"duplicates in a", []string{"a", "a", "b"}, []string{"a", "b"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intersect(tt.a, tt.b)
			if len(got) != tt.want {
				t.Errorf("intersect() got %d items, want %d", len(got), tt.want)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 2},
		{2, 1, 2},
		{5, 5, 5},
		{-1, 0, 0},
		{0, -1, 0},
	}

	for _, tt := range tests {
		got := max(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("max(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestMatcherScore(t *testing.T) {
	registry := &Registry{
		skills: []Skill{
			{
				Name:                    "refactor",
				Description:             "restructure code while preserving behavior",
				AllowImplicitInvocation: true,
			},
			{
				Name:                    "write-tests",
				Description:             "create comprehensive test suites",
				AllowImplicitInvocation: true,
			},
			{
				Name:                    "no-implicit",
				Description:             "test skill without implicit invocation",
				AllowImplicitInvocation: false,
			},
		},
	}

	matcher := NewMatcher(registry)

	tests := []struct {
		name      string
		prompt    string
		skillName string
		wantScore bool // whether score should be > 0
	}{
		{
			name:      "exact match",
			prompt:    "refactor this code",
			skillName: "refactor",
			wantScore: true,
		},
		{
			name:      "description match",
			prompt:    "restructure the code to preserve behavior",
			skillName: "refactor",
			wantScore: true,
		},
		{
			name:      "no match",
			prompt:    "deploy to production",
			skillName: "refactor",
			wantScore: false,
		},
		{
			name:      "implicit disabled",
			prompt:    "no implicit test skill",
			skillName: "no-implicit",
			wantScore: false,
		},
		{
			name:      "partial description match",
			prompt:    "create tests",
			skillName: "write-tests",
			wantScore: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var skill Skill
			for _, s := range registry.All() {
				if s.Name == tt.skillName {
					skill = s
					break
				}
			}

			score := matcher.Score(tt.prompt, skill)

			if tt.wantScore && score == 0 {
				t.Errorf("expected non-zero score, got 0")
			}
			if !tt.wantScore && score != 0 {
				t.Errorf("expected zero score, got %f", score)
			}
		})
	}
}

func TestMatcherBestMatch(t *testing.T) {
	registry := &Registry{
		skills: []Skill{
			{
				Name:                    "refactor",
				Description:             "restructure code while preserving behavior",
				AllowImplicitInvocation: true,
			},
			{
				Name:                    "write-tests",
				Description:             "create comprehensive test suites",
				AllowImplicitInvocation: true,
			},
			{
				Name:                    "explain-code",
				Description:             "generate detailed code explanations",
				AllowImplicitInvocation: true,
			},
		},
	}

	matcher := NewMatcher(registry)

	tests := []struct {
		name      string
		prompt    string
		wantSkill *string // nil if no match expected
	}{
		{
			name:      "clear refactor match",
			prompt:    "refactor this code to improve structure",
			wantSkill: strPtr("refactor"),
		},
		{
			name:      "clear test match",
			prompt:    "write comprehensive tests for this function",
			wantSkill: strPtr("write-tests"),
		},
		{
			name:      "no match below threshold",
			prompt:    "deploy to production",
			wantSkill: nil,
		},
		{
			name:      "explain match",
			prompt:    "explain what this code does",
			wantSkill: strPtr("explain-code"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skill := matcher.BestMatch(tt.prompt)

			if tt.wantSkill == nil {
				if skill != nil {
					t.Errorf("expected no match, got %q", skill.Name)
				}
			} else {
				if skill == nil {
					t.Errorf("expected match %q, got nil", *tt.wantSkill)
				} else if skill.Name != *tt.wantSkill {
					t.Errorf("expected skill %q, got %q", *tt.wantSkill, skill.Name)
				}
			}
		})
	}
}

func TestMatcherAllMatches(t *testing.T) {
	registry := &Registry{
		skills: []Skill{
			{
				Name:                    "refactor",
				Description:             "restructure code improve quality",
				AllowImplicitInvocation: true,
			},
			{
				Name:                    "optimize",
				Description:             "improve code performance",
				AllowImplicitInvocation: true,
			},
			{
				Name:                    "document",
				Description:             "add documentation comments",
				AllowImplicitInvocation: true,
			},
		},
	}

	matcher := NewMatcher(registry)

	tests := []struct {
		name       string
		prompt     string
		wantCount  int
		wantSorted bool
	}{
		{
			name:       "multiple matches",
			prompt:     "improve code quality and performance",
			wantCount:  2, // refactor and optimize both match "improve" + "code"
			wantSorted: true,
		},
		{
			name:      "no matches",
			prompt:    "deploy to production",
			wantCount: 0,
		},
		{
			name:       "single match",
			prompt:     "add documentation to functions",
			wantCount:  1,
			wantSorted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := matcher.AllMatches(tt.prompt)

			if len(matches) < tt.wantCount {
				t.Errorf("expected at least %d matches, got %d", tt.wantCount, len(matches))
			}

			// Verify sorted by score (descending)
			if tt.wantSorted && len(matches) > 1 {
				for i := 1; i < len(matches); i++ {
					if matches[i].Score > matches[i-1].Score {
						t.Errorf("matches not sorted: matches[%d].Score (%f) > matches[%d].Score (%f)",
							i, matches[i].Score, i-1, matches[i-1].Score)
					}
				}
			}

			// Verify all scores are above threshold
			for i, m := range matches {
				if m.Score <= implicitThreshold {
					t.Errorf("match[%d] score %f is not above threshold %f", i, m.Score, implicitThreshold)
				}
			}
		})
	}
}

func TestScoredSkill(t *testing.T) {
	// Just verify the struct can be created and used
	skill := Skill{
		Name:        "test",
		Description: "test skill",
	}

	scored := ScoredSkill{
		Skill: skill,
		Score: 0.75,
	}

	if scored.Skill.Name != "test" {
		t.Errorf("expected skill name 'test', got %q", scored.Skill.Name)
	}

	if scored.Score != 0.75 {
		t.Errorf("expected score 0.75, got %f", scored.Score)
	}
}

func TestMatcherScoreEmptyDescription(t *testing.T) {
	registry := &Registry{
		skills: []Skill{
			{
				Name:                    "empty-desc",
				Description:             "",
				AllowImplicitInvocation: true,
			},
		},
	}

	matcher := NewMatcher(registry)
	score := matcher.Score("some prompt", registry.skills[0])

	if score != 0 {
		t.Errorf("expected score 0 for empty description, got %f", score)
	}
}

// Helper function
func strPtr(s string) *string {
	return &s
}
