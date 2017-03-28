package routing

import (
	"context"
	"fmt"
	"net/http"
	"sort"
)

func NewRouter() *Router {
	return &Router{
		rexRoutes:      []*Route{},
		equalityRoutes: make(map[string]*Route),
		routesByTag:    make(map[string][]*Route),
	}
}

type Router struct {
	rexRoutes      []*Route
	equalityRoutes map[string]*Route
	routesByTag    map[string][]*Route
}

// ServeHTTP satisfied the net/http.Handler interface.
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route := router.Match(r.Method, r.URL.Path)

	if route == nil {
		http.NotFound(w, r)
	} else {
		ctx := context.WithValue(r.Context(), "route", route)

		names := route.Rex.SubexpNames()
		vals := route.Rex.FindStringSubmatch(r.URL.Path)

		for i, n := range names {
			if i == 0 {
				continue
			}
			ctx = context.WithValue(ctx, n, vals[i])
		}

		route.ServeHTTP(w, r.WithContext(ctx))
	}
}

// Match will search and return the route registered for the
// HTTP method and URL path.
//
// Returns nil if none found.
func (r *Router) Match(method, path string) *Route {
	if route, exists := r.equalityRoutes[uniquePath(method, path)]; exists {
		return route
	}

	for _, route := range r.rexRoutes {
		if method == route.Method && route.Rex.MatchString(path) {
			return route
		}
	}

	return nil
}

func uniquePath(method, path string) string {
	return method + path
}

// Middlewares returns all middlewares for all routes.
func (r *Router) Middlewares() map[string]Middlewarer {
	mw := make(map[string]Middlewarer)
	for _, r := range r.rexRoutes {
		for _, m := range r.Middleware {
			mw[absoluteTypeName(m)] = m
		}
	}
	for _, r := range r.equalityRoutes {
		for _, m := range r.Middleware {
			mw[absoluteTypeName(m)] = m
		}
	}
	return mw
}

func (r *Router) Tags() []string {
	tags := make([]string, 0, len(r.routesByTag))
	for k := range r.routesByTag {
		tags = append(tags, k)
	}
	sort.Strings(tags)
	return tags
}

func (r *Router) RoutesForTag(t string) []*Route {
	return r.routesByTag[t]
}

func (r *Router) Add(route *Route) error {
	if _, exists := r.routesByTag[route.Tag]; !exists {
		r.routesByTag[route.Tag] = []*Route{}
	}
	r.routesByTag[route.Tag] = append(r.routesByTag[route.Tag], route)

	if route.MatchType == equalityMatch {
		uniq := uniquePath(route.Method, route.Rex.String())
		r.equalityRoutes[uniq] = route
	} else if route.MatchType == regexMatch {
		r.rexRoutes = append(r.rexRoutes, route)
	} else {
		return &ErrInvalidRouteMatchType{route}
	}

	return nil
}

type ErrInvalidRouteMatchType struct{ r *Route }

func (e *ErrInvalidRouteMatchType) Error() string {
	return fmt.Sprintf("invalid match type %s", e.r.MatchType)
}
