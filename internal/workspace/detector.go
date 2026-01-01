// Package workspace provides types and operations for managing multiple modules
// in a monorepo or multi-module context.
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/core"
	"github.com/indaco/semver-cli/internal/semver"
)

// DetectionMode indicates how the CLI should operate based on discovered .version files.
type DetectionMode int

const (
	// SingleModule indicates a single .version file was found (in cwd or exactly one in subdirs).
	SingleModule DetectionMode = iota

	// MultiModule indicates multiple .version files were found in subdirectories.
	MultiModule

	// NoModules indicates no .version files were found.
	NoModules
)

// String returns a human-readable representation of the detection mode.
func (m DetectionMode) String() string {
	switch m {
	case SingleModule:
		return "SingleModule"
	case MultiModule:
		return "MultiModule"
	case NoModules:
		return "NoModules"
	default:
		return fmt.Sprintf("Unknown(%d)", m)
	}
}

// Context represents the execution context determined by the detector.
// It encapsulates the mode and any discovered modules.
type Context struct {
	// Mode indicates the execution mode based on discovery.
	Mode DetectionMode

	// Path is the .version file path for SingleModule mode.
	Path string

	// Modules contains discovered modules for MultiModule mode.
	Modules []*Module
}

// Detector discovers .version files and determines execution context.
type Detector struct {
	fs  core.FileSystem
	cfg *config.Config
}

// NewDetector creates a new Detector with the given filesystem and configuration.
func NewDetector(fs core.FileSystem, cfg *config.Config) *Detector {
	return &Detector{
		fs:  fs,
		cfg: cfg,
	}
}

// DetectContext determines the execution context from the given directory.
// It follows this algorithm:
//  1. Check if .version exists in cwd -> SingleModule mode
//  2. If not, scan subdirectories for .version files
//  3. If 0 found -> NoModules
//  4. If 1 found -> SingleModule (use that file)
//  5. If 2+ found -> MultiModule
//
// If the configuration has explicit modules defined, they are used instead of discovery.
func (d *Detector) DetectContext(cwd string) (*Context, error) {
	// First, check for .version in current directory
	versionPath := filepath.Join(cwd, ".version")
	if _, err := d.fs.Stat(versionPath); err == nil {
		return &Context{
			Mode: SingleModule,
			Path: versionPath,
		}, nil
	}

	// If explicit modules are configured, use them
	if d.cfg.HasExplicitModules() {
		modules, err := d.loadExplicitModules(cwd)
		if err != nil {
			return nil, fmt.Errorf("failed to load explicit modules: %w", err)
		}

		if len(modules) == 0 {
			return &Context{Mode: NoModules}, nil
		} else if len(modules) == 1 {
			return &Context{
				Mode: SingleModule,
				Path: modules[0].Path,
			}, nil
		}

		return &Context{
			Mode:    MultiModule,
			Modules: modules,
		}, nil
	}

	// Otherwise, discover modules in subdirectories
	modules, err := d.DiscoverModules(cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to discover modules: %w", err)
	}

	if len(modules) == 0 {
		return &Context{Mode: NoModules}, nil
	} else if len(modules) == 1 {
		return &Context{
			Mode: SingleModule,
			Path: modules[0].Path,
		}, nil
	}

	return &Context{
		Mode:    MultiModule,
		Modules: modules,
	}, nil
}

// DiscoverModules finds all .version files in subdirectories according to configuration.
// It respects exclude patterns, max depth, and recursive settings.
func (d *Detector) DiscoverModules(root string) ([]*Module, error) {
	discovery := d.cfg.GetDiscoveryConfig()

	// Check if discovery is disabled
	if discovery.Enabled != nil && !*discovery.Enabled {
		return nil, nil
	}

	// Load ignore patterns
	ignorePatterns, err := d.loadIgnorePatterns(root)
	if err != nil {
		return nil, fmt.Errorf("failed to load ignore patterns: %w", err)
	}

	// Merge config excludes with .semverignore patterns
	excludes := d.cfg.GetExcludePatterns()
	allPatterns := make([]string, 0, len(excludes)+len(ignorePatterns))
	allPatterns = append(allPatterns, excludes...)
	allPatterns = append(allPatterns, ignorePatterns...)

	// Create a pattern matcher
	matcher := newPatternMatcher(allPatterns)

	// Scan directories recursively
	return d.scanDirectory(root, 0, root, discovery, matcher)
}

