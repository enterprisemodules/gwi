package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/enterprisemodules/gwi/internal/config"
	"github.com/enterprisemodules/gwi/internal/git"
	"github.com/enterprisemodules/gwi/internal/hooks"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start dev server in tmux session",
	Long:  `Start the development server in a background tmux session.`,
	Run:   runUp,
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop dev server",
	Long:  `Stop the running development server tmux session.`,
	Run:   runDown,
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View server logs (attach to tmux)",
	Long:  `Attach to the tmux session to view server logs. Detach with Ctrl+B D.`,
	Run:   runLogs,
}

func getSessionName() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "gwi"
	}
	return filepath.Base(cwd)
}

func hasTmux() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

func tmuxSessionExists(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	return cmd.Run() == nil
}

func runUp(cmd *cobra.Command, args []string) {
	if !hasTmux() {
		config.Die("tmux is required for 'gwi up'. Install with: brew install tmux")
	}

	cfg := config.Load()
	sessionName := getSessionName()

	// Check if session already exists
	if tmuxSessionExists(sessionName) {
		config.Info("Session '%s' already running", sessionName)
		fmt.Println("  gwi logs    # to view")
		fmt.Println("  gwi down    # to stop")
		return
	}

	// Find the 'up' hook
	cwd, _ := os.Getwd()
	repoInfo, _ := git.GetRepoInfo()

	upScript := hooks.FindHook("up", cwd, cfg, repoInfo)
	if upScript == "" {
		config.Die("No 'up' hook found. Create .gwi/up with your server start command.")
	}

	config.Info("Starting server in tmux session: %s", sessionName)

	tmuxCmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-c", cwd, upScript)
	if err := tmuxCmd.Run(); err != nil {
		config.Die("Failed to start tmux session: %v", err)
	}

	config.Success("Server started")
	fmt.Println("  gwi logs    # to view")
	fmt.Println("  gwi down    # to stop")
}

func runDown(cmd *cobra.Command, args []string) {
	sessionName := getSessionName()

	if !tmuxSessionExists(sessionName) {
		config.Warn("No session '%s' running", sessionName)
		os.Exit(1)
	}

	config.Info("Stopping session: %s", sessionName)

	tmuxCmd := exec.Command("tmux", "kill-session", "-t", sessionName)
	if err := tmuxCmd.Run(); err != nil {
		config.Die("Failed to stop session: %v", err)
	}

	config.Success("Server stopped")
}

func runLogs(cmd *cobra.Command, args []string) {
	sessionName := getSessionName()

	if !tmuxSessionExists(sessionName) {
		config.Die("No session '%s' running. Start with: gwi up", sessionName)
	}

	config.Info("Attaching to session (Ctrl+B D to detach)")

	tmuxCmd := exec.Command("tmux", "attach-session", "-t", sessionName)
	tmuxCmd.Stdin = os.Stdin
	tmuxCmd.Stdout = os.Stdout
	tmuxCmd.Stderr = os.Stderr
	tmuxCmd.Run()
}
