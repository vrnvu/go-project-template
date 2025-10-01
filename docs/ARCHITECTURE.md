# Architecture

## System Overview

This Go project template provides a well-structured foundation for building maintainable, testable, and reproducible Go applications following modern best practices.

## Core Components

### 1. Build System (`cmd/ci/main.go`)
- **Single command setup**: `go run cmd/ci/main.go -setup` installs all dependencies
- **Fast/slow test split**: `go run cmd/ci/main.go -test` for fast tests, `-test-slow` for all tests
- **Quality gates**: Linting, testing, coverage, and size checks
- **Reproducible builds**: Pinned dependencies and deterministic outputs

### 2. Application Structure
- **CLI Interface** (`cmd/app/`): Main application entry point
- **Internal Packages** (`internal/`): Private application code
- **Documentation** (`docs/`): Architecture and contribution guidelines

### 3. CI/CD Pipeline (`.github/workflows/ci.yml`)
- **Automated testing**: Runs on push and pull requests
- **Concurrency control**: Prevents duplicate runs
- **Quality gates**: Lint, build, test, and size checks
- **Fast feedback**: Parallel job execution

## Project Structure

```
├── cmd/app/           # CLI application
├── internal/          # Private packages
├── docs/             # Documentation
├── .github/workflows/ # CI/CD pipeline
├── cmd/ci/main.go     # Go-based build system
└── .golangci.yml     # Linting configuration
```

## Design Principles

- **Two-party contracts**: Validate pre/post conditions at boundaries
- **Public API first**: Design for users, test public surfaces
- **Deterministic performance**: Explicit limits and pre-allocation
- **Error handling discipline**: Fail fast, precise error messages
- **Testing strategy**: Integration tests, property tests, snapshot tests
- **Documentation that can't drift**: Encode expectations as tests

## Quality Gates

- **Linting**: Comprehensive Go linting with golangci-lint
- **Testing**: Fast/slow test split with race detection
- **Coverage**: Test coverage reporting and thresholds
- **Size hygiene**: File size limits to keep clones fast
- **CI/CD**: Automated quality checks on every commit
