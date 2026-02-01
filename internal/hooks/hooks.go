package hooks

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/enterprisemodules/gwi/internal/config"
	"github.com/enterprisemodules/gwi/internal/git"
)

// FindHook searches for a hook script in the standard locations
func FindHook(hookName, worktreePath string, cfg *config.Config, repoInfo *git.RepoInfo) string {
	// 1. Check worktree-specific hook
	worktreeHook := filepath.Join(worktreePath, ".gwi", hookName)
	if isExecutable(worktreeHook) {
		return worktreeHook
	}

	// 2. Check main repo hook
	mainPath, err := git.GetMainWorktreePath()
	if err == nil && mainPath != "" {
		mainHook := filepath.Join(mainPath, ".gwi", hookName)
		if isExecutable(mainHook) {
			return mainHook
		}
	}

	// 3. Check global config hook
	if repoInfo != nil {
		globalHook := filepath.Join(cfg.HookDir, repoInfo.Org, repoInfo.Repo, hookName)
		if isExecutable(globalHook) {
			return globalHook
		}
	}

	return ""
}

// isExecutable checks if a file exists and is executable
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode()&0111 != 0
}

// RunHook executes a hook script
func RunHook(hookName, worktreePath string, cfg *config.Config, repoInfo *git.RepoInfo) (bool, error) {
	hookScript := FindHook(hookName, worktreePath, cfg, repoInfo)
	if hookScript == "" {
		return false, nil
	}

	config.Info("Running %s hook...", hookName)

	cmd := exec.Command(hookScript)
	cmd.Dir = worktreePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		config.Warn("Hook exited with error: %v", err)
		return true, err
	}

	config.Success("Hook completed")
	return true, nil
}
