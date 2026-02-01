package cmd

import (
	"fmt"

	"github.com/enterprisemodules/gwi/internal/config"
	"github.com/enterprisemodules/gwi/internal/git"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Select open issue and create worktree",
	Long:  `Interactively select an open issue and create a worktree for it. This command is meant to be used with shell integration.`,
	Run:   runStart,
}

var internalStartCmd = &cobra.Command{
	Use:    "_start",
	Hidden: true,
	Run:    runInternalStart,
}

func runStart(cmd *cobra.Command, args []string) {
	// Public start just calls internal start for now
	runInternalStart(cmd, args)
}

func runInternalStart(cmd *cobra.Command, args []string) {
	cfg := config.Load()
	repoInfo, err := git.GetRepoInfo()
	if err != nil {
		config.Die("%v", err)
	}

	issueNumber, err := selectIssue(repoInfo)
	if err != nil {
		config.Die("No issue selected")
	}

	// Create worktree silently (for shell integration)
	worktreePath := createWorktree(cfg, repoInfo, issueNumber, true)

	// Output just the path for shell function to use
	fmt.Println(worktreePath)
}
