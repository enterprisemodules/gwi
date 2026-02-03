package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds all gwi configuration
type Config struct {
	WorktreeBase  string       `yaml:"worktree_base"`
	MergeStrategy string       `yaml:"merge_strategy"`
	AutoActivate  bool         `yaml:"auto_activate"`
	HookDir       string       `yaml:"hook_dir"`
	MainBranch    string       `yaml:"main_branch"`
	GitHub        GitHubConfig `yaml:"github"`
	Verbose       bool         `yaml:"verbose"`
}

// GitHubConfig holds GitHub Projects integration settings
type GitHubConfig struct {
	ProjectsEnabled bool   `yaml:"projects_enabled"`
	StatusFieldName string `yaml:"status_field_name"`
	InProgressValue string `yaml:"in_progress_value"`
	InReviewValue   string `yaml:"in_review_value"`
	DoneValue       string `yaml:"done_value"`
	CheckScopes     bool   `yaml:"check_scopes"`
}

// Load returns the configuration from YAML file and environment variables
func Load() *Config {
	home, _ := os.UserHomeDir()

	// Start with defaults
	cfg := &Config{
		WorktreeBase:  filepath.Join(home, "worktrees"),
		MergeStrategy: "squash",
		AutoActivate:  false,
		HookDir:       filepath.Join(home, ".config", "gwi", "hooks"),
		MainBranch:    "main",
		Verbose:       false,
		GitHub: GitHubConfig{
			ProjectsEnabled: true,
			StatusFieldName: "Status",
			InProgressValue: "In Progress",
			InReviewValue:   "In Review",
			DoneValue:       "Done",
			CheckScopes:     true,
		},
	}

	// Try to load from YAML config file
	configPath := filepath.Join(home, ".config", "gwi", "config.yaml")
	if data, err := os.ReadFile(configPath); err == nil {
		_ = yaml.Unmarshal(data, cfg)
	}

	// Override with environment variables (they take precedence)
	if val := os.Getenv("GWI_WORKTREE_BASE"); val != "" {
		cfg.WorktreeBase = val
	}
	if val := os.Getenv("GWI_MERGE_STRATEGY"); val != "" {
		cfg.MergeStrategy = val
	}
	if val := os.Getenv("GWI_AUTO_ACTIVATE"); val == "1" {
		cfg.AutoActivate = true
	}
	if val := os.Getenv("GWI_HOOK_DIR"); val != "" {
		cfg.HookDir = val
	}
	if val := os.Getenv("GWI_MAIN_BRANCH"); val != "" {
		cfg.MainBranch = val
	}
	if val := os.Getenv("GWI_VERBOSE"); val == "1" {
		cfg.Verbose = true
	}

	// GitHub Projects configuration
	if val := os.Getenv("GWI_GITHUB_PROJECTS_ENABLED"); val != "" {
		cfg.GitHub.ProjectsEnabled = val != "false" && val != "0"
	}
	if val := os.Getenv("GWI_GITHUB_STATUS_FIELD"); val != "" {
		cfg.GitHub.StatusFieldName = val
	}
	if val := os.Getenv("GWI_GITHUB_IN_PROGRESS"); val != "" {
		cfg.GitHub.InProgressValue = val
	}
	if val := os.Getenv("GWI_GITHUB_IN_REVIEW"); val != "" {
		cfg.GitHub.InReviewValue = val
	}
	if val := os.Getenv("GWI_GITHUB_DONE"); val != "" {
		cfg.GitHub.DoneValue = val
	}
	if val := os.Getenv("GWI_GITHUB_CHECK_SCOPES"); val == "0" || val == "false" {
		cfg.GitHub.CheckScopes = false
	}

	return cfg
}

// WorktreeBasePath returns the worktree base path for a given org/repo
func (c *Config) WorktreeBasePath(org, repo string) string {
	return filepath.Join(c.WorktreeBase, "github.com", org, repo)
}
