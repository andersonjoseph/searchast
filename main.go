package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

type lineNumber = uint32

type lineInfo struct {
	text string
	parentLine lineNumber
	scope scopeInfo
}

type scopeInfo struct {
	startLine lineNumber
	endLine lineNumber
}

func (s scopeInfo) Size() uint32 {
	return s.endLine - s.startLine
}

type SourceTree struct {
	lines []lineInfo
	linesToShow Set[lineNumber]
	seenParents Set[lineNumber]
}

func NewSourceTree(sourceCode string) SourceTree {
	sourceLines := strings.Split(string(sourceCode), "\n")

	lines := make([]lineInfo, len(sourceLines))
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
		st.lines[startLine].scope.startLine = startLine
		st.lines[startLine].scope.endLine = endLine
	}

	childCount := int(node.ChildCount())
	for i := range childCount {
		child := node.Child(i)
		childLine := child.StartPoint().Row 

		if startLine != childLine {
			if st.lines[childLine].parentLine == 0 {
				st.lines[childLine].parentLine = startLine
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
		if lineInfo.scope.Size() > 0 { // if a full scope is part of the lines, add the scope
			for currentLine := lineInfo.scope.startLine; currentLine <= lineInfo.scope.endLine; currentLine++ {
				st.linesToShow.Add(currentLine)
			}
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
	parentLine := st.lines[line].parentLine
	if st.seenParents.Has(parentLine) {
		return
	}
	st.seenParents.Add(parentLine)

	parentLineInfo := st.lines[parentLine]

	st.linesToShow.Add(parentLineInfo.scope.startLine)
	st.linesToShow.Add(parentLineInfo.scope.endLine)

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
	limitLine := min(lineNumber(3) + lineInfo.scope.startLine, lineInfo.scope.endLine)
	threshold := lineInfo.scope.startLine + ((lineInfo.scope.Size() * 70) / 100)

	if limitLine > threshold {
		limitLine = lineInfo.scope.endLine
	}

	for currentLine := lineInfo.scope.startLine; currentLine <= limitLine; currentLine++ {
		st.linesToShow.Add(currentLine)
		st.linesToShow.Add(st.lines[currentLine].scope.endLine)
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
