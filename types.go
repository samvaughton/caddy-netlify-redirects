package naddy

import (
	"github.com/tj/go-redirects"
	"github.com/ucarion/urlpath"
	"go.uber.org/zap"
	"net/url"
)

type Middleware struct {
	Logger    *zap.Logger
	Redirects []redirects.Rule
}

type MatchContext struct {
	Scheme string
}

type MatchResult struct {
	Match      *urlpath.Match
	ResolvedTo *url.URL

	IsMatched      bool
	IsHostRedirect bool

	Error error
}
