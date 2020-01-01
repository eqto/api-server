module gitlab.com/tuxer/go-api

go 1.13

replace gitlab.com/tuxer/go-logger => /Users/tuxer/Projects/go/go-logger

replace gitlab.com/tuxer/go-db => /Users/tuxer/Projects/go/go-db

replace gitlab.com/tuxer/go-json => /Users/tuxer/Projects/go/go-json

require (
	github.com/pkg/errors v0.8.1
	gitlab.com/tuxer/go-db v0.0.0-20191030051519-cd407f09d6db
	gitlab.com/tuxer/go-json v0.0.0-00010101000000-000000000000
	gitlab.com/tuxer/go-logger v0.0.0-20181117123619-174e86aaabc1
)
