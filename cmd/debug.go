package cmd

import (
	"fmt"
	"os/exec"
	"strconv"

	"github.com/enterprisemodules/gwi/internal/config"
	"github.com/enterprisemodules/gwi/internal/github"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug [issue-number]",
	Short: "Debug GitHub Projects integration",
	Long:  `Test GitHub Projects integration and show detailed information about configuration and API calls.`,
	Args:  cobra.ExactArgs(1),
	Run:   runDebug,
}

func runDebug(cmd *cobra.Command, args []string) {
	cfg := config.Load()

	issueNumber, err := strconv.Atoi(args[0])
	if err != nil {
		config.Die("Invalid issue number: %s", args[0])
	}

	fmt.Println("=== Configuration ===")
	fmt.Printf("Verbose: %v\n", cfg.Verbose)
	fmt.Printf("Projects Enabled: %v\n", cfg.GitHub.ProjectsEnabled)
	fmt.Printf("Status Field Name: %s\n", cfg.GitHub.StatusFieldName)
	fmt.Printf("Todo Value: %s\n", cfg.GitHub.TodoValue)
	fmt.Printf("In Progress Value: %s\n", cfg.GitHub.InProgressValue)
	fmt.Printf("In Review Value: %s\n", cfg.GitHub.InReviewValue)
	fmt.Printf("Done Value: %s\n", cfg.GitHub.DoneValue)
	fmt.Printf("Check Scopes: %v\n", cfg.GitHub.CheckScopes)
	fmt.Println()

	fmt.Println("=== GitHub CLI ===")
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		config.Die("gh CLI not found in PATH")
	}
	fmt.Printf("gh CLI Path: %s\n", ghPath)

	// Check auth status
	authCmd := exec.Command("gh", "auth", "status")
	authOutput, _ := authCmd.CombinedOutput()
	fmt.Printf("Auth Status:\n%s\n", string(authOutput))

	fmt.Println("=== Issue Information ===")
	fmt.Printf("Testing with issue #%d\n\n", issueNumber)

	// Get project items
	fmt.Println("→ Getting project items...")
	items, err := github.GetProjectItemsForIssue(issueNumber)
	if err != nil {
		config.Die("Failed to get project items: %v", err)
	}

	if len(items) == 0 {
		config.Warn("Issue #%d is not in any GitHub Project!", issueNumber)
		fmt.Println("\nTo fix this:")
		fmt.Println("1. Go to your GitHub Project board")
		fmt.Println("2. Add this issue to the project")
		fmt.Println("3. Run this command again")
		return
	}

	fmt.Printf("✓ Found issue in %d project(s)\n\n", len(items))

	for i, item := range items {
		fmt.Printf("=== Project %d ===\n", i+1)
		fmt.Printf("Item ID: %s\n", item.ID)
		fmt.Printf("Project ID: %s\n\n", item.ProjectID)

		// Get field information
		fmt.Printf("→ Getting '%s' field...\n", cfg.GitHub.StatusFieldName)
		field, err := github.GetProjectField(item.ProjectID, cfg.GitHub.StatusFieldName)
		if err != nil {
			config.Warn("Failed to get field: %v", err)
			fmt.Println("\nAvailable fields in this project:")
			listFieldsCmd := exec.Command("gh", "project", "field-list", item.ProjectID, "--format", "json")
			if output, err := listFieldsCmd.Output(); err == nil {
				fmt.Printf("%s\n", string(output))
			}
			continue
		}

		fmt.Printf("✓ Field ID: %s\n", field.ID)
		fmt.Printf("  Field Name: %s\n", field.Name)
		fmt.Printf("  Available Options:\n")
		for _, opt := range field.Options {
			fmt.Printf("    - %s (ID: %s)\n", opt.Name, opt.ID)
		}
		fmt.Println()

		// Check if the configured status values exist
		fmt.Printf("→ Checking status values...\n")
		if _, err := github.GetFieldOptionID(field, cfg.GitHub.TodoValue); err != nil {
			config.Warn("'%s' not found in options!", cfg.GitHub.TodoValue)
		} else {
			fmt.Printf("✓ '%s' exists\n", cfg.GitHub.TodoValue)
		}

		if _, err := github.GetFieldOptionID(field, cfg.GitHub.InProgressValue); err != nil {
			config.Warn("'%s' not found in options!", cfg.GitHub.InProgressValue)
		} else {
			fmt.Printf("✓ '%s' exists\n", cfg.GitHub.InProgressValue)
		}

		if _, err := github.GetFieldOptionID(field, cfg.GitHub.InReviewValue); err != nil {
			config.Warn("'%s' not found in options!", cfg.GitHub.InReviewValue)
		} else {
			fmt.Printf("✓ '%s' exists\n", cfg.GitHub.InReviewValue)
		}

		if _, err := github.GetFieldOptionID(field, cfg.GitHub.DoneValue); err != nil {
			config.Warn("'%s' not found in options!", cfg.GitHub.DoneValue)
		} else {
			fmt.Printf("✓ '%s' exists\n", cfg.GitHub.DoneValue)
		}
		fmt.Println()
	}

	fmt.Println("=== Test Update ===")
	fmt.Printf("Would you like to test updating issue #%d to '%s'? [y/N]: ", issueNumber, cfg.GitHub.InProgressValue)
	var response string
	fmt.Scanln(&response)
	if response == "y" || response == "Y" {
		fmt.Println("\n→ Updating issue status...")
		if err := github.UpdateIssueStatus(issueNumber, cfg.GitHub.InProgressValue, cfg); err != nil {
			config.Die("Update failed: %v", err)
		}
		config.Success("Issue #%d updated to '%s'", issueNumber, cfg.GitHub.InProgressValue)
	} else {
		fmt.Println("Skipped update test.")
	}
}
