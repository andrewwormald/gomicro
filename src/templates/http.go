package templates

import (
	"os"
	"text/template"
)

type HttpHandler struct {
	Method         string
	API            string
	RequestType    string
	Params         []string
	Results        []string
	ResponseType   string
	ResponseParams []string
}

func (f *HttpHandler) AddTo(file *os.File) error {
	return template.Must(template.New("").Parse(httpHandlerTemplate)).Execute(file, f)
}

var httpHandlerTemplate = `
func Handle{{.Method}}(api {{.API}}) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var req {{.RequestType}}
		err = json.Unmarshal(b, &req)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		{{ range $key, $value := .Results }}{{if $key}}, {{end}}{{ $value }}{{ end }}{{ if eq (len .Results) 1 }} = {{ end }}{{ if not (eq (len .Results) 1) }} := {{ end }}api.{{.Method}}(r.Context(), {{range $key, $value := .Params }}{{if $key}}, {{end}}req.{{ $value }}{{end }})
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var resp {{.ResponseType}}
		{{range $key, $value := .ResponseParams }}{{if $key}}, {{end}}{{if eq $value "_"}}{{else if $value}}resp.{{end}}{{$value}}{{end}} = {{ range $key, $value := .Results }}{{if $key}}, {{end}}{{ $value }}{{ end }}
	
		respBody, err := json.Marshal(resp)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		
		_, err = w.Write(respBody)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
`

type HttpClient struct {
	Service string

	Type    string
	Method  string
	Params  []string
	Results []string

	RequestType   string
	Request map[string]string

	ResponseType   string
	ResponseParams []string

	Return []string
}

func (cl *HttpClient) AddTo(file *os.File) error {
	return template.Must(template.New("").Parse(httpClientTemplate)).Execute(file, cl)
}

var httpClientTemplate = `
func (hc * HttpClient) {{.Method}}(ctx context.Context, {{ range $key, $value := .Params }}{{if $key}}, {{end}}{{ $value }}{{ end }}) ({{- range $key, $value := .Results }}{{if $key}}, {{end}}{{ $value }}{{- end }}) {
	req := {{.RequestType}} {
	{{- range $key, $value := .Request }}
		{{$key}}: {{$value}},
	{{- end}}
	}

	b, err := json.Marshal(req)
	if err != nil {
		return {{ range $key, $value := .Return }}{{if $key}}, {{end}}{{ $value }}{{ end }}
	}

	uniquePath := "/{{.Service}}/{{.Method}}" 
	buf := bytes.NewBuffer(b)
	httpResp, err := ctxhttp.Post(ctx, hc.cl, hc.address + uniquePath, "application/json", buf)
	if err != nil {
		return {{ range $key, $value := .Return }}{{if $key}}, {{end}}{{ $value }}{{ end }}
	}

	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return {{ range $key, $value := .Return }}{{if $key}}, {{end}}{{ $value }}{{ end }}
	}

	var resp {{.ResponseType}}
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return {{ range $key, $value := .Return }}{{if $key}}, {{end}}{{ $value }}{{ end }}
	}

	{{ if not (eq (len .ResponseParams) 0)}}return {{range $key, $value := .ResponseParams }}{{if $key}}, {{end}}{{if eq $value "_"}}{{else if $value}}resp.{{end}}{{$value}}{{end}}, nil{{end}}{{ if eq (len .ResponseParams) 0}}return nil{{end}}
}
`
