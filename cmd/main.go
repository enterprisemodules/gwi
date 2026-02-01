package cmd

import (
	"fmt"

	"github.com/enterprisemodules/gwi/internal/config"
	"github.com/enterprisemodules/gwi/internal/git"
	"github.com/spf13/cobra"
)

var mainCmd = &cobra.Command{
	Use:   "main",
	Short: "Navigate back to main repository",
	Long:  `Navigate to the main repository (not a worktree).`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use 'gwi main' with shell integration. Add to your shell config:")
		fmt.Println("  eval \"$(gwi init zsh)\"")
	},
}

var internalMainCmd = &cobra.Command{
	Use:    "_main",
	Hidden: true,
	Run:    runInternalMain,
}

func runInternalMain(cmd *cobra.Command, args []string) {
	mainPath, err := git.GetMainWorktreePath()
	if err != nil {
		config.Die("Could not find main repository")
	}
	fmt.Println(mainPath)
}
