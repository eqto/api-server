package api

import (
	"regexp"
)

var (
	regexDuplicate     = regexp.MustCompile(`(?Uis)Error 1062: (Duplicate entry '.*')`)
	regexUnknownColumn = regexp.MustCompile(`(?Uis)Error 1054: (Unknown column '.*')`)
)

func parseError(e error) string {
	msg := e.Error()
	if matches := regexDuplicate.FindStringSubmatch(msg); len(matches) > 0 {
		return matches[1]
	} else if matches := regexDuplicate.FindStringSubmatch(msg); len(matches) > 0 {
		return matches[1]
	}
	return `Error`
}
