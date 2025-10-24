package main

import "github.com/andersonjoseph/findctx"

func main() {
	sourceTree, err := findctx.NewSourceTree("./sourcetree.go")
	if err != nil {
		panic(err)
	}

	linesOfInterest, err := sourceTree.Search("AI\\?")
	if err != nil {
		panic(err)
	}
	linesToShow := findctx.NewContextBuilder().AddContext(sourceTree, linesOfInterest)

	print(findctx.FormatOutput(sourceTree, linesToShow, linesOfInterest))
}
