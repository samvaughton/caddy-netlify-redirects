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
	Scheme      string
	OriginalUrl *url.URL
}

type MatchResult struct {
	Match      *urlpath.Match
	ResolvedTo *url.URL
	Source     redirects.Rule

	IsMatched      bool
	IsHostRedirect bool

	Error error
}
