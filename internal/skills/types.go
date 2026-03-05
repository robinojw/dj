package skills

// Skill represents a loaded SKILL.md file following the OpenAI Codex spec.
type Skill struct {
	// Frontmatter fields
	Name                    string `yaml:"name"`
	Description             string `yaml:"description"`
	AllowImplicitInvocation bool   `yaml:"allow_implicit_invocation"`

	// Body = instructions injected into system prompt
	Instructions string

	// Optional executable scripts in the same directory
	Scripts []SkillScript

	// Resources (files, templates, examples) in the skill directory
	Resources []SkillResource

	// Source path on disk
	Dir string
}

// SkillScript represents an executable script bundled with a skill.
type SkillScript struct {
	Filename string // e.g. "setup.sh", "validate.py"
	Language string // "bash", "python", "node"
	Path     string // absolute path
}

// SkillResource represents a file resource bundled with a skill.
type SkillResource struct {
	Filename string
	Path     string
}
