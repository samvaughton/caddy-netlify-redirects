# caddy-netlify-redirects
Enables Caddy to use Netlify's `_redirect` file format

## Building

Within a `Dockerfile` to build:

# Second stage of build

As an example, within a dockerfile you can build caddy with this custom module:

```dockerfile
FROM caddy:2.4.3-builder AS builder

RUN xcaddy build \
   --with github.com/samvaughton/caddy-netlify-redirects/v2
   
FROM caddy:2.4.3-alpine as serve

COPY --from=builder /usr/bin/caddy /usr/bin/caddy

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