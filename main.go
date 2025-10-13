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
	blockInfo blockInfo
}

type blockInfo struct {
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

	for i := startLine; i <= endLine; i++ {
		if parentLine != i && (sc.lines[i].parentLine == 0 || parentLine > sc.lines[i].parentLine) {
			sc.lines[i].parentLine = parentLine
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
		if sc.lines[startLine].blockInfo.size == 0 || nodeSize > sc.lines[startLine].blockInfo.size {
			sc.lines[startLine].blockInfo = blockInfo {
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
	for _, line := range lines {
		if line == 0 {
			sc.linesToShow[0] = struct{}{}
			continue
		}

		gap := lineNumber(1)

		for i := line - gap; i <= line + gap; i++ {
			if i >= lineNumber(len(sc.lines)) {
				continue
			}
			sc.linesToShow[i] = struct{}{}
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

	sc.linesToShow[parentLineInfo.blockInfo.startLine] = struct{}{}
	sc.linesToShow[parentLineInfo.blockInfo.endLine] = struct{}{}

	if parentLine == line {
		return
	}

	sc.addParentContext(parentLine)
}

func (sc *sourceHandler) addChildContext(line lineNumber) {
	lineInfo := sc.lines[line]

	if lineInfo.blockInfo.size < 5 {
		for i := lineInfo.blockInfo.startLine; i <= lineInfo.blockInfo.endLine; i++ {
			sc.linesToShow[i] = struct{}{}
		}
		return
	}

	maxToShow := int(max(min(float64(lineInfo.blockInfo.size)* 0.10, 25), 5))
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

	linesOfInterest, err := handler.findLines("j\\+\\+")
	if err != nil {
		panic(err)
	}
	linesToShow := handler.addContext(linesOfInterest)

	for _, line := range linesToShow {
		fmt.Printf("%d %v\n", line+1, handler.lines[line].text)
	}
}
