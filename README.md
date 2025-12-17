# Go Grep AST

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

A Go tool that combines grep functionality with Abstract Syntax Tree (AST) analysis for powerful code search and pattern matching. This tool is heavily inspired by [Aider-AI/grep-ast](https://github.com/Aider-AI/grep-ast) and provides similar functionality with Go performance and tree-sitter integration.

## Installation

### Install as CLI Tool

#### Using Go Install

```bash
go install github.com/andersonjoseph/searchast/cmd/searchast@latest
```

### Install as Go Package

```bash
go get github.com/andersonjoseph/searchast
```

## Usage

### CLI Usage

The project provides two main CLI tools:

#### searchast - Pattern-based search

**Basic usage:**

```bash
searchast -filename <file> -pattern <regex>
```

**Examples:**

##### Example 1: Find all function definitions

```bash
searchast -filename main.go -pattern "func.*\\("
```

**Output:**
```
  ⋮
 3 │ import (
 4 │ 	"fmt"
 5 │ 	"log"
 6 │ 	"os"
 7 │ )
 8 │ 
 9 █ func main() {
10 │ 	fmt.Println("Starting application...")
11 │ 	
12 │ 	if err := run(); err != nil {
13 │ 		log.Fatalf("Application failed: %v", err)
14 │ 	}
15 │ 	
16 │ 	fmt.Println("Application finished")
17 │ }
  ⋮
```

##### Example 2: Find error handling patterns

```bash
searchast -filename main.go -pattern "err.*nil"
```

**Output:**
```
  ⋮
 9 │ func main() {
10 │ 	fmt.Println("Starting application...")
11 │ 	
12 █ 	if err := run(); err != nil {
13 │ 		log.Fatalf("Application failed: %v", err)
14 │ 	}
15 │ 	
16 │ 	fmt.Println("Application finished")
17 │ }
18 │ 
19 │ func run() error {
20 │ 	// Process some data
21 │ 	data := []string{"item1", "item2", "item3"}
22 │ 	
23 │ 	for i, item := range data {
24 █ 		if err := processItem(i, item); err != nil {
25 │ 			return fmt.Errorf("failed to process item %d: %w", i, err)
26 │ 		}
27 │ 	}
28 │ 	
29 │ 	return saveResults("output.txt")
30 │ }
  ⋮
```

##### Example 3: Custom formatting

```bash
searchast -filename main.go -pattern "func.*\\(" -highlight-symbol ">>>" -context-symbol " | "
```

**Output:**
```
  ⋮
 3  |  import (
 4  | 	"fmt"
 5  | 	"log"
 6  | 	"os"
 7  |  )
 8  | 
 9 >>> func main() {
10  | 	fmt.Println("Starting application...")
11  |  	
12  | 	if err := run(); err != nil {
13  | 		log.Fatalf("Application failed: %v", err)
14  | 	}
15  |  	
16  | 	fmt.Println("Application finished")
17  |  }
  ⋮
```

##### Example 4: Find specific function calls

```bash
searchast -filename main.go -pattern "fmt\\.Print"
```

**Output:**
```
  ⋮
 3 │ import (
 4 │ 	"fmt"
 5 │ 	"log"
 6 │ 	"os"
 7 │ )
 8 │ 
 9 │ func main() {
10 █ 	fmt.Println("Starting application...")
11 │ 	
12 │ 	if err := run(); err != nil {
13 │ 		log.Fatalf("Application failed: %v", err)
14 │ 	}
15 │ 	
16 █ 	fmt.Println("Application finished")
17 │ }
18 │ 
19 │ func run() error {
20 │ 	// Process some data
21 │ 	data := []string{"item1", "item2", "item3"}
22 │ 	
23 │ 	for i, item := range data {
24 │ 		if err := processItem(i, item); err != nil {
25 │ 			return fmt.Errorf("failed to process item %d: %w", i, err)
26 │ 		}
27 │ 	}
28 │ 	
29 │ 	return saveResults("output.txt")
30 │ }
31 │ 
32 │ func processItem(index int, item string) error {
33 █ 	fmt.Printf("Processing item %d: %s\n", index, item)
34 │ 	
35 │ 	if item == "" {
36 │ 		return fmt.Errorf("empty item at index %d", index)
37 │ 	}
38 │ 	
39 │ 	return nil
40 │ }
  ⋮
```

### Package Usage

```go
package main

import (
    "context"
    "fmt"
    "os"
    "regexp"
    
    "github.com/andersonjoseph/searchast"
)

func main() {
    // Open a file
    file, err := os.Open("example.go")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    // Create a source tree
    sourceTree, err := searchast.NewSourceTree(context.Background(), file, "example.go")
    if err != nil {
        panic(err)
    }

    // Search for a pattern
    linesOfInterest, err := sourceTree.Search("func.*main")
    if err != nil {
        panic(err)
    }

    // Add context
    linesToShow := searchast.NewContextBuilder().AddContext(sourceTree, linesOfInterest)

    // Format output
    formatter := searchast.NewTextFormatter()
    output := formatter.Format(sourceTree.Lines(), linesToShow, linesOfInterest)
    fmt.Print(output)
}
```

### Advanced Usage

#### Custom Context Builder

```go
// Create a context builder with custom options
contextBuilder := searchast.NewContextBuilder(
    searchast.WithSurroundingLines(5),      // Show 5 lines before/after matches
    searchast.WithParentContext(false),    // Don't include parent context
    searchast.WithCloseScopeGaps(true),    // Close gaps within scopes
    searchast.WithChildLines(2),           // Show 2 lines of child context
)

linesToShow := contextBuilder.AddContext(sourceTree, linesOfInterest)
```

#### Custom Formatter

```go
// Create a formatter with custom symbols
formatter := searchast.NewTextFormatter(
    searchast.WithHighlightSymbol(">>>"),
    searchast.WithContextSymbol(" | "),
    searchast.WithGapSymbol("..."),
    searchast.WithLineNumbers(true),
    searchast.WithSpacer("  "),
)

output := formatter.Format(sourceTree.Lines(), linesToShow, linesOfInterest)
```

## Inspiration

This project is heavily inspired by [Aider-AI/grep-ast](https://github.com/Aider-AI/grep-ast), which provides similar functionality for Python. This Go implementation aims to provide:

- Better performance through Go's concurrency
- Easy integration with Go toolchains
- Cross-platform deployment
- Type-safe API for Go developers
