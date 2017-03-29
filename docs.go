package routing

import (
	"bytes"
	"encoding/json"
	"html"
	"html/template"
	"io"
	"reflect"
	"time"

	"github.com/davecgh/go-spew/spew"
)

// Type APIStruct holds data needed to render the HTML document.
type API struct {
	Name    string
	Version string
	BaseURL string
	Router  *Router
}

func WriteDocs(
	out io.Writer,
	name string,
	version string,
	baseURL string,
	router *Router,
) error {
	return tmpl.ExecuteTemplate(out, "main", &API{name, version, baseURL, router})
}

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{
	"EscapeHTML": html.EscapeString,
	"MarshalJSON": func(v interface{}) (string, error) {
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}

		var buf bytes.Buffer
		if err := json.Indent(&buf, b, "", "\t"); err != nil {
			return "", err
		}

		return buf.String(), nil
	},
	"PrettyPrint": func(v interface{}) string {
		spew.Config.Indent = "\t"
		return spew.Sdump(v)
	},
	"HasField": func(v interface{}, name string) bool {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		if rv.Kind() != reflect.Struct {
			return false
		}
		return rv.FieldByName(name).IsValid()
	},
	"Now": func() string {
		return time.Now().Format(time.RFC850)
	},
}).Parse(`
{{define "main"}}
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<link href="https://cdnjs.cloudflare.com/ajax/libs/skeleton/2.0.4/skeleton.min.css" rel="stylesheet">
	<style>
		.route, nav {
			margin-bottom: 40px;
		}
		.middleware,
		.route > .title,
		.argument {
			margin-left: 40px;
		}

		.middleware .description,
		.route > .content {
			margin-left: 80px;
		}
	</style>
	<title>{{.Name}} API documentation</title>
</head>
<body>
	<div class="container">
		<h1>{{.Name}} docs</h1>
		<p>
			<strong>{{.BaseURL}}</strong>
			<br>
			<strong>v{{.Version}}</strong>
		</p>

		<hr>

		<h3>Index</h3>
		<h4>Routes</h4>
		<nav>
			{{range $tag := .Router.Tags}}
				{{if $tag}}
					<strong>{{$tag}}</strong>
				{{else}}
					<strong>Untagged</strong>
				{{end}}
				{{range $.Router.RoutesForTag $tag}}
					<div>
						<a href="#route-{{EscapeHTML .Rex.String}}">
						{{.Method}} {{.Rex.String}}
						</a>
					</div>
				{{end}}
			{{end}}
		</nav>

		<hr>

		<main>
			{{range $tag := .Router.Tags}}
				{{if $tag}}
					<h3>{{$tag}}</h3>
				{{else}}
					<h3>Untagged</h3>
				{{end}}
				{{range $.Router.RoutesForTag $tag}}
					<div class="route" id="route-{{EscapeHTML .Rex.String}}">
						<div class="title">
							<h4>
								{{.Method}} {{.Rex.String}}
							</h4>
						</div>
						<div class="content">
							{{if .Description}}
								<p>{{.Description}}</p>
							{{end}}
							{{if .RequestBody}}
								<h5>Request body</h5>
								{{template "argument" .RequestBody}}
							{{end}}
							{{if .PathArguments}}
								<h5>Path arguments</h5>
								{{range .PathArguments}}
									{{template "argument" .}}
								{{end}}
							{{end}}
							{{if .QueryParameters}}
								<h5>Query parameters</h5>
								{{range .QueryParameters}}
									{{template "argument" .}}
								{{end}}
							{{end}}
							{{if .Response}}
								<h5>Response</h5>
								{{template "argument" .Response}}
							{{end}}
							{{if .Middleware}}
								<h5>Middlewares</h5>
								<ol>
									{{range .Middleware}}
										<li>
											<strong>{{.Name}}</strong><br/>
											{{.Description}}
										</li>
									{{end}}
								</ol>
							{{end}}
						</div>
					</div>
				{{end}}
			{{end}}

			<hr>

			Generated {{Now}}
		</main>
	</div>
	</body>
</html>
{{end}}
{{define "argument"}}
	<div class="argument">
		{{if .Name}}
			<label>Name:</label>
			<p>{{.Name}}</p>
		{{end}}
		{{if HasField . "ContentType"}}
			<label>Content type:</label>
			<p>{{.ContentType}}</p>
		{{end}}
		{{if .Example}}
			<label>Example:</label>
			<p>
				<strong>JSON</strong>
				<pre>{{MarshalJSON .Example}}</pre>
			</p>
			<p>
				<strong>Go type</strong>
				<pre>{{PrettyPrint .Example}}</pre>
			</p>
		{{end}}
	</div>
{{end}}
`))
