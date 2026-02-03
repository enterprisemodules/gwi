package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/enterprisemodules/gwi/internal/config"
	"github.com/enterprisemodules/gwi/internal/git"
	"github.com/enterprisemodules/gwi/internal/github"
	"github.com/spf13/cobra"
)

var mergeCmd = &cobra.Command{
	Use:   "merge [issue-number]",
	Short: "Merge worktree into main and cleanup",
	Long:  `Merge the worktree branch into main, close the issue with commit info, and remove the worktree.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runMerge,
}

func runMerge(cmd *cobra.Command, args []string) {
	cfg := config.Load()
	repoInfo, err := git.GetRepoInfo()
	if err != nil {
		config.Die("%v", err)
	}

	var issueNumber int
	base := cfg.WorktreeBasePath(repoInfo.Org, repoInfo.Repo)

	if len(args) > 0 {
		issueNumber, err = strconv.Atoi(args[0])
		if err != nil {
			config.Die("Invalid issue number: %s", args[0])
		}
	} else {
		// Try to detect from current directory
		if num, ok := git.DetectIssueNumber(base); ok {
			issueNumber = num
		} else {
			// Interactive selection
			issueNumber, err = selectWorktree(repoInfo, cfg)
			if err != nil {
				config.Die("No worktree selected")
			}
		}
	}

	worktreePath := git.FindWorktreeByIssue(base, issueNumber)
	if worktreePath == "" {
		config.Die("No worktree found for issue #%d", issueNumber)
	}

	branchName := filepath.Base(worktreePath)

	// Check for uncommitted changes
	if git.HasUncommittedChanges(worktreePath) {
		config.Die("Worktree has uncommitted changes. Commit or stash them first.")
	}

	// Get the main worktree path
	mainWorktree, err := git.GetMainWorktreePath()
	if err != nil {
		config.Die("Failed to get main worktree: %v", err)
	}

	// Get the last commit message before merge
	lastCommitMsg, _ := git.GetLastCommitMessage("")

	mainBranch := cfg.MainBranch

	// Checkout main branch
	config.Info("Switching to %s branch...", mainBranch)
	if err := git.Checkout(mainWorktree, mainBranch); err != nil {
		config.Die("Failed to checkout %s: %v", mainBranch, err)
	}

	// Merge the worktree branch
	config.Info("Merging %s into %s...", branchName, mainBranch)
	if err := git.MergeBranch(mainWorktree, branchName); err != nil {
		config.Die("Merge failed: %v", err)
	}

	// Push main to origin
	config.Info("Pushing %s to origin...", mainBranch)
	if err := git.PushMain(mainWorktree, mainBranch); err != nil {
		config.Die("Failed to push: %v", err)
	}

	// Close the issue with the commit message
	config.Info("Closing issue #%d...", issueNumber)
	comment := fmt.Sprintf("**Merged into %s**\n\n%s", mainBranch, lastCommitMsg)
	if err := github.CloseIssue(issueNumber, comment); err != nil {
		config.Warn("Failed to close issue: %v", err)
	}

	// Update GitHub Project status to "Done"
	if cfg.GitHub.ProjectsEnabled {
		if err := github.UpdateIssueStatus(issueNumber, cfg.GitHub.DoneValue, cfg); err != nil {
			if cfg.Verbose {
				config.Warn("Failed to update project status: %v", err)
			}
		} else {
			config.Info("Updated issue #%d to '%s' in GitHub Projects", issueNumber, cfg.GitHub.DoneValue)
		}
	}

	// Remove worktree
	config.Info("Removing worktree...")
	if err := git.RemoveWorktree(worktreePath, false); err != nil {
		// Try force remove
		git.RemoveWorktree(worktreePath, true)
	}
	// Prune any stale worktree entries
	git.PruneWorktrees()

	// Clean up local and remote branch
	config.Info("Deleting branch %s...", branchName)
	git.DeleteBranch(branchName)
	git.DeleteRemoteBranch(branchName)

	config.Success("Merged into %s and cleaned up!", mainBranch)

	// Output cd marker for shell integration
	fmt.Printf("__GWI_CD_TO__:%s\n", mainWorktree)
}