// scanDirectory scans a single directory for .version files and subdirectories.
func (d *Detector) scanDirectory(dir string, depth int, root string, discovery *config.DiscoveryConfig, matcher *patternMatcher) ([]*Module, error) {
	// Check max depth
	maxDepth := 10
	if discovery.MaxDepth != nil {
		maxDepth = *discovery.MaxDepth
	}
	if depth > maxDepth {
		return nil, nil
	}

	// Read directory entries
	entries, err := d.fs.ReadDir(dir)
	if err != nil {
		// Skip directories we can't read (permission issues, etc.)
		return nil, nil
	}

	var modules []*Module

	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(dir, name)

		// Skip if matches exclude patterns
		relPath, err := filepath.Rel(root, fullPath)
		if err != nil {
			relPath = name
		}
		if matcher.matches(name) || matcher.matches(relPath) {
			continue
		}

		if entry.IsDir() {
			// Check if recursive scanning is enabled
			recursive := true
			if discovery.Recursive != nil {
				recursive = *discovery.Recursive
			}

			// Scan subdirectory recursively if enabled
			if recursive {
				subModules, err := d.scanDirectory(fullPath, depth+1, root, discovery, matcher)
				if err != nil {
					return nil, err
				}
				modules = append(modules, subModules...)
			}
		} else if name == ".version" {
			// Found a .version file
			module, err := d.createModule(fullPath, root)
			if err != nil {
				return nil, fmt.Errorf("failed to create module for %s: %w", fullPath, err)
			}
			modules = append(modules, module)
		}
	}

	return modules, nil
}

// createModule creates a Module from a .version file path.
func (d *Detector) createModule(versionPath string, root string) (*Module, error) {
	// Get relative path
	relPath, err := filepath.Rel(root, versionPath)
	if err != nil {
		relPath = versionPath
	}

	// Module name is the directory name
	dir := filepath.Dir(versionPath)
	name := filepath.Base(dir)

	// Load current version
	vm := semver.NewVersionManager(d.fs, nil)
	version, err := vm.Read(versionPath)
	if err != nil {
		// If we can't read the version, use empty string
		return &Module{
			Name:           name,
			Path:           versionPath,
			RelPath:        relPath,
			CurrentVersion: "",
			Dir:            dir,
		}, nil
	}

	return &Module{
		Name:           name,
		Path:           versionPath,
		RelPath:        relPath,
		CurrentVersion: version.String(),
		Dir:            dir,
	}, nil
}

// loadExplicitModules loads modules from explicit configuration.
func (d *Detector) loadExplicitModules(root string) ([]*Module, error) {
	if d.cfg.Workspace == nil || len(d.cfg.Workspace.Modules) == 0 {
		return nil, nil
	}

	var modules []*Module
	for _, moduleConfig := range d.cfg.Workspace.Modules {
		// Skip disabled modules
		if !moduleConfig.IsEnabled() {
			continue
		}

		// Resolve path relative to root
		versionPath := moduleConfig.Path
		if !filepath.IsAbs(versionPath) {
			versionPath = filepath.Join(root, versionPath)
		}

		// Check if file exists
		if _, err := d.fs.Stat(versionPath); err != nil {
			return nil, fmt.Errorf("module %s: version file not found at %s", moduleConfig.Name, versionPath)
		}

		// Create module
		module, err := d.createModuleFromConfig(moduleConfig, versionPath, root)
		if err != nil {
			return nil, fmt.Errorf("failed to create module %s: %w", moduleConfig.Name, err)
		}

		modules = append(modules, module)
	}

	return modules, nil
}

// createModuleFromConfig creates a Module from explicit configuration.
func (d *Detector) createModuleFromConfig(moduleConfig config.ModuleConfig, versionPath string, root string) (*Module, error) {
	// Get relative path
	relPath, err := filepath.Rel(root, versionPath)
	if err != nil {
		relPath = versionPath
	}

	dir := filepath.Dir(versionPath)

	// Load current version
	vm := semver.NewVersionManager(d.fs, nil)
	version, err := vm.Read(versionPath)
	if err != nil {
		// If we can't read the version, use empty string
		return &Module{
			Name:           moduleConfig.Name,
			Path:           versionPath,
			RelPath:        relPath,
			CurrentVersion: "",
			Dir:            dir,
		}, nil
	}

	return &Module{
		Name:           moduleConfig.Name,
		Path:           versionPath,
		RelPath:        relPath,
		CurrentVersion: version.String(),
		Dir:            dir,
	}, nil
}

// loadIgnorePatterns loads patterns from .semverignore file if it exists.
func (d *Detector) loadIgnorePatterns(root string) ([]string, error) {
	ignorePath := filepath.Join(root, ".semverignore")
	data, err := d.fs.ReadFile(ignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read .semverignore: %w", err)
	}

	return parseIgnoreFile(string(data)), nil
}

// parseIgnoreFile parses a .semverignore file and returns the patterns.
func parseIgnoreFile(content string) []string {
	var patterns []string
	lines := strings.SplitSeq(content, "\n")

	for line := range lines {
		// Trim whitespace
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		patterns = append(patterns, line)
	}

	return patterns
}

// patternMatcher provides efficient pattern matching for exclude patterns.
type patternMatcher struct {
	ignoreFile *IgnoreFile
}

// newPatternMatcher creates a new pattern matcher with the given patterns.
func newPatternMatcher(patterns []string) *patternMatcher {
	// Convert patterns to .semverignore format
	content := strings.Join(patterns, "\n")
	return &patternMatcher{
		ignoreFile: NewIgnoreFile(content),
	}
}

// matches checks if the given path matches any of the patterns.
func (pm *patternMatcher) matches(path string) bool {
	return pm.ignoreFile.Matches(path)
}
