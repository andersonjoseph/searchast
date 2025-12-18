package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/andersonjoseph/searchast"
)

func main() {
	var (
		filename  string
		colorFlag string
	)

	flag.StringVar(&filename, "filename", "", "Source code file to search (required)")
	flag.StringVar(&colorFlag, "color", "auto", "Color output: auto, always, never")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "Example: %s -filename sourcetree.go\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Color options: auto (detect terminal), always, never\n")
	}

	flag.Parse()

	if filename == "" {
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

	sourceTree, err := searchast.NewSourceTree(context.Background(), f, filename)
	if err != nil {
		log.Fatalf("Error opening source file '%s': %v", filename, err)
	}

	linesOfInterest := sourceTree.TopLevel()
	if len(linesOfInterest) == 0 {
		log.Fatalf("No matches found")
	}

	overivewContextBuilder := searchast.NewContextBuilder(
		searchast.WithSurroundingLines(2),
		searchast.WithParentContext(false),
		searchast.WithCloseScopeGaps(false),
		searchast.WithExpandChildScopes(false),
		searchast.WithChildLines(3),
	)

	linesToShow := overivewContextBuilder.AddContext(sourceTree, linesOfInterest)

	var enableColors bool
	switch colorFlag {
	case "always":
		enableColors = true
	case "never":
		enableColors = false
	case "auto":
		fallthrough
	default:
		enableColors = true
	}

	formatter := searchast.NewTextFormatter(searchast.WithColors(enableColors))
	output := formatter.Format(sourceTree.Lines(), linesToShow, linesOfInterest)
	fmt.Print(output)
}
