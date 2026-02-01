package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/enterprisemodules/gwi/internal/config"
	"github.com/enterprisemodules/gwi/internal/git"
	"github.com/spf13/cobra"
)

var forceRemove bool

var rmCmd = &cobra.Command{
	Use:   "rm [issue-number]",
	Short: "Delete worktree",
	Long:  `Remove a worktree for the given issue number.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runRm,
}

func init() {
	rmCmd.Flags().BoolVarP(&forceRemove, "force", "f", false, "Force remove even with uncommitted changes")
}

func runRm(cmd *cobra.Command, args []string) {
	cfg := config.Load()
	repoInfo, err := git.GetRepoInfo()
	if err != nil {
		config.Die("%v", err)
	}

	var issueNumber int
	base := cfg.WorktreeBasePath(repoInfo.Org, repoInfo.Repo)

	if len(args) > 0 && args[0] != "--force" && args[0] != "-f" {
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

	// Confirm removal
	fmt.Printf("Remove worktree %s%s%s?\n", config.Yellow(""), worktreeName, config.Yellow(""))
	if !confirmPrompt("Confirm") {
		config.Die("Aborted")
	}

	// If we're inside, output the cd instruction for shell function
	if needCd {
		mainPath, err := git.GetMainWorktreePath()
		if err == nil && mainPath != "" {
			fmt.Printf("__GWI_CD_TO__:%s\n", mainPath)
		}
	}

	config.Info("Removing worktree: %s", worktreePath)

	if err := git.RemoveWorktree(worktreePath, forceRemove); err != nil {
		if !forceRemove {
			config.Die("Worktree has uncommitted changes. Use --force to remove anyway.")
		}
		config.Die("Failed to remove worktree: %v", err)
	}

	config.Success("Worktree removed.")
}
