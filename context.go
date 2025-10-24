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

func NewContextBuilder(opts ...Option) *ContextBuilder {
	cb := &ContextBuilder{
		GapToShow:        3,
		GapToClose:       3,
		ParentContext:    true,
		ChildContext:     true,
		SurroundingLines: true,
		seenParents:      newSet[lineNumber](),
		linesToShow:      newSet[lineNumber](),
	}

	for _, opt := range opts {
		opt(cb)
	}

	return cb
}

func (cb *ContextBuilder) AddContext(st *sourceTree, linesOfInterest set[lineNumber]) set[lineNumber] {
	defer func() {
		cb.linesToShow.clear()
		cb.seenParents.clear()
	}()

	// Add initial lines of interest, with or without surrounding lines
	if cb.SurroundingLines {
		cb.addSurroundingLines(st, linesOfInterest)
	} else {
		for line := range linesOfInterest {
			cb.linesToShow.add(line)
		}
	}

	linesSoFar := cb.linesToShow.toSlice()
	for _, line := range linesSoFar {
		lineInfo := st.lines[line]
		// if the linesOfInterest and their surrounding lines contains scopes, add them
		for childLine := range lineInfo.scope.children() {
			cb.linesToShow.add(childLine)
		}
	}

	if cb.ParentContext {
		linesSoFar = cb.linesToShow.toSlice()
		for _, line := range linesSoFar {
			cb.addParentContext(st, line)
		}
	}

	if cb.ChildContext {
		linesSoFar = cb.linesToShow.toSlice()
		for _, line := range linesSoFar {
			cb.addChildContext(st, line)
		}
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

type Option func(*ContextBuilder)

// WithGapToShow sets the number of lines to show around a line of interest.
func WithGapToShow(gap lineNumber) Option {
	return func(cb *ContextBuilder) {
		cb.GapToShow = gap
	}
}

// WithGapToClose sets the maximum gap between lines to fill in.
func WithGapToClose(gap lineNumber) Option {
	return func(cb *ContextBuilder) {
		cb.GapToClose = gap
	}
}

// WithParentContext enables or disables the inclusion of parent context.
func WithParentContext(enabled bool) Option {
	return func(cb *ContextBuilder) {
		cb.ParentContext = enabled
	}
}

// WithChildContext enables or disables the inclusion of child context.
func WithChildContext(enabled bool) Option {
	return func(cb *ContextBuilder) {
		cb.ChildContext = enabled
	}
}

// WithSurroundingLines enables or disables the inclusion of surrounding lines.
func WithSurroundingLines(enabled bool) Option {
	return func(cb *ContextBuilder) {
		cb.SurroundingLines = enabled
	}
}
