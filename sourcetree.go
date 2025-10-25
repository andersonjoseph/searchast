// Package findctx parses a source code file to build a tree structure representing
// code scopes. It allows for searching specific patterns and understanding their
// context within the code's hierarchy.
package findctx

import (
	"context"
	"fmt"
	"iter"
	"os"
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/andersonjoseph/findctx/internal/language"
)

type lineNumber = uint32

// line represents a single line in the source file, including its text and scope information.
type line struct {
	text  string
	scope scope
}

// scope defines a block of code, linking it to a parent and tracking its start and end lines.
type scope struct {
	parent lineNumber
	start  lineNumber
	end    lineNumber
}

// size calculates the number of lines contained within a scope.
func (s scope) size() uint32 {
	return s.end - s.start
}

// children returns an iterator sequence for all line numbers within a scope.
func (s scope) children() iter.Seq[lineNumber] {
	return func(yield func(lineNumber) bool) {
		for currentChild := s.start; currentChild <= s.end; currentChild++ {
			if !yield(currentChild) {
				return
			}
		}
	}
}

// sourceTree holds the entire parsed source file, line by line, with scope annotations.
type sourceTree struct {
	lines []line
}

// NewSourceTree reads a source file, parses it,
// and constructs a new sourceTree.
func NewSourceTree(filename string) (*sourceTree, error) {
	sourceCode, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	parser := sitter.NewParser()

	parser.SetLanguage(language.FromFilename(filename).SitterLang)

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

// build recursively traverses the tree-sitter abstract syntax tree (AST)
// to populate the scope information for each line.
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

// Search finds all lines that match a given regular expression pattern and returns
// their line numbers.
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
