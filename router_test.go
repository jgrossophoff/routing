package routing

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type TestModel struct {
	Email  string
	Active bool
}

var testRouter *Router

func TestMain(m *testing.M) {
	testRouter = NewRouter()

	reqResp := NewBody(
		"application/json",
		"",
		&TestModel{
			Email:  "foo@bar.de",
			Active: true,
		},
	)

	testRouter.Add(NewRoute("/foo/bar").SetMethod(http.MethodGet).
		SetHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("passt"))
		}))

	testRouter.Add(NewRexRoute("/foo/(?P<user_id>[0-9]+)/bar").
		SetDescription("Return bar of foo with user_id").
		SetRequestBody(reqResp).
		AddQueryParameter(NewQueryParameter(
			"offset",
			1,
		)).
		AddQueryParameter(NewQueryParameter(
			"limit",
			15,
		)).
		SetResponse(reqResp).
		SetTag("Foos").
		SetMethod(http.MethodPost).
		SetHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := []byte(r.Context().Value("user_id").(string))
			w.Write(id)
		}))

	testRouter.Add(NewRoute("/middleware").
		SetMethod(http.MethodGet).
		AddQueryParameter(NewQueryParameter("offset", 3)).
		AddMiddleware(new(TestMiddlewareFirst)).
		AddMiddleware(new(testMiddlewareSecond)).
		SetHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mw := []byte(r.Context().Value(testMiddlewareCtxValue).(string))
			w.Write(mw)
		}))

	testRouter.Add(NewRoute("/canceled").
		SetMethod(http.MethodGet).
		AddMiddleware(new(testMiddlewareCancel)).
		SetHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("non empty response"))
		}))

	os.Exit(m.Run())
}

func TestDocGen(t *testing.T) {
	/*
		tmpFilePath := "/tmp/gendocs.html"
		f, err := os.Create(tmpFilePath)
		if err != nil {
			t.Errorf("unable to create temp file %s: %s", tmpFilePath, err)
		}
	*/
	var buf bytes.Buffer
	if err := WriteDocs(&buf, "Foo API", "1.0.0", "https://foo.bar.de/api", testRouter); err != nil {
		t.Errorf("unable to render template: %s", err)
	}
}

func TestMiddlewareCancel(t *testing.T) {
	req := httptest.NewRequest("GET", "/canceled", nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	if string(body) != "" {
		t.Errorf("expected empty response body from cancel middleware, was %q", body)
	}
}

func TestMiddlewareOrdered(t *testing.T) {
	req := httptest.NewRequest("GET", "/middleware", nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	if string(body) != overrittenCtxValue {
		t.Errorf("expected returned middleware value from ctx to be %q, was %q", overrittenCtxValue, body)
	}
}

func TestPathParams(t *testing.T) {
	id := "123"
	req := httptest.NewRequest("POST", "/foo/"+id+"/bar", nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	if string(body) != id {
		t.Errorf("expected returned path param user_id to be %q, was %q", id, body)
	}
}

func BenchmarkMatching(b *testing.B) {
	const numRoutes = 10000

	rr := NewRouter()
	for i := 0; i < numRoutes; i++ {
		rr.Add(NewRexRoute(fmt.Sprintf("/test/route/%d", i)))
	}
	er := NewRouter()
	for i := 0; i < numRoutes; i++ {
		er.Add(NewRoute(fmt.Sprintf("/test/route/%d", i)))
	}

	b.Run("match regex", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rr.Match("GET", fmt.Sprintf("/test/route/%d", rand.Intn(numRoutes)))
		}
	})
	b.Run("match equal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			er.Match("GET", fmt.Sprintf("/test/route/%d", rand.Intn(numRoutes)))
		}
	})

}

func TestMatching(t *testing.T) {
	matches := []*http.Request{
		httptest.NewRequest("GET", "/foo/bar", nil),
		httptest.NewRequest("POST", "/foo/1/bar", nil),
		httptest.NewRequest("POST", "/foo/123/bar", nil),
	}
	noMatches := []*http.Request{
		httptest.NewRequest("GET", "/", nil),
		httptest.NewRequest("GET", "/not/found", nil),
		httptest.NewRequest("GET", "/123", nil),
	}

	for _, v := range matches {
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, v)
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected to get 200 for %s %s, status was %d", v.Method, v.URL.Path, resp.StatusCode)
		}
	}
	for _, v := range noMatches {
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, v)
		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected to get 404 for /not/found, status was %d", resp.StatusCode)
		}
	}
}

const testMiddlewareCtxValue = "middleware"
const overrittenCtxValue = "overwritten"

type testMiddlewareCancel struct{}

type TestMiddlewareFirst struct{}

type testMiddlewareSecond struct{}

func (mw *testMiddlewareSecond) Middleware() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), testMiddlewareCtxValue, "overwritten"))
			h.ServeHTTP(w, r)
		})
	}
}

func (mw *TestMiddlewareFirst) Middleware() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), testMiddlewareCtxValue, testMiddlewareCtxValue))
			h.ServeHTTP(w, r)
		})
	}
}

func (mw *testMiddlewareCancel) Middleware() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// handle nothing
		})
	}
}

func (mw *TestMiddlewareFirst) Description() string  { return "Test Description" }
func (mw *testMiddlewareSecond) Description() string { return "Test Description" }
func (mw *testMiddlewareCancel) Description() string { return "Test Description" }

func (mw *TestMiddlewareFirst) Name() string  { return "First Middleware" }
func (mw *testMiddlewareSecond) Name() string { return "Second middleware" }
func (mw *testMiddlewareCancel) Name() string { return "Cancel miiddleware" }
