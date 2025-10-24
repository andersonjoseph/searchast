package findctx

import (
	"fmt"
	"strings"
)

func FormatOutput(st *sourceTree, linesToShow set[lineNumber], linesOfInterest set[lineNumber]) string {
	if len(linesOfInterest) == 0 {
		return ""
	}

	output := strings.Builder{}
	isGapPrinted := false

	for i, line := range st.lines {
		if !linesToShow.has(lineNumber(i)) {
			if !isGapPrinted {
				output.WriteString("⋮\n")
				isGapPrinted = true
			}

			continue
		}

		isGapPrinted = false
		var spacer string
		if linesOfInterest.has(lineNumber(i)) {
			spacer = "█"
		} else {
			spacer = "│"
		}

		output.WriteString(fmt.Sprintf("%s %s\n", spacer, line.text))
	}

	return output.String()
}
