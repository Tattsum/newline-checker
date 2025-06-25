//go:build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
var Default = Build

// Build builds the binary
func Build() error {
	fmt.Println("🔨 Building binary...")
	return sh.Run("go", "build", "-o", getBinaryName(), ".")
}

// Test runs the test suite
func Test() error {
	fmt.Println("🧪 Running tests...")
	return sh.Run("go", "test", "-v", "-cover", "./...")
}

// TestCoverage runs tests with coverage report
func TestCoverage() error {
	fmt.Println("📊 Running tests with coverage...")
	if err := sh.Run("go", "test", "-coverprofile=coverage.out", "./..."); err != nil {
		return err
	}
	return sh.Run("go", "tool", "cover", "-html=coverage.out", "-o=coverage.html")
}

// Benchmark runs benchmark tests
func Benchmark() error {
	fmt.Println("⚡ Running benchmarks...")
	return sh.Run("go", "test", "-bench=.", "-benchmem", "./...")
}

// Fmt formats Go source code using gofumpt
func Fmt() error {
	fmt.Println("🎨 Formatting code...")

	// Check if gofumpt is available
	if err := sh.Run("gofumpt", "-version"); err != nil {
		fmt.Println("⚠️  gofumpt not found, installing...")
		if err := sh.Run("go", "install", "mvdan.cc/gofumpt@latest"); err != nil {
			return fmt.Errorf("failed to install gofumpt: %w", err)
		}
		fmt.Println("✅ gofumpt installed successfully")
	}

	// Run gofumpt
	if err := sh.Run("gofumpt", "-w", "."); err != nil {
		return fmt.Errorf("gofumpt failed: %w", err)
	}

	fmt.Println("✅ Code formatted with gofumpt")
	return nil
}

// Lint runs linting tools
func Lint() error {
	fmt.Println("🔍 Running linters...")

	// go vet
	fmt.Println("Running go vet...")
	if err := sh.Run("go", "vet", "./..."); err != nil {
		return fmt.Errorf("go vet failed: %w", err)
	}

	// golangci-lint if available
	if err := sh.Run("golangci-lint", "run", "./..."); err != nil {
		fmt.Println("⚠️  golangci-lint not available, skipping (install from: https://golangci-lint.run/)")
		fmt.Println("    Quick install: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin")
	} else {
		fmt.Println("✅ golangci-lint passed")
	}

	// staticcheck if available
	if err := sh.Run("staticcheck", "./..."); err != nil {
		fmt.Println("⚠️  staticcheck not available, skipping (install with: go install honnef.co/go/tools/cmd/staticcheck@latest)")
	} else {
		fmt.Println("✅ staticcheck passed")
	}

	return nil
}

// Clean removes build artifacts
func Clean() error {
	fmt.Println("🧹 Cleaning build artifacts...")

	filesToRemove := []string{
		getBinaryName(),
		"coverage.out",
		"coverage.html",
	}

	for _, file := range filesToRemove {
		if _, err := os.Stat(file); err == nil {
			if err := os.Remove(file); err != nil {
				fmt.Printf("Warning: failed to remove %s: %v\n", file, err)
			} else {
				fmt.Printf("Removed %s\n", file)
			}
		}
	}

	return nil
}

// Install installs the binary to GOPATH/bin
func Install() error {
	fmt.Println("📦 Installing binary...")
	return sh.Run("go", "install", ".")
}

// Check runs formatting, linting, and tests
func Check() error {
	fmt.Println("🔍 Running full check (fmt, lint, test)...")
	mg.Deps(Fmt, Lint, Test)
	return nil
}

// CI runs the full CI pipeline
func CI() error {
	fmt.Println("🚀 Running CI pipeline...")
	mg.Deps(Fmt, Lint, Test, Build)
	return nil
}

