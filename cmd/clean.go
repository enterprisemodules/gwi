package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/enterprisemodules/gwi/internal/config"
	"github.com/enterprisemodules/gwi/internal/git"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove orphaned worktrees and branches",
	Long:  `Prune worktrees that no longer exist and remove branches whose remotes have been deleted.`,
	Run:   runClean,
}

func runClean(cmd *cobra.Command, args []string) {
	cfg := config.Load()
	repoInfo, err := git.GetRepoInfo()
	if err != nil {
		config.Die("%v", err)
	}

	config.Info("Checking for orphaned worktrees...")

	// Prune worktrees that no longer exist on disk
	output, err := git.PruneWorktrees()
	if err == nil && output != "" && output != "nothing to prune\n" {
		fmt.Print(output)
	}

	// Find merged branches that can be cleaned up
	config.Info("Checking for merged branches...")
	if err := git.FetchPrune(); err != nil {
		config.Warn("Failed to fetch: %v", err)
	}

	base := cfg.WorktreeBasePath(repoInfo.Org, repoInfo.Repo)
	var branchesToDelete []string

	// Find local branches that track deleted remotes
	branches, err := git.GetLocalBranches()
	if err != nil {
		config.Die("Failed to list branches: %v", err)
	}

	for _, branch := range branches {
		if branch == "main" || branch == "master" {
			continue
		}

		// Check if branch has a worktree
		worktreePath := filepath.Join(base, branch)
		if _, err := os.Stat(worktreePath); err == nil {
			// Has worktree, skip
			continue
		}

		// Check if remote branch exists
		if !git.RemoteBranchExists(branch) {
			branchesToDelete = append(branchesToDelete, branch+" (remote deleted)")
		}
	}

	if len(branchesToDelete) == 0 {
		config.Success("No orphaned branches found.")
		return
	}

	fmt.Println()
	fmt.Println("Branches to clean up:")
	for _, branch := range branchesToDelete {
		fmt.Printf("  - %s\n", branch)
	}
	fmt.Println()

	if !confirmPrompt("Delete these branches?") {
		return
	}

	for _, branchInfo := range branchesToDelete {
		// Extract just the branch name (remove the reason suffix)
		branch := branchInfo
		for i, c := range branchInfo {
			if c == ' ' {
				branch = branchInfo[:i]
				break
			}
		}
		if err := git.DeleteBranch(branch); err == nil {
			config.Success("Deleted branch: %s", branch)
		}
	}
}
