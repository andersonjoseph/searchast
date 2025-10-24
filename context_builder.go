package main

import (
	"slices"
)

type ContextBuilder struct {
	GapToShow        lineNumber
	GapToClose       lineNumber
	ParentContext    bool
	ChildContext     bool
	SurroundingLines bool
	seenParents      Set[lineNumber]
}

func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{
		GapToShow:        3,
		GapToClose:       3,
		ParentContext:    true,
		ChildContext:     true,
		SurroundingLines: true,
		seenParents:      NewSet[lineNumber](),
	}
}

func (cb *ContextBuilder) AddContext(st *SourceTree) {
	// add some surrounding lines as lines of interest
	for line := range st.linesOfInterest {
		gap := cb.GapToShow

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
		cb.addParentContext(st, line)
	}

	linesSoFar = st.linesToShow.ToSlice()
	for _, line := range linesSoFar {
		cb.addChildContext(st, line)
	}

	cb.closeGaps(st)
}

func (cb *ContextBuilder) addParentContext(st *SourceTree, line lineNumber) {
	parentLine := st.lines[line].scope.parent
	if cb.seenParents.Has(parentLine) {
		return
	}
	cb.seenParents.Add(parentLine)

	parentLineInfo := st.lines[parentLine]

	st.linesToShow.Add(parentLineInfo.scope.start)
	st.linesToShow.Add(parentLineInfo.scope.end)

	if parentLine == line {
		return
	}

	cb.addParentContext(st, parentLine)
}

func (cb *ContextBuilder) addChildContext(st *SourceTree, line lineNumber) {
	if line == 0 || st.lines[line].scope.Size() == 0 {
		return
	}

	lineInfo := st.lines[line]

	limitLine := min(cb.GapToShow + lineInfo.scope.start, lineInfo.scope.end)
	threshold := lineInfo.scope.start + ((lineInfo.scope.Size() * 70) / 100)

	if limitLine > threshold {
		limitLine = lineInfo.scope.end
	}

	for currentLine := lineInfo.scope.start; currentLine <= limitLine; currentLine++ {
		st.linesToShow.Add(currentLine)
		st.linesToShow.Add(st.lines[currentLine].scope.end)
	}
}

func (cb *ContextBuilder) closeGaps(st *SourceTree) {
	sortedLines := st.linesToShow.ToSlice()
	slices.Sort(sortedLines)

	for i := range sortedLines {
		if i == len(sortedLines)-1 {
			continue
		}

		diff := sortedLines[i+1] - sortedLines[i]
		if diff <= cb.GapToClose {
			for currentLine := sortedLines[i] + 1; currentLine <= sortedLines[i+1]; currentLine++ {
				st.linesToShow.Add(currentLine)
			}
		}
	}
}