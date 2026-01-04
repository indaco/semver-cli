package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Input represents the JSON input from verso
type Input struct {
	Hook            string                 `json:"hook"`
	Version         string                 `json:"version"`
	PreviousVersion string                 `json:"previous_version"`
	BumpType        string                 `json:"bump_type"`
	Prerelease      *string                `json:"prerelease"`
	Metadata        *string                `json:"metadata"`
	ProjectRoot     string                 `json:"project_root"`
	Config          map[string]interface{} `json:"config"`
}

// Output represents the JSON output to verso
type Output struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// PolicyRules contains the policy configuration
type PolicyRules struct {
	NoPrereleaseOnMain        bool `json:"no_prerelease_on_main"`
	RequireCleanWorkdir       bool `json:"require_clean_workdir"`
	MaxPrereleaseIterations   int  `json:"max_prerelease_iterations"`
	RequireEvenMinorForStable bool `json:"require_even_minor_for_stable"`
}

func main() {
	var input Input

	// Read JSON from stdin
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		outputError(fmt.Sprintf("Invalid JSON input: %v", err))
		return
	}

	// Extract policy rules from config
	rules := extractPolicyRules(input.Config)

	// Validate based on the hook
	if err := validatePolicies(input, rules); err != nil {
		outputError(err.Error())
		return
	}

	// Success
	outputSuccess("All version policies passed")
}

// extractPolicyRules extracts policy rules from config with defaults
func extractPolicyRules(config map[string]interface{}) PolicyRules {
	rules := PolicyRules{
		NoPrereleaseOnMain:        false,
		RequireCleanWorkdir:       false,
		MaxPrereleaseIterations:   10,
		RequireEvenMinorForStable: false,
	}

	if config == nil {
		return rules
	}

	if val, ok := config["no_prerelease_on_main"].(bool); ok {
		rules.NoPrereleaseOnMain = val
	}

	if val, ok := config["require_clean_workdir"].(bool); ok {
		rules.RequireCleanWorkdir = val
	}

	if val, ok := config["max_prerelease_iterations"].(float64); ok {
		rules.MaxPrereleaseIterations = int(val)
	}

	if val, ok := config["require_even_minor_for_stable"].(bool); ok {
		rules.RequireEvenMinorForStable = val
	}

	return rules
}

// validatePolicies validates all applicable policies
func validatePolicies(input Input, rules PolicyRules) error {
	if err := validateNoPrereleaseOnMain(input, rules); err != nil {
		return err
	}
	if err := validateCleanWorkdir(input, rules); err != nil {
		return err
	}
	if err := validatePrereleaseIterations(input, rules); err != nil {
		return err
	}
	if err := validateEvenMinorForStable(input, rules); err != nil {
		return err
	}
	return nil
}

// validateNoPrereleaseOnMain checks if prerelease versions are allowed on main/master branch
func validateNoPrereleaseOnMain(input Input, rules PolicyRules) error {
	if !rules.NoPrereleaseOnMain || input.Prerelease == nil || *input.Prerelease == "" {
		return nil
	}

	currentBranch, err := getCurrentGitBranch(input.ProjectRoot)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	if isMainBranch(currentBranch) {
		return fmt.Errorf("policy violation: prerelease versions are not allowed on main/master branch (current branch: %s)", currentBranch)
	}
	return nil
}

// validateCleanWorkdir checks if the git working directory is clean
func validateCleanWorkdir(input Input, rules PolicyRules) error {
	if !rules.RequireCleanWorkdir {
		return nil
	}

	isClean, err := isGitWorkdirClean(input.ProjectRoot)
	if err != nil {
		return fmt.Errorf("failed to check git working directory: %w", err)
	}

	if !isClean {
		return fmt.Errorf("policy violation: working directory must be clean (no uncommitted changes)")
	}
	return nil
}

// validatePrereleaseIterations checks if prerelease iteration is within limits
func validatePrereleaseIterations(input Input, rules PolicyRules) error {
	if rules.MaxPrereleaseIterations <= 0 || input.Prerelease == nil || *input.Prerelease == "" {
		return nil
	}

	iteration := extractPrereleaseIteration(*input.Prerelease)
	if iteration > rules.MaxPrereleaseIterations {
		return fmt.Errorf("policy violation: prerelease iteration %d exceeds maximum allowed (%d)", iteration, rules.MaxPrereleaseIterations)
	}
	return nil
}

// validateEvenMinorForStable checks if stable releases have even minor versions
func validateEvenMinorForStable(input Input, rules PolicyRules) error {
	if !rules.RequireEvenMinorForStable || (input.Prerelease != nil && *input.Prerelease != "") {
		return nil
	}

	minorVersion := extractMinorVersion(input.Version)
	if minorVersion%2 != 0 {
		return fmt.Errorf("policy violation: stable releases must have even minor version (got %s with minor version %d)", input.Version, minorVersion)
	}
	return nil
}

// getCurrentGitBranch gets the current git branch name
func getCurrentGitBranch(projectRoot string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = projectRoot

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// isMainBranch checks if the branch is a main branch
func isMainBranch(branch string) bool {
	mainBranches := []string{"main", "master"}
	for _, mainBranch := range mainBranches {
		if branch == mainBranch {
			return true
		}
	}
	return false
}

// isGitWorkdirClean checks if the git working directory is clean
func isGitWorkdirClean(projectRoot string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = projectRoot

	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return len(strings.TrimSpace(string(output))) == 0, nil
}

// extractPrereleaseIteration extracts the iteration number from prerelease identifier
// e.g., "alpha.1" -> 1, "rc.5" -> 5
func extractPrereleaseIteration(prerelease string) int {
	parts := strings.Split(prerelease, ".")
	for i := len(parts) - 1; i >= 0; i-- {
		if num, err := strconv.Atoi(parts[i]); err == nil {
			return num
		}
	}
	return 0
}

// extractMinorVersion extracts the minor version number from a semver string
// e.g., "1.2.3" -> 2
func extractMinorVersion(version string) int {
	// Remove prerelease and metadata
	parts := strings.Split(version, "-")
	versionCore := parts[0]

	// Split version core
	versionParts := strings.Split(versionCore, ".")
	if len(versionParts) < 2 {
		return 0
	}

	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return 0
	}

	return minor
}

// outputSuccess outputs a success JSON response
func outputSuccess(message string) {
	output := Output{
		Success: true,
		Message: message,
		Data:    make(map[string]interface{}),
	}

	encoder := json.NewEncoder(os.Stdout)
	if err := encoder.Encode(output); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode output: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

// outputError outputs an error JSON response
func outputError(message string) {
	output := Output{
		Success: false,
		Message: message,
		Data:    make(map[string]interface{}),
	}

	encoder := json.NewEncoder(os.Stdout)
	if err := encoder.Encode(output); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode error output: %v\n", err)
		os.Exit(1)
	}

	os.Exit(1)
}
