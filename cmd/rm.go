package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/enterprisemodules/gwi/internal/config"
	"github.com/enterprisemodules/gwi/internal/git"
	"github.com/spf13/cobra"
)

var forceRemove bool
var skipConfirm bool
var deleteBranch bool

var rmCmd = &cobra.Command{
	Use:   "rm [issue-number]",
	Short: "Delete worktree",
	Long:  `Remove a worktree for the given issue number.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runRm,
}

func init() {
	rmCmd.Flags().BoolVarP(&forceRemove, "force", "f", false, "Force remove even with uncommitted changes")
	rmCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt")
	rmCmd.Flags().BoolVarP(&deleteBranch, "delete-branch", "D", false, "Also delete the local and remote branch")
}

func runRm(cmd *cobra.Command, args []string) {
	cfg := config.Load()
	repoInfo, err := git.GetRepoInfo()
	if err != nil {
		config.Die("%v", err)
	}

	var issueNumber int
	base := cfg.WorktreeBasePath(repoInfo.Org, repoInfo.Repo)

	if len(args) > 0 && args[0] != "--force" && args[0] != "-f" && args[0] != "--yes" && args[0] != "-y" {
		issueNumber, err = strconv.Atoi(args[0])
		if err != nil {
			config.Die("Invalid issue number: %s", args[0])
		}
	} else {
		// Always show interactive selection for rm
		issueNumber, err = selectWorktree(repoInfo, cfg)
		if err != nil {
			config.Die("No worktree selected")
		}
	}

	worktreePath := git.FindWorktreeByIssue(base, issueNumber)
	if worktreePath == "" {
		config.Die("No worktree found for issue #%d", issueNumber)
	}

	worktreeName := filepath.Base(worktreePath)

	// Check if we're inside the worktree
	needCd := false
	if git.IsInsideWorktree(worktreePath) {
		config.Warn("You are inside the worktree you want to remove")
		needCd = true
	}

	// Confirm removal (unless --yes flag is set)
	if !skipConfirm {
		if deleteBranch {
			fmt.Fprintf(os.Stderr, "Remove worktree %s%s%s and delete branch?\n", config.Yellow(""), worktreeName, config.Yellow(""))
		} else {
			fmt.Fprintf(os.Stderr, "Remove worktree %s%s%s?\n", config.Yellow(""), worktreeName, config.Yellow(""))
		}
		if !confirmPrompt("Confirm") {
			config.Die("Aborted")
		}
	}

	// If we're inside, output the cd instruction for shell function
	if needCd {
		mainPath, err := git.GetMainWorktreePath()
		if err == nil && mainPath != "" {
			fmt.Printf("__GWI_CD_TO__:%s\n", mainPath)
		}
	}

	// Get branch name before removing (it's the same as the worktree directory name)
	branchName := worktreeName

	config.Info("Removing worktree: %s", worktreePath)

	if err := git.RemoveWorktree(worktreePath, forceRemove); err != nil {
		if !forceRemove {
			config.Die("Worktree has uncommitted changes. Use --force to remove anyway.")
		}
		config.Die("Failed to remove worktree: %v", err)
	}

	// Prune any stale worktree entries to ensure clean state
	git.PruneWorktrees()

	config.Success("Worktree removed.")

	// Delete branches if requested
	if deleteBranch {
		// Delete local branch
		if git.BranchExists(branchName) {
			config.Info("Deleting local branch: %s", branchName)
			if err := git.DeleteBranch(branchName); err != nil {
				config.Error("Failed to delete local branch: %v", err)
			} else {
				config.Success("Local branch deleted.")
			}
		}

		// Delete remote branch
		if git.RemoteBranchExists(branchName) {
			config.Info("Deleting remote branch: %s", branchName)
			if err := git.DeleteRemoteBranch(branchName); err != nil {
				config.Error("Failed to delete remote branch: %v", err)
			} else {
				config.Success("Remote branch deleted.")
			}
		}
	}
}
