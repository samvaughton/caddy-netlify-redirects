package naddy

import (
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/tj/go-redirects"
)

const ModuleName = "netlify_redirects"

func init() {
	httpcaddyfile.RegisterHandlerDirective(ModuleName, parseCaddyfile)
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m Middleware

	d := h.Dispenser

	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			allRedirects := ""

			for {
				val := d.Val()
				args := d.RemainingArgs()

				if val == "}" || val == "" {
					break
				}

				allArgs := append([]string{val}, args...)

				for _, arg := range allArgs {
					allRedirects = fmt.Sprintf("%s %s", allRedirects, arg)
				}

				allRedirects = fmt.Sprintf("%s\n", allRedirects)

				if d.Next() == false {
					break
				}
			}

			redir, err := redirects.ParseString(allRedirects)

			if err != nil {
				m.Logger.Error(err.Error())
			} else {
				m.Redirects = redir
			}
		}
	}

	return m, nil
}

var (
	_ caddy.Provisioner = (*Middleware)(nil)
)
