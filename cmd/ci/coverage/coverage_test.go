package coverage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCoverageFunctions(t *testing.T) {
	t.Parallel()

	output := `
	Running all tests...
	Running linter...
	0 issues.
	Linting complete!
	Checking file sizes...
	File sizes checked!
	CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS    PORTS     NAMES
	Running tests with coverage...
			github.com/vrnvu/go-project-template/cmd/app            coverage: 0.0% of statements
	ok      github.com/vrnvu/go-project-template/cmd/ci     12.481s coverage: 33.5% of statements
	ok      github.com/vrnvu/go-project-template/internal/circuit   1.802s  coverage: 97.1% of statements
	github.com/vrnvu/go-project-template/cmd/app/main.go:5:                 main                    0.0%
	github.com/vrnvu/go-project-template/cmd/ci/main.go:100:                runCommand              12.3%
	github.com/vrnvu/go-project-template/internal/circuit/cb.go:24:         asserts                 100.0%
	total:                                                                  (statements)            53.9%
	Coverage complete!
	All tests complete!
	`

	functions, err := GetFunctions(output)
	assert.NotNil(t, functions)
	assert.NoError(t, err)

	assert.Equal(t, 3, len(functions))
	assert.Equal(t, "github.com/vrnvu/go-project-template/cmd/app/main.go", functions[0].FileName)
	assert.Equal(t, "main", functions[0].FuncName)
	assert.Equal(t, 0.0, functions[0].Percentage)

	assert.Equal(t, "github.com/vrnvu/go-project-template/cmd/ci/main.go", functions[1].FileName)
	assert.Equal(t, "runCommand", functions[1].FuncName)
	assert.Equal(t, 12.3, functions[1].Percentage)

	assert.Equal(t, "github.com/vrnvu/go-project-template/internal/circuit/cb.go", functions[2].FileName)
	assert.Equal(t, "asserts", functions[2].FuncName)
	assert.Equal(t, 100.0, functions[2].Percentage)
}

func TestCoverage(t *testing.T) {
	t.Parallel()

	// app/main.go is whitelisted
	functions := []Function{
		{FileName: "github.com/vrnvu/go-project-template/cmd/app/main.go", FuncName: "main", Percentage: 0.0},
		{FileName: "github.com/vrnvu/go-project-template/cmd/ci/main.go", FuncName: "runCommand", Percentage: 12.3},
		{FileName: "github.com/vrnvu/go-project-template/internal/circuit/cb.go", FuncName: "asserts", Percentage: 100.0},
	}
	failures := Coverage(functions)
	assert.Equal(t, 1, len(failures))

	assert.Equal(t, "github.com/vrnvu/go-project-template/cmd/ci/main.go", failures[0].FileName)
	assert.Equal(t, "runCommand", failures[0].FuncName)
	assert.Equal(t, 12.3, failures[0].Percentage)
}
