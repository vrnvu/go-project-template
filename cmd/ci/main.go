package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/vrnvu/go-project-template/cmd/ci/coverage"
)

func exitWith(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}

func main() {
	args := os.Args[1:]
	var (
		setup        = flag.Bool("setup", false, "Setup development environment")
		build        = flag.Bool("build", false, "Build the binary")
		buildDocker  = flag.Bool("build-docker", false, "Build Docker image")
		runDocker    = flag.Bool("run-docker", false, "Run Docker image")
		test         = flag.Bool("test", false, "Run fast tests")
		testSlow     = flag.Bool("test-slow", false, "Run all tests including slow ones with coverage")
		testCoverage = flag.Bool("test-coverage", false, "Check per-function coverage from coverage.out and enforce threshold")
		clean        = flag.Bool("clean", false, "Clean build artifacts")
		cleanDocker  = flag.Bool("clean-docker", false, "Clean Docker image")
		help         = flag.Bool("help", false, "Show help")
	)

	if err := flag.CommandLine.Parse(args); err != nil {
		exitWith(fmt.Errorf("failed to parse flags: %w", err))
	}

	if len(args) != 1 {
		flag.Usage()
		exitWith(fmt.Errorf("too many arguments: %v", args))
	}

	if *help {
		flag.Usage()
		return
	}

	if !*setup && !*build && !*buildDocker && !*runDocker && !*test && !*testSlow && !*testCoverage && !*clean && !*cleanDocker {
		flag.Usage()
		exitWith(fmt.Errorf("no action specified"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	if *setup {
		if err := runSetup(ctx, os.Stdout, os.Stderr); err != nil {
			exitWith(fmt.Errorf("setup failed: %w", err))
		}
	}
	if *build {
		if err := runBuild(ctx, os.Stdout, os.Stderr); err != nil {
			exitWith(fmt.Errorf("build failed: %w", err))
		}
	}
	if *buildDocker {
		if err := runBuildDocker(ctx, os.Stdout, os.Stderr); err != nil {
			exitWith(fmt.Errorf("build-docker failed: %w", err))
		}
	}
	if *runDocker {
		if err := runRunDocker(ctx, os.Stdout, os.Stderr); err != nil {
			exitWith(fmt.Errorf("run-docker failed: %w", err))
		}
	}
	if *test {
		if err := runTestFast(ctx, os.Stdout, os.Stderr); err != nil {
			exitWith(fmt.Errorf("fast tests failed: %w", err))
		}
	}
	if *testSlow {
		if err := runTestSlow(ctx, os.Stdout, os.Stderr); err != nil {
			exitWith(fmt.Errorf("tests failed: %w", err))
		}
	}
	if *testCoverage {
		if err := runCheckCoverage(ctx, os.Stdout, os.Stderr); err != nil {
			exitWith(fmt.Errorf("test coverage failed: %w", err))
		}
	}
	if *clean {
		if err := runClean(ctx, os.Stdout, os.Stderr); err != nil {
			exitWith(fmt.Errorf("clean failed: %w", err))
		}
	}
	if *cleanDocker {
		if err := runCleanDocker(ctx, os.Stdout, os.Stderr); err != nil {
			exitWith(fmt.Errorf("clean-docker failed: %w", err))
		}
	}
}

func runCommand(ctx context.Context, stdout, stderr io.Writer, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func runCommandInDir(ctx context.Context, dir string, stdout, stderr io.Writer, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr
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

func runSetup(ctx context.Context, stdout, stderr io.Writer) error {
	fmt.Fprintln(stdout, "Setting up development environment...")
	defer fmt.Fprintln(stdout, "Setup complete!")

	if err := runCommand(ctx, stdout, stderr, "go", "mod", "download"); err != nil {
		return fmt.Errorf("failed to download modules: %w", err)
	}
	if err := runCommand(ctx, stdout, stderr, "go", "mod", "verify"); err != nil {
		return fmt.Errorf("failed to verify modules: %w", err)
	}

	gopath, err := exec.CommandContext(ctx, "go", "env", "GOPATH").Output()
	if err != nil {
		return fmt.Errorf("failed to get GOPATH: %w", err)
	}

	if err := runCommand(ctx, stdout, stderr, "curl", "--version"); err != nil {
		return fmt.Errorf("curl is required but not available: %w", err)
	}

	gobin := filepath.Join(strings.TrimSpace(string(gopath)), "bin", "golangci-lint")
	if _, err := os.Stat(gobin); os.IsNotExist(err) {
		binDir := filepath.Dir(gobin)
		script := "https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh"
		if err := runCommand(ctx, stdout, stderr, "bash", "-lc", fmt.Sprintf("curl -sSfL %s | sh -s -- -b %s v2.5.0", script, binDir)); err != nil {
			return fmt.Errorf("failed to install golangci-lint via script: %w", err)
		}
	} else {
		fmt.Fprintln(stdout, "golangci-lint already installed")
	}

	if err := runCommand(ctx, stdout, stderr, "git", "--version"); err != nil {
		return fmt.Errorf("git is required but not available: %w", err)
	}
	if err := runCommand(ctx, stdout, stderr, "git", "rev-parse", "--is-inside-work-tree"); err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	return nil
}

func runBuild(ctx context.Context, stdout, stderr io.Writer) error {
	fmt.Fprintln(stdout, "Building binary...")
	defer fmt.Fprintln(stdout, "Build complete!")

	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	binDir := filepath.Join(projectRoot, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	outputPath := filepath.Join(binDir, "app")
	if err := runCommandInDir(ctx, projectRoot, stdout, stderr, "go", "build", "-o", outputPath, "./cmd/app"); err != nil {
		return fmt.Errorf("failed to build: %w", err)
	}

	return nil
}

func runBuildDocker(ctx context.Context, stdout, stderr io.Writer) error {
	fmt.Fprintln(stdout, "Checking if Docker is running...")
	if err := runCommand(ctx, stdout, stderr, "docker", "ps"); err != nil {
		return fmt.Errorf("docker is not running or not accessible: %w", err)
	}

	fmt.Fprintln(stdout, "Docker is running, building image...")
	defer fmt.Fprintln(stdout, "Docker image built!")

	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	if err := runCommandInDir(ctx, projectRoot, stdout, stderr, "docker", "build", "--platform", "linux/amd64", "-t", "app:latest", "."); err != nil {
		return fmt.Errorf("failed to build Docker image: %w", err)
	}

	return nil
}

func runRunDocker(ctx context.Context, stdout, stderr io.Writer) error {
	fmt.Fprintln(stdout, "Running Docker image...")
	defer fmt.Fprintln(stdout, "Docker run complete!")

	if err := runCommand(ctx, stdout, stderr, "docker", "run", "--rm", "app:latest"); err != nil {
		return fmt.Errorf("failed to run Docker image: %w", err)
	}

	return nil
}

func runTestFast(ctx context.Context, stdout, stderr io.Writer) error {
	fmt.Fprintln(stdout, "Running fast tests...")
	defer fmt.Fprintln(stdout, "Fast tests complete!")

	// Change to project root directory
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	if err := runLint(ctx, stdout, stderr); err != nil {
		return fmt.Errorf("linting failed: %w", err)
	}

	if err := runCheckSize(ctx); err != nil {
		return fmt.Errorf("size check failed: %w", err)
	}

	if err := runCommandInDir(ctx, projectRoot, stdout, stderr, "go", "test", "-count=1", "-race", "-short", "./..."); err != nil {
		return fmt.Errorf("failed to run fast tests: %w", err)
	}

	return nil
}

func runTestSlow(ctx context.Context, stdout, stderr io.Writer) error {
	fmt.Fprintln(stdout, "Running all tests...")
	defer fmt.Fprintln(stdout, "All tests complete!")

	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	if err := runLint(ctx, stdout, stderr); err != nil {
		return fmt.Errorf("linting failed: %w", err)
	}

	if err := runCheckSize(ctx); err != nil {
		return fmt.Errorf("size check failed: %w", err)
	}

	if err := runDocker(ctx, stdout, stderr); err != nil {
		return fmt.Errorf("docker failed: %w", err)
	}

	fmt.Fprintln(stdout, "Running tests with coverage...")
	if err := runGoTestsWithCoverage(ctx, projectRoot, stdout, stderr); err != nil {
		return fmt.Errorf("failed to run coverage tests: %w", err)
	}

	// Run only the coverage validation test in cmd/ci (no coverage flags)
	if err := runCommandInDir(ctx, projectRoot, stdout, stderr, "go", "test", "-count=1", "-race", "./cmd/ci", "-run", "TestRunWithTestSlow"); err != nil {
		return fmt.Errorf("coverage validation failed: %w", err)
	}
	fmt.Fprintln(stdout, "Coverage complete!")

	return nil
}

func runLint(ctx context.Context, stdout, stderr io.Writer) error {
	fmt.Fprintln(stdout, "Running linter...")
	defer fmt.Fprintln(stdout, "Linting complete!")

	gopath, err := exec.CommandContext(ctx, "go", "env", "GOPATH").Output()
	if err != nil {
		return fmt.Errorf("failed to get GOPATH: %w", err)
	}

	lintPath := filepath.Join(strings.TrimSpace(string(gopath)), "bin", "golangci-lint")
	return runCommand(ctx, stdout, stderr, lintPath, "run")
}

func runCheckSize(ctx context.Context) error {
	fmt.Fprintln(os.Stdout, "Checking file sizes...")
	defer fmt.Fprintln(os.Stdout, "File sizes checked!")

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

	var out bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "ls-files", "-z")
	cmd.Stdout = &out
	cmd.Stderr = io.Discard
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to list files with git: %w", err)
	}

	entries := strings.Split(out.String(), "\x00")
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
		}
	}

	return nil
}

func runClean(ctx context.Context, stdout, stderr io.Writer) error {
	fmt.Fprintln(stdout, "Cleaning build artifacts...")
	defer fmt.Fprintln(stdout, "Clean complete!")

	if err := os.RemoveAll("bin"); err != nil {
		return fmt.Errorf("failed to remove bin directory: %w", err)
	}

	if err := runCommand(ctx, stdout, stderr, "go", "clean"); err != nil {
		return fmt.Errorf("failed to clean Go cache: %w", err)
	}

	if err := os.Remove("coverage.out"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove coverage.out: %w", err)
	}
	return nil
}

func runCleanDocker(ctx context.Context, stdout, stderr io.Writer) error {
	fmt.Fprintln(stdout, "Cleaning Docker image...")
	defer fmt.Fprintln(stdout, "Clean complete!")

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

		if err := runCommand(ctx, stdout, stderr, "docker", "rmi", "app"); err != nil {
			return fmt.Errorf("failed to remove Docker image: %w", err)
		}
	}

	return nil
}

func runDocker(ctx context.Context, stdout, stderr io.Writer) error {
	fmt.Fprintln(stdout, "runDocker")
	defer fmt.Fprintln(stdout, "Docker run complete!")
	if err := runCommand(ctx, stdout, stderr, "docker", "ps"); err == nil {
		return nil
	}

	if err := runCommand(ctx, stdout, stderr, "open", "-a", "Docker"); err != nil {
		return fmt.Errorf("failed to launch Docker Desktop: %w", err)
	}

	time.Sleep(5 * time.Second)

	if err := runCommand(ctx, stdout, stderr, "docker", "ps"); err != nil {
		return fmt.Errorf("failed docker ps")
	}

	return nil
}

func runGoTestsWithCoverage(ctx context.Context, projectRoot string, stdout, stderr io.Writer) error {
	// List packages and exclude cmd/ci to avoid recursion and timing issues
	var listOut bytes.Buffer
	if err := runCommandInDir(ctx, projectRoot, &listOut, stderr, "go", "list", "./..."); err != nil {
		return err
	}
	pkgs := make([]string, 0, 16)
	for _, p := range strings.Fields(listOut.String()) {
		if strings.HasSuffix(p, "/cmd/ci") {
			continue
		}
		pkgs = append(pkgs, p)
	}
	if len(pkgs) == 0 {
		return nil
	}

	args := append([]string{"test", "-count=1", "-race", "-covermode=atomic", "-coverprofile=coverage.out"}, pkgs...)
	return runCommandInDir(ctx, projectRoot, stdout, stderr, "go", args...)
}

func runGoToolCover(ctx context.Context, stdout, stderr io.Writer) error {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return err
	}

	if err := runCommandInDir(ctx, projectRoot, stdout, stderr, "go", "tool", "cover", "-func=coverage.out"); err != nil {
		return err
	}
	return nil
}

// runCheckCoverage reads coverage.out via `go tool cover -func`, enforces a threshold with a whitelist.
func runCheckCoverage(ctx context.Context, stdout, stderr io.Writer) error {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := runCommandInDir(ctx, projectRoot, &buf, stderr, "go", "tool", "cover", "-func=coverage.out"); err != nil {
		return fmt.Errorf("failed to show coverage: %w", err)
	}

	// Echo the coverage summary to stdout for visibility and to satisfy linters
	_, _ = fmt.Fprint(stdout, buf.String())

	functions, err := coverage.GetFunctions(buf.String())
	if err != nil {
		return err
	}

	const threshold = 70.0
	var failures []string
	for _, fn := range functions {
		if wl, ok := coverage.WhiteList[fn.FileName]; ok && wl == fn.FuncName {
			continue
		}
		if fn.Percentage < threshold {
			failures = append(failures, fmt.Sprintf("%s %s %.1f%% < %.1f%%", fn.FileName, fn.FuncName, fn.Percentage, threshold))
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("coverage below threshold:\n%s", strings.Join(failures, "\n"))
	}
	return nil
}
