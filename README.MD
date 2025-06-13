# dircat

A configurable CLI tool and Go library that recursively scans a directory, skips hidden and binary files, and bundles all text file contents into a single JSON output, with customizable ignore patterns.

---

## 📦 Features

- **Configurable**: Define ignore patterns and output filename in a `.dircat.json` file.
- **Defaults**: If no config is present or it’s invalid, dircat falls back to built-in defaults.
- **CLI & Library**: Use the `dircat` command-line tool or import the package into your Go project.
- **Skip Hidden/Binary**: Automatically ignores files/directories starting with `.` and binary files (NUL-byte sniff).

---

## ⚙️ Installation (CLI)

```bash
# Clone the repository
git clone https://github.com/Plasmatech-Studios/dircat.git
cd dircat

# Build the CLI binary
go build -o dircat ./cmd/dircat

# (Optional) Move it into your PATH (overwriting any existing dircat)
sudo mv -f dircat /usr/local/bin/dircat
```

---

## 🚀 Usage (CLI)

1. **Initialize** (optional):
   ```bash
   dircat init
   ```
   Creates a `.dircat.json` with defaults in the current directory.

2. **Run**:
   ```bash
   # Scan current directory and produce JSON in ./directorycontents.json
dircat

# Or scan a specific folder:
dircat /path/to/project
   ```

3. **Output**:
   - Reads files under the specified root (default `.`).
   - Writes JSON to the root directory, named according to your config (default `directorycontents.json`).

---

## ⚙️ Configuration

Copy the example to create a config:
```bash
cp .dircat.json.example .dircat.json
```

Edit **`.dircat.json`**:

```json
{
  "outputName": "directorycontents.json",
  "ignorePatterns": [
    ".dircat.json",
    "directorycontents.json",
    "node_modules/*"
  ]
}
```

- **outputName**: Filename for the JSON bundle, placed in the root directory scanned.
- **ignorePatterns**: Glob patterns (relative to the root) to exclude from scanning.

---

## 📦 Usage (Go Library)

Import and call from your code:

```go
import (
  "fmt"
  "github.com/Plasmatech-Studios/dircat/pkg/dircat"
)

func main() {
  // 1) Choose bundler: default or custom
  b := dircat.NewDefaultBundler()
  // Or load custom config...
  // config := dircat.Config{IgnorePatterns: []string{"vendor/*"}}
  // b := dircat.NewBundler(config)

  // 2) Bundle files under "./myproject"
  entries, err := b.Bundle("./myproject")
  if err != nil {
    panic(err)
  }

  // 3) Work with results
  for _, e := range entries {
    fmt.Printf("%s/%s: %d bytes\n", e.Directory, e.Filename, len(e.Content))
  }
}
```

---

## 🗂️ Repository Layout

```
dircat/                        # module root
├── cmd/dircat/                # CLI entrypoint
│   └── main.go                # loads config, calls package, writes JSON
├── pkg/dircat/                # importable library
│   └── dircat.go              # Bundle logic, Config, Entry definitions
├── .dircat.json.example       # sample config
├── README.md                  # this file
├── LICENSE                    # MIT License
└── go.mod                     # module declaration
```

---

## 📝 License

Released under the [MIT License](LICENSE).

