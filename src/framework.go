package main

import (
	"gomicro/config"
	"gomicro/templates"
	"os"
	"os/exec"
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
		return err
	}

	// 2. Change to the final destination
	err = os.Chdir(path)
	if err != nil {
		return err
	}

	// 3. Create physical service if it doesnt exist
	err = CreateDirIfNotExists(c.Service.Name)
	if err != nil {
		return err
	}

	// 4. Ensure physical service has a main.go file
	err = CreateFileIfNotExists( c.Service.Name + "/main.go", templates.FileConfig{
		&templates.PackageHeader{Name: "main"},
		&templates.Imports{Values: []string{"context"}},
		&templates.Function{Name: "main"},
	})
	if err != nil {
		return err
	}

	// 5. Create a module file if does not exist and call go mod tidy
	if c.Module != "" {
		exists, err := FileExists("go.mod")
		if err != nil {
			return err
		}

		if !exists {
			init := []string{"mod", "init", c.Module}
			cmd := exec.Command("go", init...)
			_, err := cmd.Output()
			if err != nil {
				return err
			}
		}

		tidy := []string{"mod", "tidy"}
		cmd := exec.Command("go", tidy...)
		_, err = cmd.Output()
		if err != nil {
			return err
		}
	}


	// Change to the physicals directory
	err = os.Chdir(c.Service.Name)
	if err != nil {
		return err
	}

	// 6. Build out each logical
	for _, logical := range c.Service.Logicals {
		// 6.1 Ensure all logicals exist or are created
		err = CreateDirIfNotExists(logical.Name)
		if err != nil {
			return err
		}

		// 6.2 Ensure logical has an api
		if logical.API.FileName == "" {
			logical.API.FileName = "api.go"
		}

		if logical.API.InterfaceName == "" {
			logical.API.InterfaceName = "API"
		}

		// 6.2 Ensure all logicals that have an api config have an api.go file
		apiFilePath := logical.Name + "/" + logical.API.FileName
		err = CreateFileIfNotExists(apiFilePath, templates.FileConfig{
			&templates.PackageHeader{Name: logical.Name},
			&templates.Imports{Values: []string{"context"}},
			&templates.Interface{Name: logical.API.InterfaceName, Functions: []string{
				"Ping(ctx context.Context) error",
			}},
		})
		if err != nil {
			return err
		}

		// 6.4 Ensure logical has a db, client, server, and dependencies packages
		d := []string{"client", "server", "dependencies"}
		for _, dir := range d {
			err = CreateDirIfNotExists(logical.Name + "/" + dir)
			if err != nil {
				return err
			}

			fileName := logical.Name + "/" + dir + "/" + dir+".go"
			err = CreateFileIfNotExists(fileName, templates.FileConfig{
				&templates.PackageHeader{Name: dir},
			})
			if err != nil {
				return err
			}
		}
	}

	err = os.Chdir("..")
	if err != nil {
		return err
	}

	return nil
}
