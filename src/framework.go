package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/luno/jettison/errors"

	"gomicro/config"
	"gomicro/templates"
)
// {{physical service}}
//	+-->main.go
//		|
//		|
//		+--> {{logical service}}
//			|
//			+--> client/
//				+--> locale/
//				+--> http/
//			|
//			+--> server/
//				+--> http/
//			|
//			***+--> db/ (under consideration)***
//			|
//			+--> dependencies/
//			|
//			+--> api.go

// CreateFrameworkWithFillInStrategy takes the approach that if the file exists it will not touch that file
// If the file does not exist then it will create it to ensure the project has all the required directories and files
func CreateFrameworkWithFillInStrategy(path string, c *config.Config) error {
	// 1. Create final destination if not exists
	err := CreateDirIfNotExists(path)
	if err != nil {
		return errors.Wrap(err, "")
	}

	// 2. Change to the final destination
	err = os.Chdir(path)
	if err != nil {
		return errors.Wrap(err, "")
	}

	// 3. Create physical service if it doesnt exist
	err = CreateDirIfNotExists(c.Service.Name)
	if err != nil {
		return errors.Wrap(err, "")
	}

	// 4. Ensure physical service has a main.go file
	err = CreateFileIfNotExists( c.Service.Name + "/main.go", templates.FileConfig{
		&templates.PackageHeader{Name: "main"},
		&templates.Function{Name: "main"},
	})
	if err != nil {
		return errors.Wrap(err, "")
	}

	// 5. Create a module file if does not exist and call go mod tidy
	if c.Module != "" {
		exists, err := FileExists("go.mod")
		if err != nil {
			return errors.Wrap(err, "")
		}

		if !exists {
			init := []string{"mod", "init", c.Module}
			cmd := exec.Command("go", init...)
			_, err := cmd.Output()
			if err != nil {
				return errors.Wrap(err, "")
			}
		}

		tidy := []string{"mod", "tidy"}
		cmd := exec.Command("go", tidy...)
		_, err = cmd.Output()
		if err != nil {
			return errors.Wrap(err, "failed run `go mod tidy`")
		}
	}

	// Change to the physicals directory
	err = os.Chdir(c.Service.Name)
	if err != nil {
		return errors.Wrap(err, "")
	}

	// 6. Build out each logical
	for _, logical := range c.Service.Logicals {
		// 6.1 Ensure all logicals exist or are created
		err = CreateDirIfNotExists(logical.Name)
		if err != nil {
			return errors.Wrap(err, "")
		}

		// 6.3 Do not build an api if not configured
		if !logical.API.Implementations.HTTP &&
			!logical.API.Implementations.Local {
			continue
		}

		if logical.API.FileName == "" {
			logical.API.FileName = "api.go"
		}

		if logical.API.InterfaceName == "" {
			logical.API.InterfaceName = "API"
		}

		// Ensure a basic example api is built for WireUp
		apiFilePath := logical.Name + "/" + logical.API.FileName
		err = CreateFileIfNotExists(apiFilePath, templates.FileConfig{
			&templates.PackageHeader{Name: logical.Name},
			&templates.Imports{Values: []string{"context"}},
			&templates.Interface{Name: logical.API.InterfaceName, Functions: []string{
				"Ping(ctx context.Context) error",
			}},
		})
		if err != nil {
			return errors.Wrap(err, "")
		}
	}

	err = os.Chdir("..")
	if err != nil {
		return errors.Wrap(err, "")
	}

	return nil
}

func CreateLogicalRuntimeSetup(path string, c *config.Config) error {
	err := os.Chdir(path + "/" + c.Service.Name)
	if err != nil {
		return errors.Wrap(err, "")
	}

	for _, log := range c.Service.Logicals {
		fileName := log.Name + "/" + log.Name + ".go"
		err = CreateFileIfNotExists(fileName, nil)
		if err != nil {
			return errors.Wrap(err, "")
		}

		// Reset file
		err := os.Truncate(fileName, 0)
		if err != nil {
			return errors.Wrap(err, "")
		}

		f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			return errors.Wrap(err, "")
		}

		ph := templates.PackageHeader{
			Name: log.Name,
		}
		err = ph.AddTo(f)
		if err != nil {
			return errors.Wrap(err, "")
		}

		imp := templates.Imports{
			Values: []string{
				strings.Join([]string{c.Module, c.Service.Name, log.Name, "dependencies"}, "/"),
				strings.Join([]string{c.Module, c.Service.Name, log.Name, "server"}, "/"),
			},
		}
		err = imp.AddTo(f)
		if err != nil {
			return errors.Wrap(err, "")
		}

		rs := templates.RuntimeSetup{}
		err = rs.AddTo(f)
		if err != nil {
			return errors.Wrap(err, "")
		}
	}

	return nil
}