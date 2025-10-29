// Package findctx parses a source code file to build a tree structure representing
// code scopes. It allows for searching specific patterns and understanding their
// context within the code's hierarchy.
package findctx

import (
	"context"
	"fmt"
	"io"
	"iter"
	"regexp"
	"strings"

	"github.com/andersonjoseph/findctx/language"
	sitter "github.com/smacker/go-tree-sitter"
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

type sourceTree struct {
	lines []line
}

// NewSourceTree costructs a new sourceTree from a reader and filename.
// the filename is used to determine the programming language.
func NewSourceTree(ctx context.Context, r io.Reader, filename string) (*sourceTree, error) {
	sourceCode, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	parser := sitter.NewParser()

	parser.SetLanguage(language.FromFilename(filename).SitterLang)

	tree, err := parser.ParseCtx(ctx, nil, sourceCode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filename, err)
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
	childCount := int(node.ChildCount())

	if !node.IsNamed() { // If the node is not named, it is a leaf node and has no scope information.
		for i := range childCount {
			st.build(node.Child(i))
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
func (st *sourceTree) Search(pattern string) (Set[lineNumber], error) {
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
