# caddy-netlify-redirects
Enables Caddy to use Netlify's `_redirect` file format

This module tries to replicate the way Netlify's _redirects file works.

It does support:

   - Host redirection
   - Path redirection
   - Other status codes such as `410 Gone` (with a redirect after returning the 410)

It does not (currently) support:

   - Header matching
   - Query string matching
   - HTTP -> HTTPS redirection

If you wish to add these features, please open an issue/PR.

## Building via Docker

As an example, within a dockerfile you can build Caddy with this custom module:

```dockerfile
FROM caddy:2.4.3-builder AS builder

RUN xcaddy build \
   --with github.com/samvaughton/caddy-netlify-redirects/v2
   
FROM caddy:2.4.3-alpine as serve

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
COPY ./Caddyfile /etc/caddy/Caddyfile

# Copy over your built assets for your webapp, this could be from gatbsy which includes a _redirects file
COPY --from=node-builder /usr/src/app/packages/rentivo-gatsby-site/public /srv
```

## Config

You will need to set the order of the module with this line:

```Caddyfile
order netlify_redirects before redir
```

## Adding redirects

Put a `netlify_redirects` directive within the `Caddyfile` eg:

```Caddyfile
netlify_redirects {
   /:param/here/:test/two /:param/:test/:two 302
   /hello/* /redirected/:splat
   /:param/hello/* /redirected/:param/:splat
}
```

You can also import a `_netlify` file:

```Caddyfile
netlify_redirects {
   import /srv/_redirects
}
```