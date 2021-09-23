package naddy

import (
	"errors"
	"fmt"
	"github.com/tj/go-redirects"
	"github.com/ucarion/urlpath"
	"net/url"
	"strings"
)

func ParseUrlWithContext(urlStr string, ctx *MatchContext) (*url.URL, error) {
	original := urlStr

	if !strings.Contains(urlStr, "http://") && !strings.Contains(urlStr, "https://") {
		urlStr = fmt.Sprintf("%s%s", ctx.Scheme, urlStr)

		// use url parse to test if it has a Host, if it doesn't revert back to original
		testStr, err := url.Parse(urlStr)

		if err != nil {
			return nil, err
		}

		if testStr.Host == "" {
			urlStr = original
		}
	}

	return url.Parse(urlStr)
}

func MatchUrlToRule(rule redirects.Rule, reqUrl *url.URL, ctx *MatchContext) MatchResult {
	if reqUrl.Host == "" || reqUrl.Scheme == "" {
		return MatchResult{
			ResolvedTo:     nil,
			IsMatched:      false,
			IsHostRedirect: false,
			Error:          errors.New("request url must have both host and scheme"),
		}
	}

	/*
	 * Perform the match as soon as possible on the path itself, as we may need the resolved To
	 */

	from, errFrom := ParseUrlWithContext(rule.From, ctx)

	if errFrom != nil {
		return MatchResult{
			ResolvedTo:     nil,
			IsMatched:      false,
			IsHostRedirect: false,
			Error:          errFrom,
		}
	}

	path := urlpath.New(strings.Trim(from.Path, "/"))
	matched, ok := path.Match(strings.Trim(reqUrl.Path, "/"))

	if !ok {
		return MatchResult{
			ResolvedTo:     nil,
			IsMatched:      false,
			IsHostRedirect: false,
		}
	}

	toPath := rule.To
	toPath = replaceParams(toPath, matched)
	toPath = replaceSplat(toPath, matched)

	to, errTo := ParseUrlWithContext(toPath, ctx)

	if errTo != nil {
		return MatchResult{
			ResolvedTo:     nil,
			IsMatched:      false,
			IsHostRedirect: false,
			Error:          errTo,
		}
	}

	hostToHost := from.Host != "" && to.Host != ""
	hostToRelative := from.Host != "" && to.Host == ""
	relativeToHost := from.Host == "" && to.Host != ""

	// dont need to redirect if on the same host, or no host on rule.To
	isHostRedirect := to.Host != "" && to.Host != reqUrl.Host

	skipMatch := MatchResult{
		ResolvedTo:     to,
		Match:          &matched,
		IsMatched:      false,
		IsHostRedirect: false,
	}

	if (hostToHost || hostToRelative) && from.Host != reqUrl.Host {
		return skipMatch
	}

	if relativeToHost && to.Host == reqUrl.Host {
		return skipMatch
	}

	specialToRules := strings.Split(rule.To, "|")

	for _, sItem := range specialToRules {
		if sItem == "$ENFORCE_TRAILING_SLASH" {
			// check to make sure this isn't a file request
			parts := strings.Split(ctx.OriginalUrl.Path, ".")
			if
			// make sure parts is greater than two, and then verify that the final element is one of these
			len(parts) >= 2 &&
				len(parts[len(parts)-1]) >= 2 &&
				len(parts[len(parts)-1]) <= 5 {
				return skipMatch
			}

			if strings.HasSuffix(ctx.OriginalUrl.Path, "/") == false {
				// redirect
				prefixedTo := reqUrl
				prefixedTo.Path = fmt.Sprintf("%s/", prefixedTo.Path)

				return MatchResult{
					ResolvedTo:     prefixedTo,
					Match:          &matched,
					IsMatched:      true,
					IsHostRedirect: isHostRedirect,
				}
			}

			return skipMatch
		}
	}

	return MatchResult{
		ResolvedTo:     to,
		Match:          &matched,
		IsMatched:      true,
		IsHostRedirect: isHostRedirect,
	}
}

func replaceParams(to string, matched urlpath.Match) string {
	if len(matched.Params) > 0 {
		for key, value := range matched.Params {
			to = strings.ReplaceAll(to, ":"+key, value)
		}
	}

	return to
}

func replaceSplat(to string, matched urlpath.Match) string {
	if matched.Trailing != "" {
		to = strings.ReplaceAll(to, ":splat", matched.Trailing)
	}

	return to
}
