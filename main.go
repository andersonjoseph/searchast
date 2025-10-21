package main

import (
	"context"
	"fmt"
	"iter"
	"os"
	"regexp"
	"slices"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

type lineNumber = uint32

type line struct {
	text string
	scope scope
}

type scope struct {
	parent lineNumber
	start lineNumber
	end lineNumber
}

func (s scope) Size() uint32 {
	return s.end - s.start
}

func (s scope) Children()iter.Seq[lineNumber] {
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
	linesToShow Set[lineNumber]
	seenParents Set[lineNumber]
}

func NewSourceTree(sourceCode string) SourceTree {
	sourceLines := strings.Split(string(sourceCode), "\n")

	lines := make([]line, len(sourceLines))
	for i := range lines {
		lines[i].text = sourceLines[i]
	}

	st := SourceTree {
		lines: lines,
		linesToShow: NewSet[lineNumber](),
		seenParents: NewSet[lineNumber](),
	}

	return st
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

func (st SourceTree) findLines(pattern string) (Set[lineNumber], error) {
	found := make(Set[lineNumber])

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	for i, line := range st.lines {
		if re.MatchString(line.text) {
			found.Add(lineNumber(i))
		}
	}

	return found, nil
}

func (st *SourceTree) addContext(linesOfInterest Set[lineNumber]) {
	// add some surrounding lines as lines of interest
	for line := range linesOfInterest {
		gap := lineNumber(3)//FIXME: this should be a parameter

		for currentLine := line - gap; currentLine <= line + gap; currentLine++ {
			if currentLine >= lineNumber(len(st.lines)) {
				continue
			}
			st.linesToShow.Add(currentLine)
		}
	}

	linesSoFar := st.linesToShow.ToSlice()

	for _, line := range linesSoFar {
		lineInfo := st.lines[line]
		for childLine := range lineInfo.scope.Children() {
			st.linesToShow.Add(childLine)
		}
	}

	linesSoFar = st.linesToShow.ToSlice()
	for _, line := range linesSoFar {
		st.addParentContext(line)
	}

	linesSoFar = st.linesToShow.ToSlice()
	for _, line := range linesSoFar {
		st.addChildContext(line)
	}

	st.closeGaps()
}

func (st *SourceTree) closeGaps() {
	sortedLines := st.linesToShow.ToSlice()
	slices.Sort(sortedLines)

	gapToClose := lineNumber(3)//FIXME: this should be a parameter too
	for i := range sortedLines {
		if i == len(sortedLines)-1 {
			continue
		}

		diff := sortedLines[i+1] - sortedLines[i]
		if diff <= lineNumber(gapToClose) {
			for currentLine := sortedLines[i] + 1; currentLine <= sortedLines[i+1]; currentLine++ {
				st.linesToShow.Add(currentLine)
			}
		}
	}
}

func (st *SourceTree) addParentContext(line lineNumber) {
	parentLine := st.lines[line].scope.parent
	if st.seenParents.Has(parentLine) {
		return
	}
	st.seenParents.Add(parentLine)

	parentLineInfo := st.lines[parentLine]

	st.linesToShow.Add(parentLineInfo.scope.start)
	st.linesToShow.Add(parentLineInfo.scope.end)

	if parentLine == line {
		return
	}

	st.addParentContext(parentLine)
}

func (st *SourceTree) addChildContext(line lineNumber) {
	if line == 0 || st.lines[line].scope.Size() == 0 {
		return
	}

	lineInfo := st.lines[line] 

	 // FIXME: most of these parameters should be configurable
	limitLine := min(lineNumber(3) + lineInfo.scope.start, lineInfo.scope.end)
	threshold := lineInfo.scope.start + ((lineInfo.scope.Size() * 70) / 100)

	if limitLine > threshold {
		limitLine = lineInfo.scope.end
	}

	for currentLine := lineInfo.scope.start; currentLine <= limitLine; currentLine++ {
		st.linesToShow.Add(currentLine)
		st.linesToShow.Add(st.lines[currentLine].scope.end)
	}
}

func main() {
	source, err := os.ReadFile("./main.go")
	if err != nil {
		panic(err)
	}

	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		panic(err)
	}
	root := tree.RootNode()

	sourceTree := NewSourceTree(string(source))
	sourceTree.build(root)

	linesOfInterest, err := sourceTree.findLines("AI\\?")
	if err != nil {
		panic(err)
	}

	sourceTree.addContext(linesOfInterest)
	output := strings.Builder{}
	isGapPrinted := false

	for i, line := range sourceTree.lines {
		if !sourceTree.linesToShow.Has(lineNumber(i)){
			if !isGapPrinted {
				output.WriteString("⋮\n")
				isGapPrinted = true
			}

			continue
		}

		isGapPrinted = false
		var spacer string
		if linesOfInterest.Has(lineNumber(i)){
			spacer = "█"
		} else {
			spacer = "│"
		}

		output.WriteString(fmt.Sprintf("%s %s\n",  spacer, line.text))
	}

	print(output.String())
}
