package findctx

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

func (s scope) size() uint32 {
	return s.end - s.start
}

func (s scope) children() iter.Seq[lineNumber] {
	return func(yield func(lineNumber) bool) {
		for currentChild := s.start; currentChild <= s.end; currentChild++ {
			if !yield(currentChild) {
				return
			}
		}
	}
}

type sourceTree struct {
	lines []line
}

func NewSourceTree(filename string) (*sourceTree, error) {
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

	st := &sourceTree{
		lines: lines,
	}

	st.build(root)

	return st, nil
}

func (st *sourceTree) build(node *sitter.Node) {
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
	if nodeSize > 0 && (st.lines[startLine].scope.size() == 0 || nodeSize > st.lines[startLine].scope.size()) {
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

func (st *sourceTree) Search(pattern string) (set[lineNumber], error) {
	linesOfInterest := newSet[lineNumber]()
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex pattern: %w", err)
	}

	for i, line := range st.lines {
		if re.MatchString(line.text) {
			linesOfInterest.add(lineNumber(i))
		}
	}

	return linesOfInterest, nil
}
