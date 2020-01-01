package api

import (
	"regexp"

	"github.com/pkg/errors"
)

var (
	ErrResourceNotFound   = errors.New(`resource not found`)
	ErrDatabaseConnection = errors.New(`database connection problem`)
	ErrAuthentication     = errors.New(`authentication error`)
	ErrRouting            = errors.New(`error processing route`)
	ErrMissingParameter   = errors.New(`error missing required parameter: %s`)
	ErrUnknownQueryType   = errors.New(`unknown query type %s`)
	ErrFunctionNotFound   = errors.New(`unknown server func %s`)
)

var (
	regexes = []*regexp.Regexp{
		regexp.MustCompile(`(?Uis)Error 1062: (Duplicate entry '.*')`),
		regexp.MustCompile(`(?Uis)Error 1054: (Unknown column '.*')`),
	}
)

//ErrWrap msg is error or string
func wrapError(e error, msg interface{}) error {
	strMsg := e.Error()
	for _, regex := range regexes {
		if matches := regex.FindStringSubmatch(strMsg); len(matches) > 0 {
			return errors.Wrap(e, matches[1])
		}
	}

	switch msg := msg.(type) {
	case error:
		return errors.Wrap(e, msg.Error())
	case string:
		return errors.Wrap(e, msg)
	}
	return e
}
