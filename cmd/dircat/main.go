// main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/plasmatechstudios/dircat"
)

const (
	configName    = ".dircat.json"
	defaultOutput = "directorycontents.json"
)

func main() {
	help := flag.Bool("h", false, "show help")
	flag.Parse()
	if *help {
		fmt.Println("Usage: dircat [path]")
		os.Exit(0)
	}

	// 1) load config
	raw, err := os.ReadFile(configName)
	if err != nil {
		log.Fatalf("cannot load %s: %v (run `dircat init`)", configName, err)
	}
	var diskCfg struct {
		OutputName     string   `json:"outputName"`
		IgnorePatterns []string `json:"ignorePatterns"`
	}
	if err := json.Unmarshal(raw, &diskCfg); err != nil {
		log.Fatalf("invalid %s: %v", configName, err)
	}

	// 2) choose root
	root := "."
	if args := flag.Args(); len(args) > 0 {
		root = args[0]
	}

	// 3) call the reusable Bundle function
	entries, err := dircat.Bundle(root, dircat.Config{
		IgnorePatterns: diskCfg.IgnorePatterns,
	})
	if err != nil {
		log.Fatalf("error bundling directory: %v", err)
	}

	// 4) write JSON output
	outPath := filepath.Join(root, diskCfg.OutputName)
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

	fmt.Printf("✅ Done—processed %d files and wrote %s\n", len(entries), outPath)
}
