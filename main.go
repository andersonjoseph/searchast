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
	size uint32
}

type sourceHandler struct {
	lines []lineInfo
	linesToShow Set[lineNumber]
	seenParents Set[lineNumber]
}

// what is this function doing? AI?
func (sc *sourceHandler) walkSourceTree(node *sitter.Node, parentLine lineNumber) {
	startLine := node.StartPoint().Row
	endLine := node.EndPoint().Row

	for currentLine := startLine; currentLine <= endLine; currentLine++ {
		if parentLine != currentLine && (sc.lines[currentLine].parentLine == 0 || parentLine > sc.lines[currentLine].parentLine) {
			sc.lines[currentLine].parentLine = parentLine
		}
	}

	nodeSize := endLine - startLine 

	if nodeSize > 0 && (sc.lines[startLine].scope.size == 0 || nodeSize > sc.lines[startLine].scope.size) {
		sc.lines[startLine].scope = scopeInfo {
			startLine: startLine,
			endLine: endLine,
			size: nodeSize,
		}
	}

	for i := range node.ChildCount() {
		sc.walkSourceTree(node.Child(int(i)), startLine)
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
		sc.linesToShow.Add(line)

		lineInfo := sc.lines[line]
		if lineInfo.scope.size > 0 { // if a full scope is part of the lines, add the scope
			for currentLine := lineInfo.scope.startLine; currentLine <= lineInfo.scope.endLine; currentLine++ {
				sc.linesToShow.Add(currentLine)
			}
		}
	}

	for line := range linesOfInterest {
		sc.addParentContext(line)
	}

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
	if sc.seenParents.Has(parentLine) || parentLine == 0 {
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

	handler.walkSourceTree(root, 0)

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
			spacer = "|"
		}

		output.WriteString(fmt.Sprintf("%d %s %s\n", i+1, spacer, line.text))
	}

	print(output.String())
}