// Dev sets up development environment
func Dev() error {
	fmt.Println("🛠️  Setting up development environment...")

	tools := []struct {
		name string
		pkg  string
	}{
		{"gofumpt", "mvdan.cc/gofumpt@latest"},
		{"goimports", "golang.org/x/tools/cmd/goimports@latest"},
		{"staticcheck", "honnef.co/go/tools/cmd/staticcheck@latest"},
		{"golangci-lint", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"},
	}

	for _, tool := range tools {
		fmt.Printf("Installing %s...\n", tool.name)
		if err := sh.Run("go", "install", tool.pkg); err != nil {
			fmt.Printf("Warning: failed to install %s: %v\n", tool.name, err)
		} else {
			fmt.Printf("✅ Installed %s\n", tool.name)
		}
	}

	return nil
}

// Release builds binaries for multiple platforms
func Release() error {
	fmt.Println("🚀 Building release binaries...")

	platforms := []struct {
		goos   string
		goarch string
	}{
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"windows", "amd64"},
	}

	// Create release directory
	if err := os.MkdirAll("dist", 0o755); err != nil {
		return fmt.Errorf("failed to create dist directory: %w", err)
	}

	for _, platform := range platforms {
		env := map[string]string{
			"GOOS":   platform.goos,
			"GOARCH": platform.goarch,
		}

		binaryName := fmt.Sprintf("check-new-line-%s-%s", platform.goos, platform.goarch)
		if platform.goos == "windows" {
			binaryName += ".exe"
		}

		outputPath := filepath.Join("dist", binaryName)

		fmt.Printf("Building %s/%s -> %s\n", platform.goos, platform.goarch, outputPath)

		if err := sh.RunWith(env, "go", "build", "-ldflags", "-s -w", "-o", outputPath, "."); err != nil {
			return fmt.Errorf("failed to build %s/%s: %w", platform.goos, platform.goarch, err)
		}
	}

	fmt.Println("✅ Release binaries built successfully!")
	return nil
}

// ModTidy runs go mod tidy
func ModTidy() error {
	fmt.Println("📦 Running go mod tidy...")
	return sh.Run("go", "mod", "tidy")
}

// ModUpdate updates dependencies
func ModUpdate() error {
	fmt.Println("📦 Updating dependencies...")
	return sh.Run("go", "get", "-u", "./...")
}

// Security runs security analysis
func Security() error {
	fmt.Println("🔒 Running security analysis...")

	// gosec if available
	if err := sh.Run("gosec", "./..."); err != nil {
		fmt.Println("⚠️  gosec not available, skipping (install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)")
	} else {
		fmt.Println("✅ gosec passed")
	}

	// nancy if available (for dependency vulnerabilities)
	if err := sh.Run("nancy", "sleuth", "--loud"); err != nil {
		fmt.Println("⚠️  nancy not available, skipping (install with: go install github.com/sonatypecommunity/nancy@latest)")
	} else {
		fmt.Println("✅ nancy passed")
	}

	return nil
}

// Deps downloads dependencies
func Deps() error {
	fmt.Println("📦 Downloading dependencies...")
	return sh.Run("go", "mod", "download")
}

// getBinaryName returns the appropriate binary name for the current platform
func getBinaryName() string {
	binaryName := "check-new-line"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	return binaryName
}

// checkToolAvailable checks if a tool is available in PATH
func checkToolAvailable(tool string) bool {
	_, err := sh.Output("which", tool)
	return err == nil
}

// listTargets lists all available mage targets
func ListTargets() {
	fmt.Println("Available targets:")
	targets := []string{
		"build      - Build the binary",
		"test       - Run tests",
		"testcoverage - Run tests with coverage report",
		"benchmark  - Run benchmark tests",
		"fmt        - Format code with gofumpt",
		"lint       - Run linters",
		"clean      - Remove build artifacts",
		"install    - Install binary to GOPATH/bin",
		"check      - Run fmt, lint, and test",
		"ci         - Run full CI pipeline",
		"dev        - Set up development environment",
		"release    - Build release binaries for multiple platforms",
		"modtidy    - Run go mod tidy",
		"modupdate  - Update dependencies",
		"security   - Run security analysis",
		"deps       - Download dependencies",
	}

	for _, target := range targets {
		fmt.Printf("  %s\n", target)
	}
}
