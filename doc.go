// Package routing provides a naive router with
// regular expression support.
//
// When using named sub groups in a regex path,
// the named groups matched expression will be set on
// the request's context by it's name. So if you register
// the route ^/foo/(?P<param>[0-9]+)$, you can get it by
// doing *http.Request.Context().Value("param") in your handler.
//
// The routes can be supplied with metadata though a
// fluent API to generate HTML documentation.
//
// If regular expression matchings overlap, they take
// precedence by the order they have been added.
package routing
