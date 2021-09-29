package naddy

import (
	"fmt"
	"github.com/tj/go-redirects"
	"net/url"
	"testing"
)

func MustParseUrl(urlStr string) *url.URL {
	u, err := url.Parse(urlStr)

	if err != nil {
		panic("cannot parse url " + urlStr)
	}

	return u
}

func TestRedirects(t *testing.T) {
	rules := redirects.Must(redirects.ParseString(`
		/relative/path/:paramOne /relative/path/redirected/:paramOne
		/relative/* /path/changed/:splat/hello
		/test/* /redirect/:splat

		/accommodation/meribel/chalet-le-lys-blanc/5-bedroom-chalet-for-12/  /  410
		/accommodation/morzine/chalet-chez-claude/5-bedroom-chalet-for-13/  /  410
		/accommodation/meribel/chalet-camarine/5-bedroom-chalet-for-13/  /  410
		/accommodation/val-disere/le-chalet-arosa/5-bedroom-for-10/  /  410
		/accommodation/*  /accommodation/  200

		http://one.test:2021/* http://two.test:2021/:splat
		http://lch.k8.rentivo.com/* http://lakecomohomes.com/:splat
		http://lakecomohomes.com/* http://www.lakecomohomes.com/:splat
		one.test:/2021/no-scheme two.test:2021/:splat
	`))

	testCases := []struct {
		RequestUrlIn   *url.URL
		IsMatched      bool
		IsHostRedirect bool
		ExpectedTo     string
	}{
		// Relative
		{
			MustParseUrl("http://www.one.test"),
			false,
			false,
			"",
		},
		{
			MustParseUrl("http://www.one.test/relative/path/foo"),
			true,
			false,
			"/relative/path/redirected/foo",
		},
		{
			MustParseUrl("http://www.one.test/relative/path-not-matched/bar"),
			true,
			false,
			"/path/changed/path-not-matched/bar/hello",
		},
		{
			MustParseUrl("http://one.test:2021/test/foobar"),
			true,
			false,
			"/redirect/foobar",
		},
		// path or no path
		{
			MustParseUrl("http://one.test:2021/accommodation/morzine/chalet-chez-claude/5-bedroom-chalet-for-13"),
			true,
			false,
			"/",
		},
		{
			MustParseUrl("http://one.test:2021/accommodation/morzine/chalet-chez-claude/5-bedroom-chalet-for-13/"),
			true,
			false,
			"/",
		},
		// Host
		{
			MustParseUrl("http://one.test:2021/redirect/foobar"),
			true,
			true,
			"http://two.test:2021/redirect/foobar",
		},
		{
			MustParseUrl("http://one.test:2021/test/foobar"),
			true,
			false,
			"/redirect/foobar",
		},
		{
			MustParseUrl("http://one.test:2021/redirect/foobar"),
			true,
			true,
			"http://two.test:2021/redirect/foobar",
		},
		{
			MustParseUrl("http://lch.k8.rentivo.com/example"),
			true,
			true,
			"http://lakecomohomes.com/example",
		},
		{
			MustParseUrl("http://lch.k8.rentivo.com/"),
			true,
			true,
			"http://lakecomohomes.com/",
		},
		{
			MustParseUrl("http://lch.k8.rentivo.com"),
			true,
			true,
			"http://lakecomohomes.com/",
		},
		{
			MustParseUrl("http://lakecomohomes.com"),
			true,
			true,
			"http://www.lakecomohomes.com/",
		},
	}

	ctx := &MatchContext{
		Scheme: "http",
	}

	for _, tc := range testCases {
		name := fmt.Sprintf("%s/match=%v/host=%v", tc.RequestUrlIn.String(), tc.IsMatched, tc.IsHostRedirect)
		t.Run(name, func(t *testing.T) {

			matchFound := false
			for _, rule := range rules {
				result := MatchUrlToRule(rule, tc.RequestUrlIn, ctx)

				if result.Error != nil {
					t.Errorf("an error occurred matching %s", result.Error.Error())

					continue
				}

				if result.IsMatched == false {
					continue // skip till we find the rule
				}

				matchFound = true

				if result.IsHostRedirect != tc.IsHostRedirect {
					t.Errorf("want host=%v, got host=%v", tc.IsHostRedirect, result.IsHostRedirect)
				}

				if tc.IsMatched {
					if result.ResolvedTo == nil {
						t.Errorf("want resolvedTo=%s, got resolvedTo=<nil>", tc.ExpectedTo)

						break
					} else if result.ResolvedTo.String() != tc.ExpectedTo {
						t.Errorf("want resolvedTo=%s, got resolvedTo=%s", tc.ExpectedTo, result.ResolvedTo.String())

						break
					}
				}

				break
			}

			if matchFound == false && tc.IsMatched != false {
				t.Errorf("want match=%v, got match=%v", tc.IsMatched, false)
			}

		})
	}
}
