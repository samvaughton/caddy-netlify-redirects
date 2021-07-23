package naddy

import (
	"bytes"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/tj/go-redirects"
	"html"
	"net/http"
	"os"
	"path"
	"strconv"
)

func init() {
	caddy.RegisterModule(Middleware{})
}

func (Middleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers." + ModuleName,
		New: func() caddy.Module { return new(Middleware) },
	}
}

func (m *Middleware) Provision(ctx caddy.Context) error {
	m.Logger = ctx.Logger(m)

	/*
	 * Try and locate a _redirects file, if found, parse it and append
	 */

	repl := caddy.NewReplacer()
	root := repl.ReplaceAll("{http.vars.root}", ".")
	redirectsFile := caddyhttp.SanitizedPathJoin(root, "_redirects")

	if file, err := os.Stat(redirectsFile); err == nil {
		fileBytes, rErr := os.ReadFile(file.Name())

		if rErr == nil {
			m.Logger.Info("located _redirects file")

			/*
			 * Any Caddyfile rules will take precedence over _redirects
			 * We need to add any _redirect rules to the end of this array
			 * Since this _redirects file is prone to a lot of change, we
			 * will make the error handling a bit more lax here
			 */

			rules, err := redirects.Parse(bytes.NewReader(fileBytes))

			if err != nil {
				m.Logger.Error("could not parse _redirects file")
			} else {
				for _, rule := range rules {
					m.Redirects = append(m.Redirects, rule)
				}
			}

		} else {
			m.Logger.Error("error reading _redirects file")
		}
	} else {
		m.Logger.Warn("could not locate _redirects file")
	}

	m.Logger.Info(fmt.Sprintf("provisioned with %v redirects", len(m.Redirects)))

	return nil
}

func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {

	scheme := "http://"
	if r.TLS != nil {
		scheme = "https://"
	}

	mc := &MatchContext{
		Scheme: scheme,
	}

	for _, rule := range m.Redirects {

		reqUrl, err := ParseUrlWithContext(path.Join(r.Host, r.URL.Path), mc)

		if err != nil {
			m.Logger.Error(err.Error())

			continue
		}

		result := MatchUrlToRule(rule, reqUrl, mc)

		if result.Error != nil {
			m.Logger.Error(result.Error.Error())

			continue
		}

		if result.IsMatched == false {
			continue
		}

		return m.handleRedirectResponse(result, rule, w, r)
	}

	return next.ServeHTTP(w, r)
}

func (m *Middleware) handleRedirectResponse(result MatchResult, rule redirects.Rule, w http.ResponseWriter, r *http.Request) error {
	body := ""
	if rule.Status < 300 || rule.Status >= 400 {
		body = buildHtmlRedirect(result.ResolvedTo.String())
	}

	s := &caddyhttp.StaticResponse{
		StatusCode: caddyhttp.WeakString(strconv.Itoa(rule.Status)),
		Headers: http.Header{
			"Location": []string{result.ResolvedTo.String()},
		},
		Body: body,
	}

	err := s.ServeHTTP(w, r, nil)
	if err != nil {
		m.Logger.Error(fmt.Sprintf("did not expect an error, but got: %v", err))
	}

	return err
}

func buildHtmlRedirect(url string) string {
	const metaRedir = `<!DOCTYPE html>
<html>
	<head>
		<title>Redirecting...</title>
		<script>window.location.replace("%s");</script>
		<meta http-equiv="refresh" content="0; URL='%s'">
	</head>
	<body>Redirecting to <a href="%s">%s</a>...</body>
</html>
`
	safeTo := html.EscapeString(url)
	return fmt.Sprintf(metaRedir, safeTo, safeTo, safeTo, safeTo)
}

var (
	_ caddyhttp.MiddlewareHandler = (*Middleware)(nil)
)
