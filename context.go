package findctx

import (
	"slices"
)

type ContextBuilder struct {
	GapToShow        lineNumber
	GapToClose       lineNumber
	ParentContext    bool
	ChildContext     bool
	SurroundingLines bool

	seenParents set[lineNumber]
	linesToShow set[lineNumber]
}

func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{
		GapToShow:        3,
		GapToClose:       3,
		ParentContext:    true,
		ChildContext:     true,
		SurroundingLines: true,
		seenParents:      newSet[lineNumber](),
		linesToShow:      newSet[lineNumber](),
	}
}

func (cb *ContextBuilder) AddContext(st *sourceTree, linesOfInterest set[lineNumber]) set[lineNumber] {
	defer func() {
		cb.linesToShow.clear()
		cb.seenParents.clear()
	}()

	cb.addSurroundingLines(st, linesOfInterest)
	linesSoFar := cb.linesToShow.toSlice()

	for _, line := range linesSoFar {
		lineInfo := st.lines[line]
		// if the linesOfInterest and their surrounding lines contains scopes, add them
		for childLine := range lineInfo.scope.children() {
			cb.linesToShow.add(childLine)
		}
	}

	linesSoFar = cb.linesToShow.toSlice()
	for _, line := range linesSoFar {
		cb.addParentContext(st, line)
	}

	linesSoFar = cb.linesToShow.toSlice()
	for _, line := range linesSoFar {
		cb.addChildContext(st, line)
	}

	cb.closeGaps()

	return cb.linesToShow
}

func (cb *ContextBuilder) addSurroundingLines(st *sourceTree, linesOfInterest set[lineNumber]) {
	gap := cb.GapToShow

	for line := range linesOfInterest {
		for currentLine := line - gap; currentLine <= line+gap; currentLine++ {
			if currentLine >= lineNumber(len(st.lines)) {
				break
			}
			cb.linesToShow.add(currentLine)
		}
	}
}

func (cb *ContextBuilder) addParentContext(st *sourceTree, line lineNumber) {
	parentLine := st.lines[line].scope.parent
	if cb.seenParents.has(parentLine) {
		return
	}
	cb.seenParents.add(parentLine)

	parentLineInfo := st.lines[parentLine]

	cb.linesToShow.add(parentLineInfo.scope.start)
	cb.linesToShow.add(parentLineInfo.scope.end)

	if parentLine == line {
		return
	}

	cb.addParentContext(st, parentLine)
}

func (cb *ContextBuilder) addChildContext(st *sourceTree, line lineNumber) {
	if line == 0 || st.lines[line].scope.size() == 0 {
		return
	}

	lineInfo := st.lines[line]

	limitLine := min(cb.GapToShow+lineInfo.scope.start, lineInfo.scope.end)
	threshold := lineInfo.scope.start + ((lineInfo.scope.size() * 70) / 100)

	if limitLine > threshold {
		limitLine = lineInfo.scope.end
	}

	for currentLine := lineInfo.scope.start; currentLine <= limitLine; currentLine++ {
		cb.linesToShow.add(currentLine)
		cb.linesToShow.add(st.lines[currentLine].scope.end)
	}
}

func (cb *ContextBuilder) closeGaps() {
	sortedLines := cb.linesToShow.toSlice()
	slices.Sort(sortedLines)

	for i := range sortedLines {
		if i == len(sortedLines)-1 {
			continue
		}

		diff := sortedLines[i+1] - sortedLines[i]
		if diff <= cb.GapToClose {
			for currentLine := sortedLines[i] + 1; currentLine <= sortedLines[i+1]; currentLine++ {
				cb.linesToShow.add(currentLine)
			}
		}
	}
}
