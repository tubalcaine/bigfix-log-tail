# BigFix Log Tail - Development Guide

## Build & Run Commands
```
# Build the application
go build

# Run with default settings
./bigfix-log-tail

# Run with custom log directory
./bigfix-log-tail /path/to/logs

# Run with custom microsecond delay
./bigfix-log-tail -usec 200
```

## Code Style Guidelines
- **Imports**: Standard library first, then third-party packages, alphabetically ordered
- **Formatting**: Use `gofmt` to format code before committing
- **Variables**: Use camelCase for local variables, avoid globals when possible
- **Errors**: Always check and handle errors, prefer returning errors over panicking
- **Comments**: Package and exported functions must be documented with comments
- **Readability**: Prefer clarity over cleverness, use meaningful variable names
- **Goroutines**: Close channels properly to avoid goroutine leaks

