# AGENTS.md

This document helps agents work effectively with the Go Grep AST codebase.

## Project Overview

This is a Go tool that combines grep functionality with Abstract Syntax Tree (AST) analysis for powerful code search and pattern matching. It supports 40+ programming languages through tree-sitter integration and is heavily inspired by Aider-AI/grep-ast.

## Essential Commands

### Building and Running
- `task build` - Build the main executable
- `go build ./cmd/searchast` - Alternative build command
- `task searchast -- <args>` - Run searchast CLI with arguments
- `task overview -- <args>` - Run overview CLI with arguments

### Testing
- `task test` - Run all tests
- `task test-verbose` - Run tests with verbose output
- `task test-watch` - Run tests in watch mode (requires watchexec)
- `go test ./...` - Direct Go test command

### Code Quality
- `task fmt` - Format Go code
- `task fmt-check` - Check if Go code is formatted
- `task vet` - Run go vet to check for issues
- `task lint` - Run golangci-lint (must be installed)
- `task check` - Run all checks (format, vet, lint)
- `task clean` - Clean build artifacts

## Project Structure

```
/
├── cmd/
│   ├── searchast/     # Main CLI tool for pattern searching
│   └── overview/       # CLI tool for code structure overview
├── language/          # Language detection and tree-sitter mappings
├── context.go         # Context building logic
├── sourcetree.go      # Core AST parsing and tree structure
├── format.go          # Output formatting
├── set.go             # Set data structure helpers
├── *_test.go          # Test files
├── go.mod             # Go module definition
├── Taskfile.yml       # Task runner configuration
└── README.md          # Project documentation
```

## Code Patterns and Conventions

### Package Structure
- Main package is `searchast`
- CLI tools are in `cmd/` subdirectory with their own `main` packages
- Language support is in the `language/` package

### Core Components

1. **SourceTree** (`sourcetree.go`)
   - Main data structure representing parsed code
   - Created via `NewSourceTree(context, reader, filename)`
   - Contains lines with associated scope information

2. **ContextBuilder** (`context.go`)
   - Expands search results with context
   - Configurable options for surrounding lines, parent context, etc.
   - Default options: 3 surrounding lines, parent context enabled

3. **TextFormatter** (`format.go`)
   - Formats output with customizable symbols
   - Supports line numbers, highlight symbols, context symbols

### Naming Conventions
- Use `camelCase` for variables and functions
- Use `PascalCase` for types and exported functions
- Private functions and variables start with lowercase
- Error handling follows Go conventions: return errors, don't panic

### Testing Patterns
- Test files are named `*_test.go`
- Use table-driven tests for multiple test cases
- Helper functions use `must` prefix (e.g., `mustNewSourceTree`)
- Tests use `t.Helper()` for helper functions
- Use `strings.NewReader` for test input

### Key Functions

#### Creating a SourceTree
```go
sourceTree, err := searchast.NewSourceTree(context.Background(), file, filename)
```

#### Searching for Patterns
```go
linesOfInterest, err := sourceTree.Search(pattern)
```

#### Adding Context
```go
contextBuilder := searchast.NewContextBuilder(opts...)
linesToShow := contextBuilder.AddContext(sourceTree, linesOfInterest)
```

#### Formatting Output
```go
formatter := searchast.NewTextFormatter(formatterOpts...)
output := formatter.Format(sourceTree.Lines(), linesToShow, linesOfInterest)
```

## CLI Tools

### searchast
- Purpose: Pattern-based search with context
- Required flags: `-filename`, `-pattern`
- Optional flags: formatting options, line numbers
- Usage: `searchast -filename main.go -pattern "func.*\("`

### overview
- Purpose: High-level code structure overview
- Required flag: `-filename`
- Uses different context builder defaults for overview format
- Usage: `overview -filename main.go`

## Dependencies and Tools

### Go Version
- Requires Go 1.24.6
- Uses modern Go features like generics

### Key Dependencies
- `github.com/smacker/go-tree-sitter` - Tree-sitter bindings
- `github.com/alexaandru/go-sitter-forest/*` - Language parsers for 40+ languages

### Development Tools
- `golangci-lint` - Linting (must be installed separately)
- `watchexec` - For test watching mode
- Task runner (Taskfile.yml) - Alternative to Make

## Language Support

The tool supports 40+ programming languages through tree-sitter. Language detection is based on file extensions in `language/language.go`. Popular languages include Go, Python, JavaScript, TypeScript, Java, C, C++, Rust, HTML, CSS, Bash, and many more.

## Important Gotchas

### Memory Usage
- Large files may consume significant memory due to AST parsing
- The tool builds complete AST structure before searching

### Performance Considerations
- Tree-sitter parsing can be CPU-intensive for large files
- Consider file size limitations when processing very large codebases

### Error Handling
- The tool uses `log.Fatalf` for fatal errors in CLI tools
- Library functions return errors for caller handling
- File operations include proper cleanup with defer

### Scope Context
- Context building is computationally expensive for complex files
- Default settings balance readability and performance
- Can be customized via ContextBuilder options

## Testing Approach

### Unit Tests
- Test all public functions and methods
- Test error conditions and edge cases
- Use helper functions for common test setup

### Integration Testing
- Test CLI tools with actual source files
- Verify output format and content
- Test with different file types and patterns

### Test Organization
- Test functions grouped by functionality
- Subtests used for different scenarios
- Clear test names describing what is being tested