package routing

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
)

func NewRoute(p string) *Route {
	r := defaultRoute(p)
	r.MatchType = equalityMatch
	return r
}

func NewRexRoute(p string) *Route {
	r := defaultRoute(p)
	r.MatchType = regexMatch
	return r
}

func defaultRoute(p string) *Route {
	return &Route{
		Rex:             regexp.MustCompile(p),
		Method:          http.MethodGet,
		Middleware:      []Middlewarer{},
		QueryParameters: []*QueryParameter{},
		PathArguments:   []*PathArgument{},
	}
}

// Type Route represents a path and is registered with the Router.
//
// A Route exposesa a fluid api to set metadata that is used to
// create an HTML document.
type Route struct {
	MatchType       MatchType
	Rex             *regexp.Regexp
	Handler         http.Handler
	RequestBody     *Body
	Response        *Body
	QueryParameters []*QueryParameter
	PathArguments   []*PathArgument
	Method          string
	Middleware      []Middlewarer
	Tag             string
	Description     string
}

// Middlewarer represents the interface for middlewares
// that can be documented automaically.
type Middlewarer interface {
	Middleware() Middleware
	Description() string
}

type Middleware func(http.Handler) http.Handler

type MatchType int

const (
	equalityMatch MatchType = iota
	regexMatch    MatchType = iota
)

func (r *Route) SetHandler(h http.Handler) *Route {
	r.Handler = h
	return r
}

func (r *Route) SetRequestBody(b *Body) *Route {
	r.RequestBody = b
	return r
}

func (r *Route) SetDescription(d string) *Route {
	r.Description = d
	return r
}
func (r *Route) SetResponse(b *Body) *Route {
	r.Response = b
	return r
}

func (r *Route) SetHandlerFunc(h http.HandlerFunc) *Route {
	r.Handler = http.HandlerFunc(h)
	return r
}

func (route *Route) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	if len(route.Middleware) == 0 {
		route.Handler.ServeHTTP(w, r)
		return
	}

	last := route.Handler
	for i := len(route.Middleware) - 1; i >= 0; i-- {
		last = route.Middleware[i].Middleware()(last)
	}
	last.ServeHTTP(w, r)
}

func (r *Route) AddMiddleware(mw Middlewarer) *Route {
	r.Middleware = append(r.Middleware, mw)
	return r
}

func (r *Route) SetMethod(m string) *Route {
	r.Method = m
	return r
}

func (r *Route) SetTag(t string) *Route {
	r.Tag = t
	return r
}
func (r *Route) AddQueryParameter(p *QueryParameter) *Route {
	r.QueryParameters = append(r.QueryParameters, p)
	return r
}

func (r *Route) AddPathArgument(a *PathArgument) *Route {
	r.PathArguments = append(r.PathArguments, a)
	return r
}

func NewResponse(contentType, name string, example interface{}) *Body {
	return NewBody(contentType, name, example)
}

type Body struct {
	*HTTPArgument

	ContentType string
}

func NewBody(contentType, name string, example interface{}) *Body {
	return &Body{NewHTTPArgument(name, example), contentType}
}

type QueryParameter HTTPArgument

func NewQueryParameter(name string, example interface{}) *QueryParameter {
	return (*QueryParameter)(NewHTTPArgument(name, example))
}

type PathArgument HTTPArgument

func NewPathArgument(name string, example interface{}) *PathArgument {
	return (*PathArgument)(NewHTTPArgument(name, example))
}

type HTTPArgument struct {
	Name    string
	Example interface{}
}

func NewHTTPArgument(name string, example interface{}) *HTTPArgument {
	return &HTTPArgument{name, example}
}

func absoluteTypeName(v interface{}) string {
	t := reflect.Indirect(reflect.ValueOf(v)).Type()
	return fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
}

func typeName(v interface{}) string {
	t := reflect.Indirect(reflect.ValueOf(v)).Type()
	return t.Name()
}
