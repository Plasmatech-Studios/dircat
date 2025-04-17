// main.go
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	toolName      = "dircat"
	configName    = ".dircat.json"
	defaultOutput = "directorycontents.json"
)

type Config struct {
	OutputName     string   `json:"outputName"`
	IgnorePatterns []string `json:"ignorePatterns"`
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage:
  %s init           Initialize configuration in current directory.
  %s [path]         Bundle files under [path] (default: .), per config.

Flags:
`, toolName, toolName)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	help := flag.Bool("h", false, "show help")
	flag.Parse()
	args := flag.Args()

	if *help {
		usage()
		return
	}
	if len(args) > 0 && args[0] == "init" {
		initConfig()
		return
	}
	runBundle(args)
}

func initConfig() {
	if _, err := os.Stat(configName); err == nil {
		fmt.Printf("⚠️  %s already exists. Delete it first to re‑init.\n", configName)
		return
	}

	r := bufio.NewReader(os.Stdin)
	fmt.Printf("Welcome to %s setup!\n\n", toolName)

	// 1) Output file name
	fmt.Printf("1) Output file name (default: %s): ", defaultOutput)
	outRaw, _ := r.ReadString('\n')
	out := strings.TrimSpace(outRaw)
	if out == "" {
		out = defaultOutput
	}

	// 2) Ignore globs
	//    by default we only ignore the config and output files;
	//    hidden files (.*) are skipped automatically in code
	defIg := []string{configName, out}
	fmt.Printf("2) Comma‑separated ignore globs (default: %s): ", strings.Join(defIg, ", "))
	igRaw, _ := r.ReadString('\n')
	ig := strings.TrimSpace(igRaw)

	var ignores []string
	if ig == "" {
		ignores = defIg
	} else {
		for _, pat := range strings.Split(ig, ",") {
			ignores = append(ignores, strings.TrimSpace(pat))
		}
	}

	cfg := Config{OutputName: out, IgnorePatterns: ignores}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Fatalf("❌ failed to build config: %v", err)
	}
	if err := os.WriteFile(configName, data, 0o644); err != nil {
		log.Fatalf("❌ unable to write %s: %v", configName, err)
	}

	fmt.Printf("\n✅ Configuration written to %s\n", configName)
	fmt.Printf("Run `%s [path]` (default path is current dir) to bundle your files.\n", toolName)
}

func runBundle(args []string) {
	// load config
	raw, err := os.ReadFile(configName)
	if err != nil {
		fmt.Printf("❌ Cannot load %s: %v\nPlease run `%s init` first.\n", configName, err, toolName)
		os.Exit(1)
	}
	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		fmt.Printf("❌ Invalid %s: %v\nPlease delete it and run `%s init`.\n", configName, err, toolName)
		os.Exit(1)
	}

	// choose root
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	outPath := filepath.Join(root, cfg.OutputName)
	outFile, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("❌ cannot create %s: %v", outPath, err)
	}
	defer outFile.Close()

	// spinner setup
	spinner := []rune{'|', '/', '-', '\\'}
	si := 0
	processedCount := 0

	// begin JSON array
	outFile.WriteString("[\n")
	first := true

	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		// always recurse into root
		if rel != "." {
			// skip hidden basenames
			if strings.HasPrefix(filepath.Base(rel), ".") {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			// skip config & output
			if rel == configName || rel == cfg.OutputName {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			// skip any user‑defined globs
			for _, pat := range cfg.IgnorePatterns {
				if matched, _ := filepath.Match(pat, rel); matched {
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}
		}

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
			return nil // binary
		}
		f.Seek(0, io.SeekStart)

		content, err := io.ReadAll(f)
		if err != nil {
			return nil
		}

		// update spinner line
		fmt.Printf("\r%c processing file %s", spinner[si], rel)
		si = (si + 1) % len(spinner)

		// build entry
		entry := struct {
			Filename  string `json:"filename"`
			Directory string `json:"directory"`
			Content   string `json:"content"`
		}{
			Filename:  filepath.Base(rel),
			Directory: filepath.Dir(rel),
			Content:   string(content),
		}

		if !first {
			outFile.WriteString(",\n")
		}
		first = false

		enc := json.NewEncoder(outFile)
		enc.SetIndent("  ", "  ")
		enc.Encode(entry)

		processedCount++
		return nil
	})

	// close JSON
	outFile.WriteString("\n]\n")

	// finish spinner line & print summary
	fmt.Printf("\r✅ Done—processed %d files and wrote JSON bundle to %s\n", processedCount, outPath)
}
