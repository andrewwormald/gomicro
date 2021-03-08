package templates

import (
	"os"
	"text/template"
)

type RuntimeSetup struct {}

func (f *RuntimeSetup) AddTo(file *os.File) error {
	return template.Must(template.New("").Parse(runtimeSetup)).Execute(file, f)
}

var runtimeSetup = `
func Run(d *dependencies.Dependencies) error {
	serverImpl := &server.Server{
		Dependencies: d,
	}
	server.RegisterHandlers(serverImpl)

	return nil
}
`
