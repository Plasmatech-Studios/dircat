package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Plasmatech-Studios/dircat/pkg/dircat"
)

const version = "1.1.0"

func main() {
	// parse flags
	help := flag.Bool("h", false, "show help")
	showVersion := flag.Bool("version", false, "show version")
	flag.Parse()

	if *help {
		fmt.Println("Usage: dircat [path]")
		fmt.Println("Options:")
		fmt.Println("  -h        show help")
		fmt.Println("  --version show version")
		os.Exit(0)
	}

	if *showVersion {
		fmt.Printf("dircat version %s\n", version)
		os.Exit(0)
	}

	// attempt to load config file
	raw, err := os.ReadFile(dircat.DefaultConfigFileName)
	var b dircat.Bundler
	var outputName string

	if err != nil {
		// no config found → use defaults
		b = dircat.NewDefaultBundler()
		outputName = dircat.DefaultOutputName
	} else {
		// parse custom config
		var diskCfg struct {
			OutputName     string   `json:"outputName"`
			IgnorePatterns []string `json:"ignorePatterns"`
		}
		if err := json.Unmarshal(raw, &diskCfg); err != nil {
			log.Printf("invalid %s: %v; using default config", dircat.DefaultConfigFileName, err)
			b = dircat.NewDefaultBundler()
			outputName = dircat.DefaultOutputName
		} else {
			b = dircat.NewBundler(dircat.Config{IgnorePatterns: diskCfg.IgnorePatterns})
			outputName = diskCfg.OutputName
		}
	}

	// determine root directory
	root := "."
	if args := flag.Args(); len(args) > 0 {
		root = args[0]
	}

	// perform bundling
	entries, err := b.Bundle(root)
	if err != nil {
		log.Fatalf("error bundling directory: %v", err)
	}

	// write JSON output
	outPath := filepath.Join(root, outputName)
	f, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("cannot create %s: %v", outPath, err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(entries); err != nil {
		log.Fatalf("failed to write JSON: %v", err)
	}

	// summary
	fmt.Printf("✅ Done—processed %d files and wrote %s\n", len(entries), outPath)
}
