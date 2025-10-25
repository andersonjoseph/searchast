package findctx

import (
	"github.com/andersonjoseph/findctx/internal"
	"slices"
)

// ContextBuilder constructs a set of line numbers to display based on an initial
// set of lines. It expands this set by including parent and child
// scopes, surrounding lines, and by closing small gaps to create a more
// readable and contextual output.
type ContextBuilder struct {
	// SurroundingLines specifies how many lines of context to show around a matched line.
	SurroundingLines lineNumber
	// GapToClose determines the maximum gap size between two lines that should be filled in.
	GapToClose lineNumber
	// ParentContext, if true, includes the start and end lines of parent scopes.
	ParentContext bool
	// ChildContext, if true, includes the beginning of any child scopes.
	ChildContext bool

	seenParents internal.Set[lineNumber]
	linesToShow internal.Set[lineNumber]
}

// NewContextBuilder creates a new ContextBuilder with default values, which can be
// customized by passing in functional options.
func NewContextBuilder(opts ...Option) *ContextBuilder {
	cb := &ContextBuilder{
		SurroundingLines: 3,
		GapToClose:       3,
		ParentContext:    true,
		ChildContext:     true,

		seenParents: internal.NewSet[lineNumber](),
		linesToShow: internal.NewSet[lineNumber](),
	}

	for _, opt := range opts {
		opt(cb)
	}

	return cb
}

// AddContext takes a sourceTree and a set of lines of interest and returns an
// expanded set of lines based on the builder's configuration. The builder's
// internal state is reset after each call
func (cb *ContextBuilder) AddContext(st *sourceTree, linesOfInterest internal.Set[lineNumber]) internal.Set[lineNumber] {
	defer func() {
		cb.linesToShow.Clear()
		cb.seenParents.Clear()
	}()

	cb.addSurroundingLines(st, linesOfInterest)

	// If the initial lines or their surroundings are part of a scope, include that entire scope.
	linesSoFar := cb.linesToShow.ToSlice()
	for _, line := range linesSoFar {
		lineInfo := st.lines[line]
		for childLine := range lineInfo.scope.children() {
			cb.linesToShow.Add(childLine)
		}
	}

	if cb.ParentContext {
		linesSoFar = cb.linesToShow.ToSlice()
		for _, line := range linesSoFar {
			cb.addParentContext(st, line)
		}
	}

	if cb.ChildContext {
		linesSoFar = cb.linesToShow.ToSlice()
		for _, line := range linesSoFar {
			cb.addChildContext(st, line)
		}
	}

	cb.closeGaps()

	return cb.linesToShow
}

// addSurroundingLines expands the set of lines to show by including a
// specified number of lines before and after each line of interest.
func (cb *ContextBuilder) addSurroundingLines(st *sourceTree, linesOfInterest internal.Set[lineNumber]) {
	gap := cb.SurroundingLines

	for line := range linesOfInterest {
		for currentLine := line - gap; currentLine <= line+gap; currentLine++ {
			if currentLine >= lineNumber(len(st.lines)) {
				break
			}
			cb.linesToShow.Add(currentLine)
		}
	}
}

func (cb *ContextBuilder) addParentContext(st *sourceTree, line lineNumber) {
	parentLine := st.lines[line].scope.parent
	if cb.seenParents.Has(parentLine) {
		return
	}
	cb.seenParents.Add(parentLine)

	parentLineInfo := st.lines[parentLine]

	cb.linesToShow.Add(parentLineInfo.scope.start)
	cb.linesToShow.Add(parentLineInfo.scope.end)

	if parentLine == line {
		return
	}

	cb.addParentContext(st, parentLine)
}

// addChildContext adds the context of child scopes. It uses a heuristic to
// show only the beginning of large child scopes to avoid excessive output,
// but shows the full scope if it's small.
func (cb *ContextBuilder) addChildContext(st *sourceTree, line lineNumber) {
	if line == 0 || st.lines[line].scope.size() == 0 {
		return
	}

	lineInfo := st.lines[line]

	limitLine := min(cb.SurroundingLines+lineInfo.scope.start, lineInfo.scope.end)

	// If showing the initial gap would cover over 70% of the scope, just show the whole thing.
	threshold := lineInfo.scope.start + ((lineInfo.scope.size() * 70) / 100)
	if limitLine > threshold {
		limitLine = lineInfo.scope.end
	}

	for currentLine := lineInfo.scope.start; currentLine <= limitLine; currentLine++ {
		cb.linesToShow.Add(currentLine)
		cb.linesToShow.Add(st.lines[currentLine].scope.end)
	}
}

// closeGaps finds small gaps between lines in the current set and adds the
// missing lines to create a more contiguous block of code.
func (cb *ContextBuilder) closeGaps() {
	sortedLines := cb.linesToShow.ToSlice()
	slices.Sort(sortedLines)

	for i := range sortedLines {
		if i == len(sortedLines)-1 {
			continue
		}

		diff := sortedLines[i+1] - sortedLines[i]
		if diff <= cb.GapToClose {
			for currentLine := sortedLines[i] + 1; currentLine <= sortedLines[i+1]; currentLine++ {
				cb.linesToShow.Add(currentLine)
			}
		}
	}
}

type Option func(*ContextBuilder)

// WithGapToClose sets the maximum gap between lines that should be filled in.
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
func WithSurroundingLines(lines lineNumber) Option {
	return func(cb *ContextBuilder) {
		cb.SurroundingLines = lines
	}
}
