# Go Project Template

A well-structured Go project template following modern best practices and the "Basic Things" principles for maintainable, testable, and reproducible software development.

## Quick Start

```bash
# Setup development environment
go run cmd/ci/main.go -setup

# Build and run
go run cmd/ci/main.go -build
./bin/app --help

# Run fast tests (default)
go run cmd/ci/main.go -test

# Run all tests including slow ones
go run cmd/ci/main.go -test-slow

# Run with coverage
go run cmd/ci/main.go -test-coverage

# Run linter
go run cmd/ci/main.go -lint

# Clean build artifacts
go run cmd/ci/main.go -clean

# Run multiple commands
go run cmd/ci/main.go -build -test -lint
```

## Documentation

- [Architecture](docs/ARCHITECTURE.md) - System design and components
- [Contributing](docs/CONTRIBUTING.md) - Development process
- [Specification](SPEC.md) - Detailed requirements

## Features

- **Reproducible builds**: Single command setup and build
- **Quality gates**: Automated linting, testing, and size checks
- **Fast feedback**: Fast/slow test split with parallel execution
- **CI/CD ready**: GitHub Actions workflow with proper concurrency control
- **Documentation**: Clear architecture and contribution guidelines
