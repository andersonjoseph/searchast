package main

import (
	"context"
	"fmt"
	"iter"
	"os"
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

type lineNumber = uint32

type line struct {
	text  string
	scope scope
}

type scope struct {
	parent lineNumber
	start  lineNumber
	end    lineNumber
}

func (s scope) Size() uint32 {
	return s.end - s.start
}

func (s scope) Children() iter.Seq[lineNumber] {
	return func(yield func(lineNumber) bool) {
		for currentChild := s.start; currentChild <= s.end; currentChild++ {
			if !yield(currentChild) {
				return
			}
		}
	}
}

type SourceTree struct {
	lines []line
}

func NewSourceTree(filename string) (*SourceTree, error) {
	sourceCode, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	parser := sitter.NewParser()

	//TODO: we need to find a way to get the language from the file extension
	parser.SetLanguage(golang.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, sourceCode)
	if err != nil {
		panic(err)
	}
	root := tree.RootNode()

	sourceLines := strings.Split(string(sourceCode), "\n")

	lines := make([]line, len(sourceLines))
	for i := range lines {
		lines[i].text = sourceLines[i]
	}

	st := &SourceTree{
		lines: lines,
	}

	st.build(root)

	return st, nil
}

func (st *SourceTree) build(node *sitter.Node) {
	if !node.IsNamed() {
		for i := range node.ChildCount() {
			st.build(node.Child(int(i)))
		}
		return
	}

	startLine := node.StartPoint().Row
	endLine := node.EndPoint().Row

	nodeSize := endLine - startLine

	// Explain the purpose of this function AI?
	if nodeSize > 0 && (st.lines[startLine].scope.Size() == 0 || nodeSize > st.lines[startLine].scope.Size()) {
		st.lines[startLine].scope.start = startLine
		st.lines[startLine].scope.end = endLine
	}

	childCount := int(node.ChildCount())
	for i := range childCount {
		child := node.Child(i)
		childLine := child.StartPoint().Row

		if startLine != childLine {
			if st.lines[childLine].scope.parent == 0 {
				st.lines[childLine].scope.parent = startLine
			}
		}

		st.build(child)
	}
}

func formatOutput(st *SourceTree, linesToShow Set[lineNumber], linesOfInterest Set[lineNumber]) string {
	output := strings.Builder{}
	isGapPrinted := false

	for i, line := range st.lines {
		if !linesToShow.Has(lineNumber(i)) {
			if !isGapPrinted {
				output.WriteString("⋮\n")
				isGapPrinted = true
			}

			continue
		}

		isGapPrinted = false
		var spacer string
		if linesOfInterest.Has(lineNumber(i)) {
			spacer = "█"
		} else {
			spacer = "│"
		}

		output.WriteString(fmt.Sprintf("%s %s\n", spacer, line.text))
	}

	return output.String()
}

func SearchInTree(pattern string, st *SourceTree) (Set[lineNumber], error) {
	linesOfInterest := NewSet[lineNumber]()
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex pattern: %w", err)
	}

	for i, line := range st.lines {
		if re.MatchString(line.text) {
			linesOfInterest.Add(lineNumber(i))
		}
	}

	return linesOfInterest, nil
}

func main() {
	sourceTree, err := NewSourceTree("./main.go")
	if err != nil {
		panic(err)
	}

	linesOfInterest, err := SearchInTree("AI\\?", sourceTree)
	if err != nil {
		panic(err)
	}

	linesToShow := NewContextBuilder().AddContext(sourceTree, linesOfInterest)

	print(formatOutput(sourceTree, linesToShow, linesOfInterest))
}
