package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gwi",
	Short: "Git Worktree Issue CLI",
	Long:  `gwi integrates GitHub issues with git worktrees for streamlined development.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add all subcommands
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(prCmd)
	rootCmd.AddCommand(mergeCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(cdCmd)
	rootCmd.AddCommand(internalCdCmd)
	rootCmd.AddCommand(mainCmd)
	rootCmd.AddCommand(internalMainCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(internalListCmd)
	rootCmd.AddCommand(internalStartCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(activateCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(completionCmd)
}
