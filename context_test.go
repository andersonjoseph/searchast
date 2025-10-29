package findctx

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func mustNewSourceTree(t *testing.T, source string) *sourceTree {
	t.Helper()
	r := strings.NewReader(source)
	st, err := NewSourceTree(context.Background(), r, "test.go")
	if err != nil {
		t.Fatalf("failed to create sourceTree: %v", err)
	}
	return st
}

func TestNewContextBuilder(t *testing.T) {
	t.Run("creates with default values", func(t *testing.T) {
		cb := NewContextBuilder()

		if cb.SurroundingLines != 3 {
			t.Errorf("expected SurroundingLines to be 3, got %d", cb.SurroundingLines)
		}
		if cb.GapToClose != 3 {
			t.Errorf("expected GapToClose to be 3, got %d", cb.GapToClose)
		}
		if !cb.ParentContext {
			t.Error("expected ParentContext to be true, got false")
		}
		if !cb.ChildContext {
			t.Error("expected ChildContext to be true, got false")
		}
	})

	t.Run("applies functional options correctly", func(t *testing.T) {
		cb := NewContextBuilder(
			WithSurroundingLines(5),
			WithGapToClose(10),
			WithParentContext(false),
			WithChildContext(false),
		)

		if cb.SurroundingLines != 5 {
			t.Errorf("expected SurroundingLines to be 5, got %d", cb.SurroundingLines)
		}
		if cb.GapToClose != 10 {
			t.Errorf("expected GapToClose to be 10, got %d", cb.GapToClose)
		}
		if cb.ParentContext {
			t.Error("expected ParentContext to be false, got true")
		}
		if cb.ChildContext {
			t.Error("expected ChildContext to be false, got true")
		}
	})
}

func TestAddContext_DefaultBehavior(t *testing.T) {
	const source = `package main // 0

func first() { // 2
	// comment   // 3
} // 4

func main() { // 6
	if true { // 7
		fmt.Println("target") // 8
	} // 9
} // 10`
	st := mustNewSourceTree(t, source)
	cb := NewContextBuilder()

	// The target line is 8.
	// Defaults: Surrounding=3, Gaps=3, Parents=true, Children=true
	// 1. Surrounding lines for 8: 9,10 and 7,6,5
	// 2. Parent context for line 8: adds scope for `if` (7,9) and `main` (6,10).
	// The final set is contiguous, so no gaps are closed.
	linesOfInterest := NewSetFromSlice([]lineNumber{8})
	expectedLines := NewSetFromSlice([]lineNumber{0, 5, 6, 7, 8, 9, 10})

	actualLines := cb.AddContext(st, linesOfInterest)

	if !reflect.DeepEqual(actualLines, expectedLines) {
		t.Errorf("\nexpected lines: %v\n     got lines: %v", expectedLines.ToSlice(), actualLines.ToSlice())
	}
}

func TestAddContext_CustomOptions(t *testing.T) {
	const source = `package main // 0
// 1
func main() { // 2
	// 3
	// 4
	fmt.Println("one") // 5
	// 6
	// 7
	// 8
	fmt.Println("two") // 9
	// 10
} // 11
`
	st := mustNewSourceTree(t, source)

	testCases := []struct {
		name          string
		opts          []Option
		interest      []lineNumber
		expectedLines []lineNumber
	}{
		{
			name:     "No surrounding lines",
			opts:     []Option{WithSurroundingLines(0)},
			interest: []lineNumber{5},
			// Expects line 5, plus its parent `main` scope (lines 2 and 11)
			// by default, 3 child scopes are included (lines 3,4,5)
			expectedLines: []lineNumber{2, 3, 4, 5, 11},
		},
		{
			name:          "Parent context disabled",
			opts:          []Option{WithParentContext(false), WithSurroundingLines(0)},
			interest:      []lineNumber{5},
			expectedLines: []lineNumber{5},
		},
		{
			name:     "Gap closing with large enough setting",
			opts:     []Option{WithGapToClose(4), WithSurroundingLines(0)},
			interest: []lineNumber{5, 9},
			// Interest is 5 and 9. Gap is 3 (lines 6,7,8).
			// With GapToClose=4, the gap should be filled.
			// Also includes parent `main` scope.
			// the gap between 8 and 10 should be also filled
			expectedLines: []lineNumber{2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		},
		{
			name:     "Gap not closing when too small",
			opts:     []Option{WithGapToClose(2), WithSurroundingLines(0)},
			interest: []lineNumber{5, 9},
			// Interest is 5 and 9. Gap is 3.
			// With GapToClose=2, the gap is NOT filled.
			// Still includes parent `main` scope.
			expectedLines: []lineNumber{2, 5, 9, 10, 11},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cb := NewContextBuilder(tc.opts...)
			linesOfInterest := NewSetFromSlice(tc.interest)
			expected := NewSetFromSlice(tc.expectedLines)

			actual := cb.AddContext(st, linesOfInterest)

			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("\nexpected lines: %v\n     got lines: %v", expected.ToSlice(), actual.ToSlice())
			}
		})
	}
}

func TestAddContext_EdgeCases(t *testing.T) {
	const source = `package main // 0
// 1
func main() { // 2
	// 3
} // 4
// 5`
	st := mustNewSourceTree(t, source)
	cb := NewContextBuilder(WithSurroundingLines(1)) // smaller surrounding for predictability

	t.Run("line of interest at start of file", func(t *testing.T) {
		linesOfInterest := NewSetFromSlice([]lineNumber{0})
		expected := NewSetFromSlice([]lineNumber{0, 1})
		actual := cb.AddContext(st, linesOfInterest)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("\nexpected lines: %v\n     got lines: %v", expected.ToSlice(), actual.ToSlice())
		}
	})

	t.Run("empty lines of interest returns empty set", func(t *testing.T) {
		linesOfInterest := NewSet[lineNumber]()
		expected := NewSet[lineNumber]()
		actual := cb.AddContext(st, linesOfInterest)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("\nexpected lines: %v\n     got lines: %v", expected.ToSlice(), actual.ToSlice())
		}
	})
}

func TestAddContext_StateIsReset(t *testing.T) {
	const source = `package main // 0
func first() { // 1
	// target 1 // 2
} // 3

func second() { // 5
	// target 2 // 6
} // 7
`
	st := mustNewSourceTree(t, source)
	cb := NewContextBuilder(WithSurroundingLines(0), WithParentContext(true))

	// First call
	firstCallLines := NewSetFromSlice([]lineNumber{2})
	expectedFirst := NewSetFromSlice([]lineNumber{1, 2, 3})
	actualFirst := cb.AddContext(st, firstCallLines)

	if !reflect.DeepEqual(actualFirst, expectedFirst) {
		t.Fatalf("first call was incorrect, cannot verify state reset.\nexpected: %v\n     got: %v",
			expectedFirst.ToSlice(), actualFirst.ToSlice())
	}

	// Second call with completely different input should not be affected by the first
	secondCallLines := NewSetFromSlice([]lineNumber{6})
	expectedSecond := NewSetFromSlice([]lineNumber{5, 6, 7})
	actualSecond := cb.AddContext(st, secondCallLines)

	if !reflect.DeepEqual(actualSecond, expectedSecond) {
		t.Errorf("second call produced incorrect result, indicating state was not reset.\nexpected: %v\n     got: %v",
			expectedSecond.ToSlice(), actualSecond.ToSlice())
	}
}
