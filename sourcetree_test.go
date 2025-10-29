package findctx

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

type failingReader struct{}

func (fr *failingReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read failed")
}

func TestNewSourceTree(t *testing.T) {
	const sourceForTreeCreation = `package main

import "fmt"

func main() { // Line 4
	if true { // Line 5
		// A comment
		fmt.Println("inside") // Line 7
	}
}
`
	t.Run("successfully creates a tree from valid source", func(t *testing.T) {
		r := strings.NewReader(sourceForTreeCreation)
		st, err := NewSourceTree(context.Background(), r, "test.go")

		if err != nil {
			t.Fatalf("expected no error, but got: %v", err)
		}
		if st == nil {
			t.Fatal("sourceTree should not be nil")
		}

		splitSource := strings.Split(sourceForTreeCreation, "\n")
		if len(st.lines) != len(splitSource) {
			t.Errorf("expected %d lines, but got: %d", len(splitSource), len(st.lines))
		}

		// Assert on a key structural relationship to have confidence in the build.
		// The `if` scope on line 5 should have the `main` function scope (line 4) as its parent.
		parentOfIf := st.lines[5].scope.parent
		if parentOfIf != 4 {
			t.Errorf("expected line 5's parent to be 4, but got %d", parentOfIf)
		}

		// The print statement on line 7 is inside the `if`, so its parent scope should be the `if` on line 5.
		parentOfPrint := st.lines[7].scope.parent
		if parentOfPrint != 5 {
			t.Errorf("expected line 7's parent to be 5, but got %d", parentOfPrint)
		}
	})

	t.Run("returns an error if reading fails", func(t *testing.T) {
		_, err := NewSourceTree(context.Background(), &failingReader{}, "test.go")
		if err == nil {
			t.Fatal("expected an error for a failing reader, but got none")
		}
	})

	t.Run("handles empty source code gracefully", func(t *testing.T) {
		r := strings.NewReader("")
		st, err := NewSourceTree(context.Background(), r, "test.go")
		if err != nil {
			t.Fatalf("expected no error for empty input, but got: %v", err)
		}
		if len(st.lines) != 1 { // A single empty string line
			t.Errorf("expected 1 line for empty input, got: %d", len(st.lines))
		}
	})
}

func TestSearch(t *testing.T) {
	const sourceForSearch = `package main

import "fmt"

func main() {
	fmt.Println("start") // Line 5

	if true { // Line 7
		// A comment
		fmt.Println("inside") // Line 9
	}

	return // Line 12
}
`
	r := strings.NewReader(sourceForSearch)
	st, err := NewSourceTree(context.Background(), r, "test.go")
	if err != nil {
		t.Fatalf("failed to setup sourceTree for search test: %v", err)
	}

	testCases := []struct {
		name          string
		pattern       string
		expectedLines Set[lineNumber]
		expectErr     bool
	}{
		{
			name:          "finds a single unique line",
			pattern:       `"start"`,
			expectedLines: NewSetFromSlice([]lineNumber{5}),
			expectErr:     false,
		},
		{
			name:          "finds multiple lines",
			pattern:       `fmt.Println`,
			expectedLines: NewSetFromSlice([]lineNumber{5, 9}),
			expectErr:     false,
		},
		{
			name:          "finds no matches",
			pattern:       `nonexistent_string`,
			expectedLines: NewSetFromSlice([]lineNumber{}),
			expectErr:     false,
		},
		{
			name:          "matches a comment",
			pattern:       `A comment`,
			expectedLines: NewSetFromSlice([]lineNumber{8}),
			expectErr:     false,
		},
		{
			name:      "returns an error for an invalid regex",
			pattern:   `[`,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lines, err := st.Search(tc.pattern)

			if tc.expectErr {
				if err == nil {
					t.Fatal("expected an error but got none")
				}
				return // Test passed
			}

			if err != nil {
				t.Fatalf("did not expect an error, but got: %v", err)
			}

			if !reflect.DeepEqual(lines, tc.expectedLines) {
				t.Errorf("expected lines %v, but got %v", tc.expectedLines, lines)
			}
		})
	}
}
