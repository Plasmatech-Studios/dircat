package dircat

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DefaultConfigFileName is the default name for the config file
	DefaultConfigFileName = ".dircat.json"
	// DefaultOutputName is the default name for the output bundle
	DefaultOutputName = "directorycontents.json"
)

// Entry represents a bundled file's metadata and content.
type Entry struct {
	Filename  string `json:"filename"`
	Directory string `json:"directory"`
	Content   string `json:"content"`
}

// Config holds ignore patterns for the bundler.
type Config struct {
	IgnorePatterns []string
}

// Bundler bundles files from a directory according to its Config.
type Bundler interface {
	Bundle(root string) ([]Entry, error)
}

// bundler is the default implementation of Bundler.
type bundler struct {
	cfg Config
}

// NewBundler creates a new Bundler with the given Config.
func NewBundler(cfg Config) Bundler {
	return &bundler{cfg: cfg}
}

// NewDefaultBundler creates a Bundler with default settings.
// It skips the default config file and default output name.
func NewDefaultBundler() Bundler {
	return &bundler{cfg: Config{
		IgnorePatterns: []string{DefaultConfigFileName, DefaultOutputName},
	}}
}

// Bundle scans root recursively, skipping hidden, binary, and ignored files,
// and returns a slice of Entry for each readable text file.
func (b *bundler) Bundle(root string) ([]Entry, error) {
	var results []Entry

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		// Always recurse into the root itself
		if rel != "." {
			// 1) skip hidden files/dirs
			if strings.HasPrefix(filepath.Base(rel), ".") {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			// 2) skip user-defined ignore patterns
			for _, pat := range b.cfg.IgnorePatterns {
				if matched, _ := filepath.Match(pat, rel); matched {
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}
		}

		// skip directories
		if d.IsDir() {
			return nil
		}

		// open file
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		// sniff for binary (NUL byte)
		buf := make([]byte, 8*1024)
		n, _ := f.Read(buf)
		if bytes.IndexByte(buf[:n], 0) != -1 {
			return nil
		}

		// rewind
		f.Seek(0, io.SeekStart)

		// read all
		data, err := io.ReadAll(f)
		if err != nil {
			return nil
		}

		// append entry
		results = append(results, Entry{
			Filename:  filepath.Base(rel),
			Directory: filepath.Dir(rel),
			Content:   string(data),
		})
		return nil
	})

	return results, err
}
