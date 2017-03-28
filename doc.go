// Package routing provides a naive router with
// regular expression support.
//
// When using named sub groups in a regex path,
// the named group's match value will be set on
// the request's context by it's name.
//
// The routes can be supplied with metadata though a
// fluent API to generate HTML documentation.
package routing
