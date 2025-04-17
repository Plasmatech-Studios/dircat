# dircat

A configurable CLI tool that recursively scans a directory, skips hidden and binary files, and bundles all text file contents into a single JSON output—with customizable ignore patterns.

## Features

- **Init**  
  `dircat init`  
  Creates a `.dircat.json` in the current folder with sensible defaults (output filename, ignore globs).

- **Bundle**  
  `dircat [path]`  
  Reads your config, scans `path` (defaults to `.`), skips:
  - Hidden files/directories (names beginning with `.`)
  - Any globs you listed
  - Binary files (detected by NUL-byte sniff)
  
  Bundles each text file into a JSON array entry:
  ```json
  {
    "filename": "main.go",
    "directory": ".",
    "content": "package main …"
  }
  ```

- **Live Spinner**  
  Shows `| / – \` cycling as it processes each file, all on one line.

- **Summary**  
  Prints `✅ Done—processed X files and wrote JSON bundle to <output>` at completion.

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/dircat.git
cd dircat

# Build the binary
go build -o dircat main.go

# (Optional) Move it to your PATH
mv dircat /usr/local/bin/
```

## Usage

```bash
# Initialize config in current directory
dircat init

# Bundle files under current directory
dircat

# Bundle files in another folder
dircat /path/to/project
```

## Configuration

Edit `.dircat.json` to tweak:
```json
{
  "outputName": "directorycontents.json",
  "ignorePatterns": [
    ".dircat.json",
    "directorycontents.json"
  ]
}
```

## License

Released under the [MIT License](LICENSE).

