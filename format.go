package findctx

import (
	"fmt"
	"strings"
)

func FormatOutput(st *sourceTree, linesToShow Set[lineNumber], linesOfInterest Set[lineNumber]) string {
	if len(linesOfInterest) == 0 {
		return ""
	}

	output := strings.Builder{}
	isGapPrinted := false

	for i, line := range st.lines {
		if !linesToShow.Has(lineNumber(i)) {
			if !isGapPrinted {
				output.WriteString("⋮\n")
				isGapPrinted = true
			}

			continue
		}

		isGapPrinted = false
		var spacer string
		if linesOfInterest.Has(lineNumber(i)) {
			spacer = "█"
		} else {
			spacer = "│"
		}

		output.WriteString(fmt.Sprintf("%s %s\n", spacer, line.text))
	}

	return output.String()
}
