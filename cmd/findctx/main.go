package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/andersonjoseph/findctx"
)

func main() {
	var filename string
	var pattern string

	flag.StringVar(&filename, "filename", "", "Source code file to search (required)")
	flag.StringVar(&pattern, "pattern", "", "Search pattern to find (required)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "Example: %s -filename sourcetree.go -pattern 'AI\\\\?'\n", os.Args[0])
	}

	flag.Parse()

	if filename == "" || pattern == "" {
		flag.Usage()
		os.Exit(1)
	}

	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening source file '%s': %v", filename, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("Error closing file '%s': %v", filename, err)
		}
	}()

	sourceTree, err := findctx.NewSourceTree(context.Background(), f, filename)
	if err != nil {
		log.Fatalf("Error opening source file '%s': %v", filename, err)
	}

	linesOfInterest, err := sourceTree.Search(pattern)
	if err != nil {
		log.Fatalf("Error searching for pattern '%s': %v", pattern, err)
	}

	if len(linesOfInterest) == 0 {
		log.Fatalf("No matches found")
	}

	linesToShow := findctx.NewContextBuilder().AddContext(sourceTree, linesOfInterest)
	formatter := findctx.NewTextFormatter()
	output := formatter.Format(sourceTree.Lines(), linesToShow, linesOfInterest)
	fmt.Print(output)
}
