package roster

type PersonaDefinition struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Triggers    []string `yaml:"triggers"`
	Content     string   `yaml:"-"`
}

type RepoSignals struct {
	RepoName   string   `json:"repo_name"`
	Languages  []string `json:"languages"`
	Frameworks []string `json:"frameworks"`
	CIProvider string   `json:"ci_provider,omitempty"`
	LintConfig string   `json:"lint_config,omitempty"`
	IsMonorepo bool     `json:"is_monorepo"`
	HasDocker  bool     `json:"has_docker"`
	HasE2E     bool     `json:"has_e2e"`
	FileCount  int      `json:"file_count"`
}
