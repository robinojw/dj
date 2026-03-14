package skills

import (
	"strings"
	"unicode"
)

const implicitThreshold = 0.4

const (
	descriptionWeight = 0.7
	nameWeight        = 0.3
)

// Matcher scores skills against prompts for implicit invocation.
type Matcher struct {
	registry *Registry
}

func NewMatcher(registry *Registry) *Matcher {
	return &Matcher{registry: registry}
}

// Score returns a relevance score between 0 and 1 for a skill against a prompt.
func (m *Matcher) Score(prompt string, skill Skill) float64 {
	if !skill.AllowImplicitInvocation {
		return 0
	}

	promptWords := tokenize(strings.ToLower(prompt))
	descWords := tokenize(strings.ToLower(skill.Description))
	nameWords := tokenize(strings.ToLower(skill.Name))

	if len(descWords) == 0 {
		return 0
	}

	// Score based on word overlap with description
	overlap := intersect(promptWords, descWords)
	descScore := float64(len(overlap)) / float64(len(descWords))

	// Bonus for name match
	nameOverlap := intersect(promptWords, nameWords)
	nameScore := float64(len(nameOverlap)) / float64(max(len(nameWords), 1))

	// Combined score with name getting a bonus
	return descScore*descriptionWeight + nameScore*nameWeight
}

// BestMatch returns the highest-scoring implicit skill above the threshold.
func (m *Matcher) BestMatch(prompt string) *Skill {
	var best *Skill
	bestScore := implicitThreshold

	for i, s := range m.registry.All() {
		if score := m.Score(prompt, s); score > bestScore {
			best = &m.registry.skills[i]
			bestScore = score
		}
	}

	return best
}

// AllMatches returns all skills scoring above the threshold, sorted by score.
func (m *Matcher) AllMatches(prompt string) []ScoredSkill {
	var matches []ScoredSkill

	for _, s := range m.registry.All() {
		score := m.Score(prompt, s)
		if score > implicitThreshold {
			matches = append(matches, ScoredSkill{Skill: s, Score: score})
		}
	}

	// Simple insertion sort (typically very few skills)
	for i := 1; i < len(matches); i++ {
		for j := i; j > 0 && matches[j].Score > matches[j-1].Score; j-- {
			matches[j], matches[j-1] = matches[j-1], matches[j]
		}
	}

	return matches
}

// ScoredSkill pairs a skill with its relevance score.
type ScoredSkill struct {
	Skill Skill
	Score float64
}

func tokenize(s string) []string {
	var words []string
	current := strings.Builder{}

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		}
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}

	return words
}

func intersect(a, b []string) []string {
	bSet := make(map[string]bool, len(b))
	for _, w := range b {
		bSet[w] = true
	}

	var result []string
	seen := make(map[string]bool)
	for _, w := range a {
		if bSet[w] && !seen[w] {
			result = append(result, w)
			seen[w] = true
		}
	}
	return result
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
