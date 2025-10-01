package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fatalf("%v", err)
	}
}

// run executes the CI tool with the given arguments
func run(args []string) error {
	var (
		setup       = flag.Bool("setup", false, "Setup development environment")
		build       = flag.Bool("build", false, "Build the binary")
		buildDocker = flag.Bool("build-docker", false, "Build Docker image")
		runDocker   = flag.Bool("run-docker", false, "Run Docker image")
		test        = flag.Bool("test", false, "Run fast tests")
		testSlow    = flag.Bool("test-slow", false, "Run all tests including slow ones")
		testCover   = flag.Bool("test-coverage", false, "Run tests with coverage")
		lint        = flag.Bool("lint", false, "Run linter")
		checkSize   = flag.Bool("check-size", false, "Check file sizes")
		clean       = flag.Bool("clean", false, "Clean build artifacts")
		cleanDocker = flag.Bool("clean-docker", false, "Clean Docker image")
		help        = flag.Bool("help", false, "Show help")
	)

	// Parse flags from the provided args
	if err := flag.CommandLine.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	if *help {
		flag.Usage()
		return nil
	}

	if !*setup && !*build && !*buildDocker && !*runDocker && !*test && !*testSlow && !*testCover && !*lint && !*checkSize && !*clean && !*cleanDocker {
		flag.Usage()
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	if *setup {
		if err := runSetup(ctx); err != nil {
			return fmt.Errorf("setup failed: %w", err)
		}
	}
	if *build {
		if err := runBuild(ctx); err != nil {
			return fmt.Errorf("build failed: %w", err)
		}
	}
	if *buildDocker {
		if err := runBuildDocker(ctx); err != nil {
			return fmt.Errorf("build-docker failed: %w", err)
		}
	}
	if *runDocker {
		if err := runRunDocker(ctx); err != nil {
			return fmt.Errorf("run-docker failed: %w", err)
		}
	}
	if *test {
		if err := runTestFast(ctx); err != nil {
			return fmt.Errorf("fast tests failed: %w", err)
		}
	}
	if *testSlow {
		if err := runTestSlow(ctx); err != nil {
			return fmt.Errorf("tests failed: %w", err)
		}
	}
	if *testCover {
		if err := runTestCoverage(ctx); err != nil {
			return fmt.Errorf("coverage tests failed: %w", err)
		}
	}
	if *lint {
		if err := runLint(ctx); err != nil {
			return fmt.Errorf("linting failed: %w", err)
		}
	}
	if *checkSize {
		if err := runCheckSize(); err != nil {
			return fmt.Errorf("size check failed: %w", err)
		}
	}
	if *clean {
		if err := runClean(ctx); err != nil {
			return fmt.Errorf("clean failed: %w", err)
		}
	}
	if *cleanDocker {
		if err := runCleanDocker(ctx); err != nil {
			return fmt.Errorf("clean-docker failed: %w", err)
		}
	}

	return nil
}

func runCommand(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCommandInDir(ctx context.Context, dir, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func findProjectRoot() (string, error) {
	// Look for go.mod file by walking up the directory tree
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func runSetup(ctx context.Context) error {
	fmt.Println("Setting up development environment...")
	defer fmt.Println("Setup complete!")

	if err := runCommand(ctx, "go", "mod", "download"); err != nil {
		return fmt.Errorf("failed to download modules: %w", err)
	}
	if err := runCommand(ctx, "go", "mod", "verify"); err != nil {
		return fmt.Errorf("failed to verify modules: %w", err)
	}

	gopath, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		return fmt.Errorf("failed to get GOPATH: %w", err)
	}

	if err := runCommand(ctx, "curl", "--version"); err != nil {
		return fmt.Errorf("curl is required but not available: %w", err)
	}

	gobin := filepath.Join(strings.TrimSpace(string(gopath)), "bin", "golangci-lint")
	if _, err := os.Stat(gobin); os.IsNotExist(err) {
		binDir := filepath.Dir(gobin)
		script := "https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh"
		if err := runCommand(ctx, "bash", "-lc", fmt.Sprintf("curl -sSfL %s | sh -s -- -b %s v2.5.0", script, binDir)); err != nil {
			return fmt.Errorf("failed to install golangci-lint via script: %w", err)
		}
	} else {
		fmt.Println("golangci-lint already installed")
	}

	if err := runCommand(ctx, "git", "--version"); err != nil {
		return fmt.Errorf("git is required but not available: %w", err)
	}
	if err := runCommand(ctx, "git", "rev-parse", "--is-inside-work-tree"); err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	return nil
}

func runBuild(ctx context.Context) error {
	fmt.Println("Building binary...")
	defer fmt.Println("Build complete!")

	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	binDir := filepath.Join(projectRoot, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	outputPath := filepath.Join(binDir, "app")
	if err := runCommandInDir(ctx, projectRoot, "go", "build", "-o", outputPath, "./cmd/app"); err != nil {
		return fmt.Errorf("failed to build: %w", err)
	}

	return nil
}

func runBuildDocker(ctx context.Context) error {
	fmt.Println("Checking if Docker is running...")
	if err := runCommand(ctx, "docker", "ps"); err != nil {
		return fmt.Errorf("docker is not running or not accessible: %w", err)
	}

	fmt.Println("Docker is running, building image...")
	defer fmt.Println("Docker image built!")

	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	if err := runCommandInDir(ctx, projectRoot, "docker", "build", "--platform", "linux/amd64", "-t", "app:latest", "."); err != nil {
		return fmt.Errorf("failed to build Docker image: %w", err)
	}

	return nil
}

func runRunDocker(ctx context.Context) error {
	fmt.Println("Running Docker image...")
	defer fmt.Println("Docker run complete!")

	if err := runCommand(ctx, "docker", "run", "--rm", "app:latest"); err != nil {
		return fmt.Errorf("failed to run Docker image: %w", err)
	}

	return nil
}

func runTestFast(ctx context.Context) error {
	fmt.Println("Running fast tests...")
	defer fmt.Println("Fast tests complete!")

	// Change to project root directory
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	if err := runCommandInDir(ctx, projectRoot, "go", "test", "-count=1", "-race", "-short", "./..."); err != nil {
		return fmt.Errorf("failed to run fast tests: %w", err)
	}
	return nil
}

func runTestSlow(ctx context.Context) error {
	fmt.Println("Running all tests...")
	defer fmt.Println("All tests complete!")

	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	if err := runCommandInDir(ctx, projectRoot, "go", "test", "-count=1", "-race", "./..."); err != nil {
		return fmt.Errorf("failed to run tests: %w", err)
	}
	return nil
}

func runTestCoverage(ctx context.Context) error {
	fmt.Println("Running tests with coverage...")
	defer fmt.Println("Coverage complete!")

	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	if err := runCommandInDir(ctx, projectRoot, "go", "test", "-covermode=atomic", "-coverprofile=coverage.out", "./..."); err != nil {
		return fmt.Errorf("failed to run coverage tests: %w", err)
	}
	if err := runCommandInDir(ctx, projectRoot, "go", "tool", "cover", "-func=coverage.out"); err != nil {
		return fmt.Errorf("failed to show coverage: %w", err)
	}
	return nil
}

func runLint(ctx context.Context) error {
	fmt.Println("Running linter...")
	defer fmt.Println("Linting complete!")

	gopath, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		return fmt.Errorf("failed to get GOPATH: %w", err)
	}

	lintPath := filepath.Join(strings.TrimSpace(string(gopath)), "bin", "golangci-lint")
	return runCommand(ctx, lintPath, "run")
}

func runCheckSize() error {
	fmt.Println("Checking file sizes...")
	defer fmt.Println("File sizes checked!")

	kb := int64(1024)
	extLimits := map[string]int64{
		".go":  20 * kb,
		".md":  10 * kb,
		".mod": 10 * kb,
		".sum": 10 * kb,
		".yml": 10 * kb,
	}
	fileLimits := map[string]int64{
		".gitignore": 1 * kb,
		"Dockerfile": 10 * kb,
	}

	cmd := exec.Command("git", "ls-files", "-z")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list files with git: %w", err)
	}

	entries := strings.Split(string(out), "\x00")
	for _, path := range entries {
		if path == "" {
			continue
		}

		base := filepath.Base(path)
		if limit, ok := fileLimits[base]; ok {
			info, statErr := os.Stat(path)
			if statErr != nil {
				return fmt.Errorf("size-check: os.Stat failed for file: %s", path)
			}
			if info.Size() > limit {
				return fmt.Errorf("size-check: %s: %s (%d bytes)", base, path, info.Size())
			}
			continue
		}

		ext := strings.ToLower(filepath.Ext(base))
		if limit, ok := extLimits[ext]; ok {
			info, statErr := os.Stat(path)
			if statErr != nil {
				return fmt.Errorf("size-check: os.Stat failed for extension: %s %s", ext, path)
			}
			if info.Size() > limit {
				return fmt.Errorf("size-check: %s: %s (%d bytes)", ext, path, info.Size())
			}
		} else {
			return fmt.Errorf("size-check: format not supported: %s %s", ext, path)
		}
	}

	return nil
}

func runClean(ctx context.Context) error {
	fmt.Println("Cleaning build artifacts...")
	defer fmt.Println("Clean complete!")

	if err := os.RemoveAll("bin"); err != nil {
		return fmt.Errorf("failed to remove bin directory: %w", err)
	}

	if err := runCommand(ctx, "go", "clean"); err != nil {
		return fmt.Errorf("failed to clean Go cache: %w", err)
	}

	if err := os.Remove("coverage.out"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove coverage.out: %w", err)
	}
	return nil
}

func runCleanDocker(ctx context.Context) error {
	fmt.Println("Cleaning Docker image...")
	defer fmt.Println("Clean complete!")

	cmd := exec.CommandContext(ctx, "docker", "ps", "-aq", "--filter", "ancestor=app")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		containerIDs := strings.TrimSpace(string(output))
		if containerIDs != "" {
			stopArgs := append([]string{"stop"}, strings.Fields(containerIDs)...)
			stopCmd := exec.CommandContext(ctx, "docker", stopArgs...)
			if err := stopCmd.Run(); err != nil {
				return fmt.Errorf("failed to stop containers: %w", err)
			}

			rmArgs := append([]string{"rm"}, strings.Fields(containerIDs)...)
			rmCmd := exec.CommandContext(ctx, "docker", rmArgs...)
			if err := rmCmd.Run(); err != nil {
				return fmt.Errorf("failed to remove containers: %w", err)
			}
		}

		if err := runCommand(ctx, "docker", "rmi", "app"); err != nil {
			return fmt.Errorf("failed to remove Docker image: %w", err)
		}
	}

	return nil
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
