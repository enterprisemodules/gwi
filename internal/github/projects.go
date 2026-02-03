package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/enterprisemodules/gwi/internal/config"
)

// ProjectItem represents an issue's association with a GitHub Project
type ProjectItem struct {
	ID        string
	ProjectID string
}

// ProjectField represents a field in a GitHub Project
type ProjectField struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	Options []FieldOption `json:"options"`
}

// FieldOption represents an option for a single-select field
type FieldOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Cache for project field IDs to minimize API calls
var (
	fieldCache = make(map[string]*ProjectField)
	cacheMutex sync.RWMutex
)

// ParseIssueFromBranch extracts issue number from branch name
// Supports formats: 42-feature-name, feature/42-name, etc.
func ParseIssueFromBranch(branchName string) (int, bool) {
	// Match one or more digits, typically at the start or after a slash
	re := regexp.MustCompile(`(?:^|/)(\d+)[-_]`)
	matches := re.FindStringSubmatch(branchName)
	if matches != nil && len(matches) > 1 {
		if num, err := strconv.Atoi(matches[1]); err == nil {
			return num, true
		}
	}
	return 0, false
}

// CheckProjectScopes verifies required GitHub CLI scopes and prompts to refresh if missing
func CheckProjectScopes() error {
	cmd := exec.Command("gh", "auth", "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("GitHub CLI not authenticated. Run: gh auth login")
	}

	outputStr := string(output)
	hasReadProject := strings.Contains(outputStr, "read:project")
	hasWriteProject := strings.Contains(outputStr, "write:project")

	if !hasReadProject || !hasWriteProject {
		config.Warn("Missing required GitHub scopes for Projects integration")
		config.Info("Attempting to refresh authentication with required scopes...")

		refreshCmd := exec.Command("gh", "auth", "refresh", "-s", "read:project,write:project")
		refreshCmd.Stdin = nil // Will prompt user interactively
		refreshCmd.Stdout = nil
		refreshCmd.Stderr = nil

		if err := refreshCmd.Run(); err != nil {
			return fmt.Errorf("failed to refresh auth. Please run manually: gh auth refresh -s read:project,write:project")
		}

		config.Success("Authentication refreshed with project scopes")
	}

	return nil
}

// GetProjectItemsForIssue finds all project items for an issue
func GetProjectItemsForIssue(issueNumber int) ([]ProjectItem, error) {
	cmd := exec.Command("gh", "issue", "view", strconv.Itoa(issueNumber),
		"--json", "projectItems",
		"--jq", ".projectItems[] | {id: .id, projectId: .project.id}")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get project items for issue #%d", issueNumber)
	}

	// Parse NDJSON (newline-delimited JSON)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var items []ProjectItem

	for _, line := range lines {
		if line == "" {
			continue
		}
		var item struct {
			ID        string `json:"id"`
			ProjectID string `json:"projectId"`
		}
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			continue
		}
		items = append(items, ProjectItem{
			ID:        item.ID,
			ProjectID: item.ProjectID,
		})
	}

	return items, nil
}

// GetProjectField retrieves field information by name with caching
func GetProjectField(projectID, fieldName string) (*ProjectField, error) {
	cacheKey := projectID + ":" + fieldName

	// Check cache first
	cacheMutex.RLock()
	if cached, ok := fieldCache[cacheKey]; ok {
		cacheMutex.RUnlock()
		return cached, nil
	}
	cacheMutex.RUnlock()

	// Fetch from API
	cmd := exec.Command("gh", "project", "field-list", projectID,
		"--owner", "@me",
		"--format", "json")

	output, err := cmd.Output()
	if err != nil {
		// Try without --owner flag (might be org project)
		cmd = exec.Command("gh", "project", "field-list", projectID,
			"--format", "json")
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to list fields for project %s", projectID)
		}
	}

	var fields struct {
		Fields []ProjectField `json:"fields"`
	}
	if err := json.Unmarshal(output, &fields); err != nil {
		return nil, fmt.Errorf("failed to parse project fields: %w", err)
	}

	// Find the field by name
	for _, field := range fields.Fields {
		if field.Name == fieldName {
			// Cache it
			cacheMutex.Lock()
			fieldCache[cacheKey] = &field
			cacheMutex.Unlock()
			return &field, nil
		}
	}

	return nil, fmt.Errorf("field '%s' not found in project", fieldName)
}

// GetFieldOptionID finds the option ID for a status value
func GetFieldOptionID(field *ProjectField, optionName string) (string, error) {
	for _, option := range field.Options {
		if option.Name == optionName {
			return option.ID, nil
		}
	}
	return "", fmt.Errorf("option '%s' not found in field '%s'", optionName, field.Name)
}

// UpdateProjectItemStatus updates the status field for a project item
func UpdateProjectItemStatus(item ProjectItem, fieldID, optionID string, cfg *config.Config) error {
	cmd := exec.Command("gh", "project", "item-edit",
		"--id", item.ID,
		"--project-id", item.ProjectID,
		"--field-id", fieldID,
		"--single-select-option-id", optionID)

	if cfg.Verbose {
		config.Info("Updating project item %s in project %s", item.ID, item.ProjectID)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update item: %s", strings.TrimSpace(string(output)))
	}

	return nil
}

// UpdateIssueStatus is the main function to update issue status in all projects
func UpdateIssueStatus(issueNumber int, statusValue string, cfg *config.Config) error {
	// Check if gh CLI is available
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI not found in PATH")
	}

	// Check scopes if enabled
	if cfg.GitHub.CheckScopes {
		if err := CheckProjectScopes(); err != nil {
			return err
		}
	}

	// Get all project items for this issue
	items, err := GetProjectItemsForIssue(issueNumber)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		if cfg.Verbose {
			config.Info("Issue #%d is not in any GitHub Project", issueNumber)
		}
		return nil
	}

	if cfg.Verbose {
		config.Info("Found issue #%d in %d project(s)", issueNumber, len(items))
	}

	// Update each project
	var lastErr error
	successCount := 0

	for _, item := range items {
		// Get the Status field for this project
		field, err := GetProjectField(item.ProjectID, cfg.GitHub.StatusFieldName)
		if err != nil {
			if cfg.Verbose {
				config.Warn("Could not get '%s' field: %v", cfg.GitHub.StatusFieldName, err)
			}
			lastErr = err
			continue
		}

		// Get the option ID for the desired status
		optionID, err := GetFieldOptionID(field, statusValue)
		if err != nil {
			if cfg.Verbose {
				config.Warn("Could not find option '%s': %v", statusValue, err)
			}
			lastErr = err
			continue
		}

		// Update the item
		if err := UpdateProjectItemStatus(item, field.ID, optionID, cfg); err != nil {
			if cfg.Verbose {
				config.Warn("Failed to update project item: %v", err)
			}
			lastErr = err
			continue
		}

		successCount++
	}

	if successCount == 0 && lastErr != nil {
		return lastErr
	}

	if successCount > 0 {
		if cfg.Verbose {
			config.Success("Updated %d project(s) to '%s'", successCount, statusValue)
		}
	}

	return nil
}
