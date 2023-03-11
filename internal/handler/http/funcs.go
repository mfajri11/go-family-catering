package handler

import (
	"fmt"
	"regexp"
)

func regexReplaceAllMultiple(src string, patternNewStrPairs ...string) string {
	if len(patternNewStrPairs)%2 != 0 {
		panic(fmt.Sprintf("invalid number of arguments: the length of patternNewStrPairs must be an even, got: %d", len(patternNewStrPairs)))
	}
	for i := 0; i < len(patternNewStrPairs)-1; i += 2 {
		pattern := regexp.MustCompile(patternNewStrPairs[i])
		src = pattern.ReplaceAllLiteralString(src, patternNewStrPairs[i+1])
	}

	return src
}
