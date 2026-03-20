package roster

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

const frontmatterDelimiter = "---"

func LoadPersonas(dir string) ([]PersonaDefinition, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read persona dir: %w", err)
	}

	var personas []PersonaDefinition
	for _, entry := range entries {
		isMarkdown := !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md")
		if !isMarkdown {
			continue
		}
		persona, err := loadPersonaFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("load persona %s: %w", entry.Name(), err)
		}
		personas = append(personas, persona)
	}
	return personas, nil
}

func loadPersonaFile(path string) (PersonaDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PersonaDefinition{}, fmt.Errorf("read file: %w", err)
	}

	frontmatter, body, err := splitFrontmatter(string(data))
	if err != nil {
		return PersonaDefinition{}, fmt.Errorf("parse frontmatter: %w", err)
	}

	var persona PersonaDefinition
	if err := yaml.Unmarshal([]byte(frontmatter), &persona); err != nil {
		return PersonaDefinition{}, fmt.Errorf("unmarshal frontmatter: %w", err)
	}
	persona.Content = strings.TrimSpace(body)
	return persona, nil
}

func splitFrontmatter(content string) (string, string, error) {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, frontmatterDelimiter) {
		return "", "", fmt.Errorf("missing opening frontmatter delimiter")
	}
	rest := trimmed[len(frontmatterDelimiter):]
	closingDelimiter := "\n" + frontmatterDelimiter
	endIndex := strings.Index(rest, closingDelimiter)
	if endIndex == -1 {
		return "", "", fmt.Errorf("missing closing frontmatter delimiter")
	}
	frontmatter := rest[:endIndex]
	body := rest[endIndex+len(closingDelimiter):]
	return frontmatter, body, nil
}

func LoadSignals(path string) (*RepoSignals, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read signals file: %w", err)
	}

	var signals RepoSignals
	if err := json.Unmarshal(data, &signals); err != nil {
		return nil, fmt.Errorf("unmarshal signals: %w", err)
	}
	return &signals, nil
}
