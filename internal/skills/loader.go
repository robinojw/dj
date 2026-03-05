package skills

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Registry holds all loaded skills.
type Registry struct {
	skills []Skill
	paths  []string
}

func NewRegistry(searchPaths []string) *Registry {
	return &Registry{paths: searchPaths}
}

// Load scans all search paths for SKILL.md files and loads them.
func (r *Registry) Load() error {
	r.skills = nil
	seen := make(map[string]bool)

	for _, base := range r.paths {
		base = expandPath(base)
		entries, err := os.ReadDir(base)
		if err != nil {
			continue // skip missing directories
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			skillDir := filepath.Join(base, entry.Name())
			skillFile := filepath.Join(skillDir, "SKILL.md")

			if _, err := os.Stat(skillFile); err != nil {
				continue
			}

			skill, err := parseSkillFile(skillFile, skillDir)
			if err != nil {
				continue
			}

			// Deduplicate by name (first found wins, matching priority order)
			if seen[skill.Name] {
				continue
			}
			seen[skill.Name] = true

			// Discover scripts and resources
			skill.Scripts = discoverScripts(skillDir)
			skill.Resources = discoverResources(skillDir)

			r.skills = append(r.skills, skill)
		}
	}

	return nil
}

// All returns all loaded skills.
func (r *Registry) All() []Skill {
	return r.skills
}

// ByName looks up a skill by name.
func (r *Registry) ByName(name string) *Skill {
	for i, s := range r.skills {
		if s.Name == name {
			return &r.skills[i]
		}
	}
	return nil
}

func parseSkillFile(path, dir string) (Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Skill{}, err
	}

	content := string(data)
	skill := Skill{Dir: dir}

	// Parse YAML frontmatter between --- delimiters
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content[3:], "---", 2)
		if len(parts) == 2 {
			if err := yaml.Unmarshal([]byte(parts[0]), &skill); err != nil {
				return Skill{}, err
			}
			skill.Instructions = strings.TrimSpace(parts[1])
		}
	} else {
		// No frontmatter — entire file is instructions
		skill.Instructions = strings.TrimSpace(content)
		// Derive name from directory
		skill.Name = filepath.Base(dir)
	}

	return skill, nil
}

func discoverScripts(dir string) []SkillScript {
	var scripts []SkillScript
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "SKILL.md" {
			continue
		}
		ext := filepath.Ext(entry.Name())
		lang := ""
		switch ext {
		case ".sh", ".bash":
			lang = "bash"
		case ".py":
			lang = "python"
		case ".js":
			lang = "node"
		default:
			// Check for shebang
			lang = detectShebang(filepath.Join(dir, entry.Name()))
		}

		if lang != "" {
			scripts = append(scripts, SkillScript{
				Filename: entry.Name(),
				Language: lang,
				Path:     filepath.Join(dir, entry.Name()),
			})
		}
	}
	return scripts
}

func discoverResources(dir string) []SkillResource {
	var resources []SkillResource
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || name == "SKILL.md" {
			continue
		}
		ext := filepath.Ext(name)
		// Skip scripts, include everything else as resources
		switch ext {
		case ".sh", ".bash", ".py", ".js":
			continue
		}
		resources = append(resources, SkillResource{
			Filename: name,
			Path:     filepath.Join(dir, name),
		})
	}
	return resources
}

func detectShebang(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#!") {
			switch {
			case strings.Contains(line, "bash"), strings.Contains(line, "/sh"):
				return "bash"
			case strings.Contains(line, "python"):
				return "python"
			case strings.Contains(line, "node"):
				return "node"
			}
		}
	}
	return ""
}

func expandPath(path string) string {
	if len(path) > 1 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
