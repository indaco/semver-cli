package auditlog

import (
	"os"
)

// DefaultFileOps implements FileOperations using standard library.
type DefaultFileOps struct{}

// ReadFile reads a file from disk.
func (f *DefaultFileOps) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile writes data to a file.
func (f *DefaultFileOps) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// FileExists checks if a file exists.
func (f *DefaultFileOps) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
