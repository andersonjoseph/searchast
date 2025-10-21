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

type sourceHandler struct {
	lines []lineInfo
	linesToShow Set[lineNumber]
	seenParents Set[lineNumber]
}

// what is this function doing? AI?
func (sc *sourceHandler) walkSourceTree(node *sitter.Node) {
	if !node.IsNamed() {
		for i := range node.ChildCount() {
			sc.walkSourceTree(node.Child(int(i)))
		}
		return
	}

	startLine := node.StartPoint().Row
	endLine := node.EndPoint().Row

	nodeSize := endLine - startLine 

	if nodeSize > 0 && (sc.lines[startLine].scope.Size() == 0 || nodeSize > sc.lines[startLine].scope.Size()) {
		sc.lines[startLine].scope.startLine = startLine
		sc.lines[startLine].scope.endLine = endLine
	}

	childCount := int(node.ChildCount())
	for i := range childCount {
		child := node.Child(i)
		childLine := child.StartPoint().Row 

		if startLine != childLine {
			if sc.lines[childLine].parentLine == 0 {
				sc.lines[childLine].parentLine = startLine
			}
		}

		sc.walkSourceTree(child)
	}
}

func (sc sourceHandler) findLines(pattern string) (Set[lineNumber], error) {
	found := make(Set[lineNumber])

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	for i, line := range sc.lines {
		if re.MatchString(line.text) {
			found.Add(lineNumber(i))
		}
	}

	return found, nil
}

func (sc *sourceHandler) addContext(linesOfInterest Set[lineNumber]) {
	// add some surrounding lines as lines of interest
	for line := range linesOfInterest {
		gap := lineNumber(1)//FIXME: this should be a parameter

		for currentLine := line - gap; currentLine <= line + gap; currentLine++ {
			if currentLine >= lineNumber(len(sc.lines)) {
				continue
			}
			sc.linesToShow.Add(currentLine)
		}
	}

	linesSoFar := sc.linesToShow.ToSlice()

	for _, line := range linesSoFar {
		lineInfo := sc.lines[line]
		if lineInfo.scope.Size() > 0 { // if a full scope is part of the lines, add the scope
			for currentLine := lineInfo.scope.startLine; currentLine <= lineInfo.scope.endLine; currentLine++ {
				sc.linesToShow.Add(currentLine)
			}
		}
	}

	linesSoFar = sc.linesToShow.ToSlice()
	for _, line := range linesSoFar {
		sc.addParentContext(line)
	}

	linesSoFar = sc.linesToShow.ToSlice()
	for _, line := range linesSoFar {
		sc.addChildContext(line)
	}

	sc.closeGaps()
}

func (sc *sourceHandler) closeGaps() {
	sortedLines := sc.linesToShow.ToSlice()
	slices.Sort(sortedLines)

	gapToClose := lineNumber(3)//FIXME: this should be a parameter too
	for i := range sortedLines {
		if i == len(sortedLines)-1 {
			continue
		}

		diff := sortedLines[i+1] - sortedLines[i]
		if diff <= lineNumber(gapToClose) {
			for currentLine := sortedLines[i] + 1; currentLine <= sortedLines[i+1]; currentLine++ {
				sc.linesToShow.Add(currentLine)
			}
		}
	}
}

func (sc *sourceHandler) addParentContext(line lineNumber) {
	parentLine := sc.lines[line].parentLine
	if sc.seenParents.Has(parentLine) {
		return
	}
	sc.seenParents.Add(parentLine)

	parentLineInfo := sc.lines[parentLine]

	sc.linesToShow.Add(parentLineInfo.scope.startLine)
	sc.linesToShow.Add(parentLineInfo.scope.endLine)

	if parentLine == line {
		return
	}

	sc.addParentContext(parentLine)
}

func (sc *sourceHandler) addChildContext(line lineNumber) {
	if line == 0 || sc.lines[line].scope.Size() == 0 {
		return
	}

	lineInfo := sc.lines[line] 

	 // FIXME: most of these parameters should be configurable
	limitLine := min(lineNumber(3) + lineInfo.scope.startLine, lineInfo.scope.endLine)
	threshold := lineInfo.scope.startLine + ((lineInfo.scope.Size() * 70) / 100)

	if limitLine > threshold {
		limitLine = lineInfo.scope.endLine
	}

	for currentLine := lineInfo.scope.startLine; currentLine <= limitLine; currentLine++ {
		sc.linesToShow.Add(currentLine)
		sc.linesToShow.Add(sc.lines[currentLine].scope.endLine)
	}
}

func main() {
	source, err := os.ReadFile("./test/test.go")
	if err != nil {
		panic(err)
	}
	sourceLines := strings.Split(string(source), "\n")

	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		panic(err)
	}

	root := tree.RootNode()

	lines := make([]lineInfo, len(sourceLines))
	for i := range lines {
		lines[i].text = sourceLines[i]
	}

	handler := sourceHandler {
		lines: lines,
		linesToShow: NewSet[lineNumber](),
		seenParents: NewSet[lineNumber](),
	}

	handler.walkSourceTree(root)

	linesOfInterest, err := handler.findLines("j\\+\\+")
	if err != nil {
		panic(err)
	}

	handler.addContext(linesOfInterest)
	output := strings.Builder{}
	isGapPrinted := false

	for i, line := range lines {
		if !handler.linesToShow.Has(lineNumber(i)){
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
