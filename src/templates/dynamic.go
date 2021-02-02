package templates

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

type LineSpace string

func (l LineSpace) String() string {
	return string(l)
}

const (
	SingleLineSpace LineSpace = "singlelinebreak"
	DoubleLineSpace LineSpace = "doubelinebreak"
)

type FileConfig []Adder

type Adder interface {
	AddTo(*os.File) error
}

type PackageHeader struct {
	Name string
}

func (ph *PackageHeader) AddTo(file *os.File) error {
	return template.Must(template.New("").Parse(packageHeaderTemplate)).Execute(file, ph)
}

var packageHeaderTemplate = `package {{.Name}}
`

// Imports must be used as the type for rendering the imports template. It is expected to be called 'Imports'
// in the provided struct.
type Imports struct {
	Values []string
}

func (i *Imports) AddTo(file *os.File) error {
	for index, line := range i.Values {
		if line == SingleLineSpace.String() {
			i.Values[index] = ""
			continue
		}

		if line == DoubleLineSpace.String() {
			i.Values[index] = "\n"
			continue
		}

		i.Values[index] = "\"" + line + "\""
	}
	return template.Must(template.New("").Parse(importsTemplate)).Execute(file, i)
}

var importsTemplate = `
import (
{{- range $key, $value := .Values }}
	{{ $value }}
{{- end }}
)
`

// Struct is the configuration of a custom struct type to be made. The template expects 'Struct'
type Struct struct {
	Name   string
	Fields map[string]string
}

func (s *Struct) AddTo(file *os.File) error {
	return template.Must(template.New("").Parse(structTemplate)).Execute(file, s)
}

var structTemplate = `
type {{.Name}} struct {
{{- range $key, $value := .Fields }}
	{{ $key }} {{ $value }}
{{- end }}
}
`

// Interface is the configuration of a custom interface type to be made. The template expects 'Interface'
type Interface struct {
	Name      string
	Functions []string
}

func (i *Interface) AddTo(file *os.File) error {
	return template.Must(template.New("").Parse(interfaceTemplate)).Execute(file, i)
}

var interfaceTemplate = `
type {{.Name}} interface {
{{- range $key, $value := .Functions }}
	{{ $value }}
{{- end }}
}
`

type Function struct {
	Name         string
	InputParams  []string
	OutputParams []string
}

func (f *Function) AddTo(file *os.File) error {
	if len(f.OutputParams) > 1 {
		return template.Must(template.New("").Parse(multiReturnFunctionTemplate)).Execute(file, f)
	}

	return template.Must(template.New("").Parse(singleReturnFunctionTemplate)).Execute(file, f)
}

var multiReturnFunctionTemplate = `
func {{.Name}}({{- range $key, $value := .InputParams }}{{ $value }}{{- end }}) ({{- range $key, $value := .OutputParams }}{{ $value }},{{- end }}) {}
`
var singleReturnFunctionTemplate = `
func {{.Name}}({{- range $key, $value := .InputParams }}{{ $value }}{{- end }}) {{- range $key, $value := .OutputParams }} {{ $value }}{{- end }} {}
`

type Method struct {
	Name         string
	ParentStruct string
	InputParams  []string
	OutputParams []string
}

func (m *Method) AddTo(file *os.File) error {
	// Abstract the method prefix syntax i.e make (cl *Client) from "Client"
	prefix := strings.ToLower(strings.Split(m.ParentStruct, "")[0])
	m.ParentStruct = fmt.Sprintf("%s *%s", prefix, m.ParentStruct)

	if len(m.OutputParams) > 1 {
		return template.Must(template.New("").Parse(multiReturnMethodTemplate)).Execute(file, m)
	}

	return template.Must(template.New("").Parse(singleReturnMethodTemplate)).Execute(file, m)
}

var multiReturnMethodTemplate = `
func ({{.ParentStruct}}) {{.Name}}({{- range $key, $value := .InputParams }}{{ $value }}{{- end }}) ({{- range $key, $value := .OutputParams }}{{ $value }}, {{- end }}) {}
`
var singleReturnMethodTemplate = `
func ({{.ParentStruct}}) {{.Name}}({{- range $key, $value := .InputParams }}{{ $value }}{{- end }}) {{- range $key, $value := .OutputParams }} {{ $value }}{{- end }} {}
`

type Statement struct {
	Value string
}

func (s *Statement) AddTo(file *os.File) error {
	return template.Must(template.New("").Parse(statementTemplate)).Execute(file, s)
}

var statementTemplate = `
{{.Value}}
`
