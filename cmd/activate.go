package cmd

import (
	"fmt"
	"os"

	"github.com/enterprisemodules/gwi/internal/config"
	"github.com/enterprisemodules/gwi/internal/git"
	"github.com/enterprisemodules/gwi/internal/hooks"
	"github.com/spf13/cobra"
)

var activateCmd = &cobra.Command{
	Use:   "activate",
	Short: "Run setup hook (install deps)",
	Long:  `Execute the activate hook script to set up the development environment.`,
	Run:   runActivate,
}

func runActivate(cmd *cobra.Command, args []string) {
	cfg := config.Load()

	worktreePath, err := os.Getwd()
	if err != nil {
		config.Die("Failed to get current directory: %v", err)
	}

	repoInfo, _ := git.GetRepoInfo()

	hookScript := hooks.FindHook("activate", worktreePath, cfg, repoInfo)
	if hookScript == "" {
		config.Warn("No activate hook found")
		fmt.Println("Create one of:")
		fmt.Println("  .gwi/activate (in worktree or main repo)")
		fmt.Printf("  %s/<org>/<repo>/activate\n", cfg.HookDir)
		os.Exit(1)
	}

	hooks.RunHook("activate", worktreePath, cfg, repoInfo)
}
