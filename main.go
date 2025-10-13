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
	childrens map[lineNumber]struct{}
	scope scopeInfo
}

type scopeInfo struct {
	startLine lineNumber
	endLine lineNumber
	size uint32
}

type sourceHandler struct {
	lines []lineInfo
	linesToShow map[lineNumber]struct{}
	seenParents map[lineNumber]struct{}
}

func (sc *sourceHandler) walkSourceTree(node *sitter.Node, parentLine lineNumber) {
	startLine := node.StartPoint().Row
	endLine := node.EndPoint().Row

	for currentLine := startLine; currentLine <= endLine; currentLine++ {
		if parentLine != currentLine && (sc.lines[currentLine].parentLine == 0 || parentLine > sc.lines[currentLine].parentLine) {
			sc.lines[currentLine].parentLine = parentLine
		}
	}

	if parentLine > 0 {
		if sc.lines[parentLine].childrens == nil {
			sc.lines[parentLine].childrens = make(map[lineNumber]struct{})
		}
		sc.lines[parentLine].childrens[startLine] = struct{}{}
	}

	nodeSize := endLine - startLine 

	if nodeSize > 0 {
		if sc.lines[startLine].scope.size == 0 || nodeSize > sc.lines[startLine].scope.size {
			sc.lines[startLine].scope = scopeInfo {
				startLine: startLine,
				endLine: endLine,
				size: nodeSize,
			}
		}
	}

	for i := range node.ChildCount() {
		sc.walkSourceTree(node.Child(int(i)), startLine)
	}
}

func (sc sourceHandler) findLines(pattern string) ([]lineNumber, error) {
	found := make([]lineNumber, 0)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	for i, line := range sc.lines {
		if re.MatchString(line.text) {
			found = append(found, lineNumber(i))
		}
	}

	return found, nil
}

func (sc *sourceHandler) addContext(lines []lineNumber) []lineNumber {
	// add some surrounding lines as lines of interest
	for _, line := range lines {
		gap := lineNumber(1)

		for currentLine := line - gap; currentLine <= line + gap; currentLine++ {
			if currentLine >= lineNumber(len(sc.lines)) {
				continue
			}
			lines = append(lines, currentLine)
		}
	}

	// if a full scope is part of the lines, add the scope
	for _, line := range lines {
		lineInfo := sc.lines[line]
		if lineInfo.scope.size == 0 {
			continue
		}
		for currentLine := lineInfo.scope.startLine; currentLine <= lineInfo.scope.endLine; currentLine++ {
			sc.linesToShow[currentLine] = struct{}{}
		}
	}

	sc.linesToShow[0] = struct{}{}

	bottomLine := lineNumber(len(sc.lines)-2)
	sc.linesToShow[bottomLine] = struct{}{}
	sc.addParentContext(bottomLine)

	for _, line := range lines {
		sc.addParentContext(line)
		sc.addChildContext(line)
	}

	sortedLines := make([]lineNumber, 0, len(sc.linesToShow))
	for ln := range sc.linesToShow {
		sortedLines = append(sortedLines, ln)
	}
	slices.Sort(sortedLines)

	return sortedLines
}

func (sc *sourceHandler) addParentContext(line lineNumber) {
	parentLine := sc.lines[line].parentLine
	if _, ok := sc.seenParents[parentLine]; ok {
		return
	}
	sc.seenParents[parentLine] = struct{}{}

	parentLineInfo := sc.lines[parentLine]

	sc.linesToShow[parentLineInfo.scope.startLine] = struct{}{}
	sc.linesToShow[parentLineInfo.scope.endLine] = struct{}{}

	if parentLine == line {
		return
	}

	sc.addParentContext(parentLine)
}

func (sc *sourceHandler) addChildContext(line lineNumber) {
	lineInfo := sc.lines[line]

	if lineInfo.scope.size < 5 {
		for currentLine := lineInfo.scope.startLine; currentLine <= lineInfo.scope.endLine; currentLine++ {
			sc.linesToShow[currentLine] = struct{}{}
		}
		return
	}

	maxToShow := int(max(min(float64(lineInfo.scope.size)* 0.10, 25), 5))
	currentlyShowing := len(sc.linesToShow)

	for childLine := range lineInfo.childrens {
		if len(sc.linesToShow) > currentlyShowing + maxToShow {
			break
		}

		sc.addParentContext(childLine)
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
		linesToShow: make(map[lineNumber]struct{}),
		seenParents: make(map[lineNumber]struct{}),
	}

	handler.walkSourceTree(root, 0)

	linesOfInterest, err := handler.findLines("AI?")
	if err != nil {
		panic(err)
	}
	linesToShow := handler.addContext(linesOfInterest)

	for _, line := range linesToShow {
		fmt.Printf("%d %v\n", line+1, handler.lines[line].text)
	}
}
