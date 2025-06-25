package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// isBinary checks if a file is likely to be binary
func isBinary(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Check for null bytes which typically indicate binary files
	for _, b := range data {
		if b == 0 {
			return true
		}
	}

	// Check if file has too many non-printable characters
	nonPrintable := 0
	for _, b := range data {
		if b < 32 && b != '\n' && b != '\r' && b != '\t' {
			nonPrintable++
		}
	}

	// If more than 30% non-printable, likely binary
	return float64(nonPrintable)/float64(len(data)) > 0.3
}

// shouldSkipFile determines if a file should be skipped based on its path
func shouldSkipFile(path string) bool {
	// Skip hidden files and directories
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") && part != "." {
			return true
		}
	}

	// Skip common binary extensions
	binaryExts := []string{
		".exe", ".dll", ".so", ".dylib", ".a", ".o",
		".jpg", ".jpeg", ".png", ".gif", ".bmp", ".ico", ".svg",
		".mp3", ".mp4", ".avi", ".mov", ".wav",
		".zip", ".tar", ".gz", ".bz2", ".7z", ".rar",
		".pdf", ".doc", ".docx", ".xls", ".xlsx",
		".pyc", ".pyo", ".class", ".jar",
		".db", ".sqlite", ".sqlite3",
	}

	ext := strings.ToLower(filepath.Ext(path))
	for _, binExt := range binaryExts {
		if ext == binExt {
			return true
		}
	}

	return false
}

// checkAndFixFile checks if a file ends with newline and fixes it if needed
func checkAndFixFile(path string, fix bool) (bool, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}

	// Skip empty files
	if len(data) == 0 {
		return true, nil
	}

	// Skip binary files
	if isBinary(data) {
		return true, nil
	}

	// Check if file ends with newline
	endsWithNewline := bytes.HasSuffix(data, []byte("\n"))

	if !endsWithNewline && fix {
		// Add newline at the end
		data = append(data, '\n')

		// Write back to file
		err = os.WriteFile(path, data, 0o644)
		if err != nil {
			return false, fmt.Errorf("failed to write file: %w", err)
		}

		return false, nil
	}

	return endsWithNewline, nil
}

// processRepository walks through the repository and processes files
func processRepository(repoPath string, fix bool) error {
	var totalFiles, fixedFiles, skippedFiles int
	var problematicFiles []string

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path for display
		relPath, err := filepath.Rel(repoPath, path)
		if err != nil {
			relPath = path
		}

		// Skip files that should be ignored
		if shouldSkipFile(relPath) {
			skippedFiles++
			return nil
		}

		totalFiles++

		// Check and potentially fix the file
		endsWithNewline, err := checkAndFixFile(path, fix)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", relPath, err)
			return nil
		}

		if !endsWithNewline {
			if fix {
				fixedFiles++
				fmt.Printf("Fixed: %s\n", relPath)
			} else {
				problematicFiles = append(problematicFiles, relPath)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk repository: %w", err)
	}

	// Print summary
	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total files checked: %d\n", totalFiles)
	fmt.Printf("Files skipped: %d\n", skippedFiles)

	if fix {
		fmt.Printf("Files fixed: %d\n", fixedFiles)
		if fixedFiles == 0 {
			fmt.Println("All files already end with newline!")
		}
	} else {
		fmt.Printf("Files missing newline: %d\n", len(problematicFiles))
		if len(problematicFiles) > 0 {
			fmt.Println("\nFiles that don't end with newline:")
			for _, file := range problematicFiles {
				fmt.Printf("  - %s\n", file)
			}
			fmt.Println("\nRun with -fix flag to automatically add newlines")
		} else {
			fmt.Println("All files end with newline!")
		}
	}

	return nil
}

func main() {
	var fix bool
	flag.BoolVar(&fix, "fix", false, "Fix files that don't end with newline")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [-fix] <repository_path>\n", os.Args[0])
		os.Exit(1)
	}

	repoPath := args[0]

	// Check if path exists
	info, err := os.Stat(repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is not a directory\n", repoPath)
		os.Exit(1)
	}

	// Process repository
	if err := processRepository(repoPath, fix); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
