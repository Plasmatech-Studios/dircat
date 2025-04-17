// dircat.go
package dircat

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Entry is one file’s metadata + content
type Entry struct {
	Filename  string `json:"filename"`
	Directory string `json:"directory"`
	Content   string `json:"content"`
}

// Config controls which files to skip
type Config struct {
	// Any paths matching these globs (relative to root) will be skipped
	IgnorePatterns []string
}

// Bundle scans the given root directory (recursively),
// skips hidden files/dirs, skips any paths matching cfg.IgnorePatterns,
// skips binaries (NUL‑byte sniff in first 8KiB),
// and returns a slice of Entry for every text file found.
func Bundle(root string, cfg Config) ([]Entry, error) {
	var results []Entry

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // unreadable → skip
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		// Always recurse into the root itself
		if rel != "." {
			// 1) hidden files/dirs
			if strings.HasPrefix(filepath.Base(rel), ".") {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			// 2) user‑defined ignore globs
			for _, pat := range cfg.IgnorePatterns {
				if matched, _ := filepath.Match(pat, rel); matched {
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}
		}

		// only files from here:
		if d.IsDir() {
			return nil
		}

		// open & sniff for binary
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		buf := make([]byte, 8*1024)
		n, _ := f.Read(buf)
		if bytes.IndexByte(buf[:n], 0) != -1 {
			return nil // binary → skip
		}
		// back to start
		f.Seek(0, io.SeekStart)

		data, err := io.ReadAll(f)
		if err != nil {
			return nil
		}

		results = append(results, Entry{
			Filename:  filepath.Base(rel),
			Directory: filepath.Dir(rel),
			Content:   string(data),
		})
		return nil
	})

	return results, err
}
