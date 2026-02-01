package config

import (
	"os"
	"path/filepath"
)

// Config holds all gwi configuration
type Config struct {
	WorktreeBase  string
	MergeStrategy string
	AutoActivate  bool
	HookDir       string
}

// Load returns the configuration from environment variables with defaults
func Load() *Config {
	home, _ := os.UserHomeDir()

	worktreeBase := os.Getenv("GWI_WORKTREE_BASE")
	if worktreeBase == "" {
		worktreeBase = filepath.Join(home, "worktrees")
	}

	mergeStrategy := os.Getenv("GWI_MERGE_STRATEGY")
	if mergeStrategy == "" {
		mergeStrategy = "squash"
	}

	autoActivate := os.Getenv("GWI_AUTO_ACTIVATE") == "1"

	hookDir := os.Getenv("GWI_HOOK_DIR")
	if hookDir == "" {
		hookDir = filepath.Join(home, ".config", "gwi", "hooks")
	}

	return &Config{
		WorktreeBase:  worktreeBase,
		MergeStrategy: mergeStrategy,
		AutoActivate:  autoActivate,
		HookDir:       hookDir,
	}
}

// WorktreeBasePath returns the worktree base path for a given org/repo
func (c *Config) WorktreeBasePath(org, repo string) string {
	return filepath.Join(c.WorktreeBase, "github.com", org, repo)
}
