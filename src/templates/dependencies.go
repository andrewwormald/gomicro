package templates

import (
	"os"
	"text/template"
)

type DependencyTemplate struct {
	Deps []Dependency
}

type Dependency struct {
	GetterName string
	VariableName string
	ImportedType string
}

func (dt *DependencyTemplate) AddTo(file *os.File) error {
	return template.Must(template.New("").Parse(dependenciesTemplate)).Execute(file, dt)
}

var dependenciesTemplate = `
// Injector defines the getter methods for obtaining the Dependency attributes
// which are generally client configurations.
type Injector interface {
{{- range $key, $value := .Deps }}
	{{ $value.GetterName }}() {{ $value.ImportedType }}
{{- end }}
}

// Dependencies is the instance that will hold all dependency configurations such as db connections,
// http clients, logical clients etc.
type Dependencies struct {
{{- range $key, $value := .Deps }}
	{{ $value.VariableName }} {{ $value.ImportedType }}
{{- end }}
}
{{ range $key, $value := .Deps }}
func (c *Dependencies) {{ $value.GetterName }}() {{ $value.ImportedType }} {
	return c.{{ $value.VariableName }}
}
{{ end }}

var _ Injector = (*Dependencies)(nil)
`