// Package fsutil provides path-safe file reads scoped to a directory root
// (via os.OpenRoot) to satisfy path traversal checks from static analysis.
package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadFileUnderRoot reads rel inside rootDir using os.OpenRoot. rel must be a
// single path segment (no ".." or separators) so it cannot escape rootDir.
func ReadFileUnderRoot(rootDir, rel string) ([]byte, error) {
	rel = filepath.ToSlash(rel)
	if rel == "" || rel == "." || strings.ContainsAny(rel, "/\\") {
		return nil, fmt.Errorf("invalid relative file name %q", rel)
	}
	base := filepath.Base(rel)
	if base != rel || base == ".." {
		return nil, fmt.Errorf("invalid relative file name %q", rel)
	}

	root, err := os.OpenRoot(filepath.Clean(rootDir))
	if err != nil {
		return nil, fmt.Errorf("open root %q: %w", rootDir, err)
	}
	defer root.Close()

	data, err := root.ReadFile(base)
	if err != nil {
		return nil, fmt.Errorf("read %q under %q: %w", base, rootDir, err)
	}
	return data, nil
}
