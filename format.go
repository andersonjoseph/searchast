package searchast

import (
	"fmt"
	"strings"
)

const (
	ansiCodeReset = "\033[0m"
	ansiCodeRed   = "\033[31m"
)

type Formatter interface {
	Format(lines []line, linesToShow Set[lineNumber], linesToHighlight Set[lineNumber]) string
}

type TextFormatter struct {
	lineNumbers     bool
	highlightSymbol string
	contextSymbol   string
	gapSymbol       string
	spacer          string
	enableColors    bool
}

type TextFormatterOption func(*TextFormatter)

func NewTextFormatter(opts ...TextFormatterOption) *TextFormatter {
	formatter := &TextFormatter{
		lineNumbers:     true,
		highlightSymbol: "█",
		contextSymbol:   "│",
		gapSymbol:       "⋮",
		spacer:          " ",
		enableColors:    false,
	}

	for _, opt := range opts {
		opt(formatter)
	}

	return formatter
}

func WithLineNumbers(lineNumbers bool) TextFormatterOption {
	return func(tf *TextFormatter) {
		tf.lineNumbers = lineNumbers
	}
}

func WithHighlightSymbol(symbol string) TextFormatterOption {
	return func(tf *TextFormatter) {
		tf.highlightSymbol = symbol
	}
}

func WithContextSymbol(symbol string) TextFormatterOption {
	return func(tf *TextFormatter) {
		tf.contextSymbol = symbol
	}
}

func WithGapSymbol(symbol string) TextFormatterOption {
	return func(tf *TextFormatter) {
		tf.gapSymbol = symbol
	}
}

func WithSpacer(spacer string) TextFormatterOption {
	return func(tf *TextFormatter) {
		tf.spacer = spacer
	}
}

func WithColors(enabled bool) TextFormatterOption {
	return func(tf *TextFormatter) {
		tf.enableColors = enabled
	}
}

func maxLineNumber(linesToShow Set[lineNumber]) int {
	// Calculate the width needed for line numbers
	var maxLineNumber lineNumber
	for line := range linesToShow {
		maxLineNumber = max(maxLineNumber, line)
	}
	return len(fmt.Sprintf("%d", maxLineNumber))
}

func (tf *TextFormatter) Format(lines []line, linesToShow Set[lineNumber], linesToHighlight Set[lineNumber]) string {
	if len(linesToShow) == 0 || len(linesToHighlight) == 0 {
		return ""
	}

	output := strings.Builder{}
	isGapPrinted := false

	lineNumberWidth := maxLineNumber(linesToShow)
	for i, line := range lines {
		if !linesToShow.Has(lineNumber(i)) {
			if !isGapPrinted {
				var gapPrefix string
				if tf.lineNumbers {
					gapPrefix = fmt.Sprintf("%*s%s", lineNumberWidth, "", tf.gapSymbol)
				} else {
					gapPrefix = tf.gapSymbol
				}
				output.WriteString(gapPrefix + "\n")
				isGapPrinted = true
			}

			continue
		}

		isGapPrinted = false
		var symbol string
		if linesToHighlight.Has(lineNumber(i)) {
			symbol = tf.highlightSymbol
		} else {
			symbol = tf.contextSymbol
		}

		var prefix string
		if tf.lineNumbers {
			prefix = fmt.Sprintf("%*d%s%s%s", lineNumberWidth, i+1, tf.spacer, symbol, tf.spacer)
		} else {
			prefix = fmt.Sprintf("%s%s", symbol, tf.spacer)
		}

		var lineText string
		if tf.enableColors && linesToHighlight.Has(lineNumber(i)) {
			lineText = ansiCodeRed + line.text + ansiCodeReset
		} else {
			lineText = line.text
		}

		output.WriteString(fmt.Sprintf("%s%s\n", prefix, lineText))
	}

	return output.String()
}
