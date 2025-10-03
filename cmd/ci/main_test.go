package main

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vrnvu/go-project-template/cmd/ci/coverage"
)

func captureOutput(t *testing.T, f func(context.Context, io.Writer, io.Writer) error) (string, string, error) {
	t.Helper()
	var stdoutBuf, stderrBuf bytes.Buffer
	err := f(context.Background(), &stdoutBuf, &stderrBuf)
	return stdoutBuf.String(), stderrBuf.String(), err
}

func TestRunWithBuild(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("slow/integration: build")
	}
	stdout, stderr, err := captureOutput(t, runBuild)
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Equal(t, "Building binary...\nBuild complete!\n", stdout)
}

func TestRunWithClean(t *testing.T) {
	t.Parallel()
	stdout, stderr, err := captureOutput(t, runClean)
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Equal(t, "Cleaning build artifacts...\nClean complete!\n", stdout)
}

func TestRunWithTestSlow(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("slow/integration: test-slow")
	}

	projectRoot, err := findProjectRoot()
	require.NoError(t, err)

	_, _, err = captureOutput(t, func(ctx context.Context, out, errw io.Writer) error {
		return runGoTestsWithCoverage(ctx, projectRoot, out, errw)
	})
	require.NoError(t, err)

	stdout, _, err := captureOutput(t, runGoToolCover)
	require.NoError(t, err)
	require.NotEmpty(t, stdout)

	coverageFunctions, err := coverage.GetFunctions(stdout)
	require.NoError(t, err)

	failures := coverage.Coverage(coverageFunctions)
	require.Empty(t, failures)
}

func TestRunWithBuildDocker(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("slow/integration: docker")
	}
	stdout, stderr, err := captureOutput(t, runBuildDocker)
	require.NoError(t, err)
	require.NotEmpty(t, stdout)
	require.NotEmpty(t, stderr)

	stdout, stderr, err = captureOutput(t, runDocker)
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.NotEmpty(t, stdout)

	stdout, stderr, err = captureOutput(t, runCleanDocker)
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.NotEmpty(t, stdout)
}
