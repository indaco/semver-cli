package extensionmgr

import (
	"io"
	"os"
	"path/filepath"
)

var (
	// walkFn is used to walk the file tree; override in tests if needed.
	walkFn = filepath.Walk

	// relFn computes relative file paths; override in tests to simulate failure.
	relFn = filepath.Rel

	// openSrcFile is a hook for opening source files (for mocking in tests).
	openSrcFile = os.Open

	// openDstFile is a hook for creating destination files (for mocking in tests).
	openDstFile = os.OpenFile

	// copyFn performs the actual file copy; override in tests to simulate errors.
	copyFn = io.Copy

	// skipNames defines a set of directory or file names excluded during directory copying.
	skipNames = map[string]struct{}{
		".git":         {},
		".DS_Store":    {},
		"node_modules": {},
	}

	copyDirFn = copyDir
)

// copyDir recursively copies all files and subdirectories from src to dst.
// It preserves permissions and creates necessary subfolders automatically.
func copyDir(src, dst string) error {
	return walkFn(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		skipFile, skipDir := shouldSkipEntry(info)
		if skipDir {
			return filepath.SkipDir
		}
		if skipFile {
			return nil
		}

		rel, err := relFn(src, path)
		if err != nil {
			return err
		}

		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		return copyFile(path, target, info.Mode())
	})
}

// copyFile copies a single file from src to dst, preserving the given permissions.
// Used internally by CopyDir.
func copyFile(src, dst string, perm os.FileMode) error {
	in, err := openSrcFile(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := openDstFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := copyFn(out, in); err != nil {
		return err
	}

	return nil
}

// shouldSkipEntry determines whether a file should be skipped or a directory subtree should be skipped.
func shouldSkipEntry(info os.FileInfo) (skipFile bool, skipDir bool) {
	_, skip := skipNames[info.Name()]
	if !skip {
		return false, false
	}
	if info.IsDir() {
		return false, true // skip entire directory
	}
	return true, false // skip just the file
}
