package api

import (
	"regexp"
	"strings"
)

func normalizePath(path string) string {
	regex := regexp.MustCompile(`([A-Z]+)`)
	path = regex.ReplaceAllString(path, `_$1`)
	path = strings.ToLower(path)
	validPath := false
	if strings.HasPrefix(path, `/`) {
		validPath = true
		path = path[1:]
	}
	if strings.HasPrefix(path, `_`) {
		path = path[1:]
	}
	if validPath {
		path = `/` + path
	}
	path = strings.ReplaceAll(path, `/_`, `/`)
	return path
}
