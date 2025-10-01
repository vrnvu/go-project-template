# Contributing

## Development Setup

```bash
# Clone and setup
git clone <repo>
cd go-database-tx-simulator
make setup

# Run tests
make test
make test-property
```

## Code Review Process

### Ownership
- Core logic: `internal/simulator/` - requires maintainer review
- CLI interface: `cmd/gdts/` - any contributor can review
- Documentation: `docs/` - any contributor can review

### Review Goals
- **Correctness**: Logic matches specification
- **Performance**: No unnecessary allocations or copying
- **Clarity**: Code reads like documentation
- **Testing**: Adequate coverage for changes

### Review Checklist
- [ ] Two-party contracts validated at boundaries
- [ ] No panic-driven control flow
- [ ] Tests cover both success and error paths
- [ ] Performance implications considered
- [ ] Documentation updated if needed

## Quality Gates

- All tests pass: `make test`
- Linting clean: `make lint`
- No large files: `make check-size`
- Property tests: `make test-property`

## Release Process

- Weekly releases on Fridays
- Update version in `go.mod`
- Tag release: `git tag v1.0.0`
- Update changelog in `docs/`
