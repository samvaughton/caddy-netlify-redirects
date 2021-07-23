# caddy-netlify-redirects
Enables Caddy to use Netlify's `_redirect` file format

## Building

Within a `Dockerfile` to build:

# Second stage of build
FROM caddy:2.4.3-builder AS builder

RUN xcaddy build \
--with github.com/samvaughton/caddy-netlify-redirects/v2

# Third stage of build - container that will actually be saved
FROM caddy:2.4.3-alpine as serve

COPY --from=builder /usr/bin/caddy /usr/bin/caddy

## Config

You will need to set the order of the module with this line:

```Caddyfile
order netlify_redirects before redir
```

## Adding redirects

Two methods are supported for loading in Netlify's redirects:

 - Adding a `netlify_redirects` directive within the `Caddyfile` eg:
   ```Caddyfile
    netlify_redirects {
		    /:param/here/:test/two /:param/:test/:two 302
		    /hello/* /redirected/:splat
		    /:param/hello/* /redirected/:param/:splat
    }
   
 - Creating a `_redirects` file within the root of your site directory that is the exact same format as Netlify.