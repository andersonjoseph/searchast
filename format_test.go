package searchast

import (
	"strings"
	"testing"
)

func TestTextFormatter_WithColors(t *testing.T) {
	t.Run("colors are disabled by default", func(t *testing.T) {
		formatter := NewTextFormatter()
		if formatter.enableColors != false {
			t.Errorf("expected enableColors to be false by default, got %v", formatter.enableColors)
		}
	})

	t.Run("enables colors when WithColors(true) is called", func(t *testing.T) {
		formatter := NewTextFormatter(WithColors(true))
		if formatter.enableColors != true {
			t.Errorf("expected enableColors to be true, got %v", formatter.enableColors)
		}
	})

	t.Run("disables colors when WithColors(false) is called", func(t *testing.T) {
		formatter := NewTextFormatter(WithColors(false))
		if formatter.enableColors != false {
			t.Errorf("expected enableColors to be false, got %v", formatter.enableColors)
		}
	})
}

func TestTextFormatter_FormatWithColors(t *testing.T) {
	source := `package main

func main() {
	// Comment
	fmt.Println("hello")
}

func test() {
	return nil
}`

	st := mustNewSourceTree(t, source)
	linesOfInterest, err := st.Search("func main")
	if err != nil {
		t.Fatalf("failed to search: %v", err)
	}

	contextBuilder := NewContextBuilder()
	linesToShow := contextBuilder.AddContext(st, linesOfInterest)

	t.Run("applies ANSI color codes to matched lines", func(t *testing.T) {
		formatter := NewTextFormatter(WithColors(true))
		output := formatter.Format(st.Lines(), linesToShow, linesOfInterest)

		if !strings.Contains(output, ansiCodeRed) {
			t.Errorf("expected output to contain ANSI red color code, got: %s", output)
		}

		if !strings.Contains(output, ansiCodeReset) {
			t.Errorf("expected output to contain ANSI reset code, got: %s", output)
		}
	})

	t.Run("does not apply color when colors disabled", func(t *testing.T) {
		formatter := NewTextFormatter(WithColors(false))
		output := formatter.Format(st.Lines(), linesToShow, linesOfInterest)

		if strings.Contains(output, ansiCodeRed) {
			t.Errorf("expected output to not contain ANSI color codes, got: %s", output)
		}
	})

	t.Run("color codes appear only on matched lines", func(t *testing.T) {
		formatter := NewTextFormatter(WithColors(true))
		output := formatter.Format(st.Lines(), linesToShow, linesOfInterest)

		redCodeCount := strings.Count(output, ansiCodeRed)
		resetCodeCount := strings.Count(output, ansiCodeReset)

		if redCodeCount != 1 || resetCodeCount != 1 {
			t.Errorf("expected exactly one red and reset code each, got red: %d, reset: %d", redCodeCount, resetCodeCount)
		}
	})
}
