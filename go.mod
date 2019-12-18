module gitlab.com/tuxer/go-api

go 1.13

replace gitlab.com/tuxer/go-logger => /Users/tuxer/Projects/go/go-logger

replace gitlab.com/tuxer/go-db => /Users/tuxer/Projects/go/go-db

replace gitlab.com/tuxer/go-json => /Users/tuxer/Projects/go/go-json

require (
	gitlab.com/tuxer/go-db v0.0.0-20191209074147-f2cc5ad9b3f4
	gitlab.com/tuxer/go-json v0.0.0-20191205071804-11fe6c9c1e64
	gitlab.com/tuxer/go-logger v0.0.0-20181117123619-174e86aaabc1
	google.golang.org/appengine v1.6.5 // indirect
)
