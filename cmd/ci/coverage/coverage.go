package coverage

import (
	"fmt"
	"regexp"
	"strconv"
)

type Function struct {
	FileName   string
	FuncName   string
	Percentage float64
}

// WhiteList is a map of file names to function names that are exempt from coverage checks.
var WhiteList = map[string]string{
	"github.com/vrnvu/go-project-template/cmd/app/main.go":          "main",
	"github.com/vrnvu/go-project-template/internal/circuit/time.go": "Now",
}

// Coverage checks if the coverage percentage is below the threshold.
func Coverage(functions []Function) []Function {
	var failures []Function
	for _, match := range functions {
		if funcName, ok := WhiteList[match.FileName]; ok {
			if funcName == match.FuncName {
				continue
			}
		}

		if match.Percentage < 70.0 {
			failures = append(failures, match)
		}
	}

	return failures
}

// GetFunctions parses `go tool cover -func` output into functions.
func GetFunctions(output string) ([]Function, error) {
	functionRegex := regexp.MustCompile(`(?m)^\s*(github.com/vrnvu/go-project-template/\S+):\d+:\s+([A-Za-z0-9_]+)\s+(\d+(?:\.\d+)?)%`)
	rows := functionRegex.FindAllStringSubmatch(output, -1)
	if len(rows) == 0 {
		return nil, fmt.Errorf("no function coverage lines matched; check output format or regex")
	}
	functions := make([]Function, 0, len(rows))
	for _, r := range rows {
		pct, err := strconv.ParseFloat(r[3], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse function coverage percentage: %w", err)
		}
		functions = append(functions, Function{FileName: r[1], FuncName: r[2], Percentage: pct})
	}

	return functions, nil
}
