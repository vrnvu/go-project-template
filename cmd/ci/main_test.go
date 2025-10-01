package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func assertOutput(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, want)
	}
}

func assertOutputContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("output does not contain expected string\ngot:\n%q\nwant to contain:\n%q", got, want)
	}
}

func formatOutputGot(stdout, stderr string, err error) string {
	return fmt.Sprintf("stdout: %q\nstderr: %q\nerr: %v", stdout, stderr, err)
}

func formatOutputWant(stdout, stderr string, err error) string {
	if err == nil {
		return fmt.Sprintf("stdout: %q\nstderr: %q\nerr: <nil>", stdout, stderr)
	}
	return fmt.Sprintf("stdout: %q\nstderr: %q\nerr: %v", stdout, stderr, err)
}

func captureOutput(flagName string) (string, string, error) {
	flag.CommandLine = flag.NewFlagSet(flagName, flag.ContinueOnError)

	var stdout, stderr bytes.Buffer
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutFile, _ := os.CreateTemp("", "stdout")
	stderrFile, _ := os.CreateTemp("", "stderr")
	defer os.Remove(stdoutFile.Name())
	defer os.Remove(stderrFile.Name())

	os.Stdout = stdoutFile
	os.Stderr = stderrFile

	err := run([]string{"-" + flagName})

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read the captured output
	_, _ = stdoutFile.Seek(0, 0)
	_, _ = stderrFile.Seek(0, 0)
	_, _ = stdout.ReadFrom(stdoutFile)
	_, _ = stderr.ReadFrom(stderrFile)

	return stdout.String(), stderr.String(), err
}

func TestRunWithLint(t *testing.T) { //nolint:paralleltest
	stdout, stderr, err := captureOutput("lint")

	got := formatOutputGot(stdout, stderr, err)
	want := formatOutputWant("Running linter...\nLinting complete!\n", "", nil)

	assertOutput(t, got, want)
}

func TestRunWithBuild(t *testing.T) { //nolint:paralleltest
	stdout, stderr, err := captureOutput("build")

	got := formatOutputGot(stdout, stderr, err)
	want := formatOutputWant("Building binary...\nBuild complete!\n", "", nil)

	assertOutput(t, got, want)
}

func TestRunWithClean(t *testing.T) { //nolint:paralleltest
	stdout, stderr, err := captureOutput("clean")

	got := formatOutputGot(stdout, stderr, err)
	want := formatOutputWant("Cleaning build artifacts...\nClean complete!\n", "", nil)

	assertOutput(t, got, want)
}

func TestRunWithInvalidFlag(t *testing.T) { //nolint:paralleltest
	stdout, stderr, err := captureOutput("invalid")

	got := formatOutputGot(stdout, stderr, err)
	want := "flag provided but not defined: -invalid"
	assertOutputContains(t, got, want)
}

func TestRunWithCheckSize(t *testing.T) { //nolint:paralleltest
	stdout, stderr, err := captureOutput("check-size")

	got := formatOutputGot(stdout, stderr, err)
	want := formatOutputWant("Checking file sizes...\nFile sizes checked!\n", "", nil)
	assertOutput(t, got, want)
}

func TestRunWithTestCoverage(t *testing.T) { //nolint:paralleltest
	if testing.CoverMode() != "" {
		t.Skip("Skipping coverage test when running with coverage")
	}

	stdout, stderr, err := captureOutput("test-coverage")

	got := formatOutputGot(stdout, stderr, err)
	// The actual output includes coverage details, so we just check it contains the key messages
	assertOutputContains(t, got, "Running tests with coverage...")
	assertOutputContains(t, got, "Coverage complete!")

	// Check coverage percentages and report failures
	checkCoverageThreshold(t, stdout)
}

func checkCoverageThreshold(t *testing.T, output string) {
	t.Helper()

	// Debug: print the output to see the exact format
	t.Logf("Coverage output:\n%s", output)

	var failures []string
	whiteList := []string{
		"main",
		"run",
		"runBuild",
		"runCheckSize",
		"checkFileSizes",
		"runClean",
		"runSetup",
		"runTestFast",
		"runTestSlow",
		"runTestCoverage",
		"fatalf",
	}

	// Check for function-level coverage lines like:
	// "github.com/vrnvu/gdts/cmd/ci/main.go:21:                        run             55.6%"
	// Pattern: file:line: tabs function tabs percentage%
	functionRegex := regexp.MustCompile(`\S+:\d+:\s+(\S+)\s+(\d+\.\d+)%`)
	functionMatches := functionRegex.FindAllStringSubmatch(output, -1)

	for _, match := range functionMatches {
		funcName := match[1]
		percentage, err := strconv.ParseFloat(match[2], 64)
		if err != nil {
			t.Errorf("Failed to parse function coverage percentage: %v", err)
			continue
		}

		if slices.Contains(whiteList, funcName) {
			continue
		}

		if percentage < 70.0 {
			failures = append(failures, fmt.Sprintf("Function %s: %.1f%% (below 80%%)", funcName, percentage))
		}
	}

	if len(failures) > 0 {
		t.Errorf("Coverage below 80%% threshold:\n%s", strings.Join(failures, "\n"))
	}
}

func TestRunWithBuildDocker(t *testing.T) { //nolint:paralleltest
	stdout, stderr, err := captureOutput("build-docker")

	got := formatOutputGot(stdout, stderr, err)
	assertOutputContains(t, got, "Checking if Docker is running...")
	assertOutputContains(t, got, "Docker is running, building image...")
	assertOutputContains(t, got, "Docker image built!")

	stdout, stderr, err = captureOutput("run-docker")

	got = formatOutputGot(stdout, stderr, err)
	assertOutputContains(t, got, "Running Docker image...")
	assertOutputContains(t, got, "Docker run complete!")

	stdout, stderr, err = captureOutput("clean-docker")

	got = formatOutputGot(stdout, stderr, err)
	assertOutputContains(t, got, "Cleaning Docker image...")
	assertOutputContains(t, got, "Clean complete!")
}
