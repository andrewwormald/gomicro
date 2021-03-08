package main

import (
	"os"
	"strings"

	"github.com/luno/jettison/errors"

	"gomicro/config"
	"gomicro/reader"
	"gomicro/templates"
)

// WireUpDependencies attempts to setup and inject inter-service dependencies
func WireUpDependencies(path string, c *config.Config) error {
	err := os.Chdir(path + "/" + c.Service.Name)
	if err != nil {
		return errors.Wrap(err, "")
	}

	for _, log := range c.Service.Logicals {
		// 6.2 Ensure that dependencies are built out
		err = CreateDirIfNotExists(log.Name + "/" + "dependencies")
		if err != nil {
			return errors.Wrap(err, "")
		}

		fileName := log.Name + "/dependencies/dependencies.go"
		err = CreateFileIfNotExists(fileName, nil)
		if err != nil {
			return errors.Wrap(err, "")
		}

		// Reset file
		err := os.Truncate(fileName, 0)
		if err != nil {
			return errors.Wrap(err, "")
		}

		serverFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			return errors.Wrap(err, "")
		}

		var deps []templates.Dependency
		var imports []string
		for _, d := range log.Dependencies {
			deps = append(deps, templates.Dependency{
				GetterName:   d.Name,
				VariableName: strings.ToLower(d.Name),
				ImportedType: d.Type,
			})

			imports = append(imports, d.Path)
		}

		p := templates.PackageHeader{Name: "dependencies"}
		err = p.AddTo(serverFile)
		if err != nil {
			return errors.Wrap(err, "")
		}

		imps := templates.Imports{Values: imports}
		err = imps.AddTo(serverFile)
		if err != nil {
			return errors.Wrap(err, "")
		}

		dt := templates.DependencyTemplate{
			Deps: deps,
		}

		err = dt.AddTo(serverFile)
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

// WireUpHttpClientServer takes an interface as an http API and generates a connected http client and server
func WireUpHttpClientServer(path string, c *config.Config) error {
	err := os.Chdir(path + "/" + c.Service.Name)
	if err != nil {
		return errors.Wrap(err, "")
	}

	for _, logical := range c.Service.Logicals {
		if logical.API.FileName == "" {
			continue
		}

		path := logical.Name + "/" + logical.API.FileName
		fs, err := reader.ReadAPI(logical.Name, path)
		if err != nil {
			return errors.Wrap(err, "")
		}

		err = CreateServerImpl(c, logical, fs)
		if err != nil {
			return errors.Wrap(err, "")
		}

		err = CreateHttpClientImpl(c, logical, fs)
		if err != nil {
			return errors.Wrap(err, "")
		}

		err = CreateLogicalClientImpl(c, logical, fs)
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

func CreateServerImpl(c *config.Config, logical config.Logical, fs []reader.FunctionSignature) error {
	err := CreateDirIfNotExists(logical.Name + "/server")
	if err != nil {
		return errors.Wrap(err, "")
	}

	fileName := logical.Name + "/server/server_gen.go"
	err = CreateFileIfNotExists(fileName, templates.FileConfig{
		&templates.PackageHeader{Name: "server"},
	})
	if err != nil {
		return errors.Wrap(err, "")
	}

	fileName = logical.Name + "/server/server.go"
	err = CreateFileIfNotExists(fileName, templates.FileConfig{
		&templates.Statement{Value: "// GoMicro expects a struct type called Server to exist and is dependant on it."},
		&templates.Statement{Value: "// You may edit this file and will not be overwritten."},
		&templates.PackageHeader{Name: "server"},
		&templates.Imports{
			Values: []string{
				c.Module + "/" + c.Service.Name + "/" + logical.Name,
				c.Module + "/" + c.Service.Name + "/" + logical.Name + "/dependencies",
			},
		},
		&templates.Struct{
			Name: "Server",
			Fields: map[string]string{
				"Dependencies": "dependencies.Dependencies",
			},
		},
		new(templates.Linebreak),
		&templates.Statement{Value: "var _ " + logical.Name + "." + logical.API.InterfaceName + " = (*Server)(nil)"},
	})
	if err != nil {
		return errors.Wrap(err, "")
	}

	serverFilePath := logical.Name + "/server/server_gen.go"

	// Reset server file
	err = os.Truncate(serverFilePath, 0)
	if err != nil {
		return errors.Wrap(err, "")
	}

	serverFile, err := os.OpenFile(serverFilePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return errors.Wrap(err, "")
	}

	packageHeader := templates.PackageHeader{
		Name: "server",
	}

	err = packageHeader.AddTo(serverFile)
	if err != nil {
		return errors.Wrap(err, "")
	}

	// Add required imports to file
	imps := templates.Imports{
		Values: []string{
			"encoding/json",
			"io/ioutil",
			"net/http",
			templates.SingleLineSpace.String(),
			c.Module + "/" + c.Service.Name + "/" + logical.Name,
		},
	}
	err = imps.AddTo(serverFile)
	if err != nil {
		return errors.Wrap(err, "")
	}

	htr := templates.HttpRegister{
		API: strings.Join([]string{logical.Name, logical.API.InterfaceName}, "."),
	}

	for _, method := range fs {
		htr.Handlers = append(htr.Handlers, templates.Handler{
			URI: strings.Join([]string{strings.ToLower(logical.Name), strings.ToLower(method.Name)}, "/"),
			Method: method.Name,
		})

		// Create request type
		requestFields := make(map[string]string)
		var params []string
		for _, v := range method.Params {
			if v.ImportType == "context.Context" {
				continue
			}
			v.Name = ExportiseName(v.Name)
			requestFields[v.Name] = v.ImportType
			params = append(params, v.Name)
		}
		req := templates.Struct{
			Name:   method.Name + "Request",
			Fields: requestFields,
		}

		err = req.AddTo(serverFile)
		if err != nil {
			return errors.Wrap(err, "")
		}

		// Create response type
		responseFields := make(map[string]string)
		var responseParams []string
		var results []string
		for _, v := range method.Results {
			// Do not include errors in the response type but do in the results
			if v.ImportType == "error" {
				results = append(results, "err")
				continue
			}

			if v.Name == "" {
				sp := strings.Split(v.ImportType, "")
				if len(sp) > 2 {
					sp = sp[:3]
				} else if len(sp) == 2{
					sp = sp[:2]
				}

				v.Name = strings.Join(sp, "")
			}

			results = append(results, v.Name)
			nameExported := ExportiseName(v.Name)
			responseParams = append(responseParams, nameExported)
			responseFields[nameExported] = v.ImportType
		}

		// Error is always returned last so ignore this one. Definitely hacky
		responseParams = append(responseParams, "_")
		resp := templates.Struct{
			Name:   method.Name + "Response",
			Fields: responseFields,
		}

		err = resp.AddTo(serverFile)
		if err != nil {
			return errors.Wrap(err, "")
		}

		// Add handlers to file
		h := templates.HttpHandler{
			Method:         method.Name,
			API:            strings.Join([]string{logical.Name, logical.API.InterfaceName}, "."),
			RequestType:    req.Name,
			Params:         params,
			Results:        results,
			ResponseType:   resp.Name,
			ResponseParams: responseParams,
		}

		err = h.AddTo(serverFile)
		if err != nil {
			return errors.Wrap(err, "")
		}
	}

	err = htr.AddTo(serverFile)
	if err != nil {
		return errors.Wrap(err, "")
	}

	return nil
}

func CreateHttpClientImpl(c *config.Config, logical config.Logical, fs []reader.FunctionSignature) error {
	err := CreateDirIfNotExists(logical.Name + "/client")
	if err != nil {
		return errors.Wrap(err, "")
	}

	err = CreateDirIfNotExists(logical.Name + "/client/http")
	if err != nil {
		return errors.Wrap(err, "")
	}

	filePath := logical.Name + "/client/http/client_gen.go"
	err = CreateFileIfNotExists(filePath, templates.FileConfig{
		&templates.PackageHeader{Name: "http"},
	})
	if err != nil {
		return errors.Wrap(err, "")
	}

	// Reset server file
	err = os.Truncate(filePath, 0)
	if err != nil {
		return errors.Wrap(err, "")
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return errors.Wrap(err, "")
	}

	packageHeader := templates.PackageHeader{
		Name: "http",
	}

	err = packageHeader.AddTo(file)
	if err != nil {
		return errors.Wrap(err, "")
	}

	// Add required imports to file
	imps := templates.Imports{
		Values: []string{
			"bytes",
			"context",
			"encoding/json",
			"io/ioutil",
			"net/http",
			templates.SingleLineSpace.String(),
			"golang.org/x/net/context/ctxhttp",
			templates.SingleLineSpace.String(),
			c.Module + "/" + c.Service.Name + "/" + logical.Name,
			c.Module + "/" + c.Service.Name + "/" + logical.Name + "/" + "server",
		},
	}
	err = imps.AddTo(file)
	if err != nil {
		return errors.Wrap(err, "")
	}

	// Add client type
	client := templates.HttpClientType{
		API: strings.Join([]string{logical.Name, logical.API.InterfaceName}, "."),
	}

	err = client.AddTo(file)
	if err != nil {
		return errors.Wrap(err, "")
	}

	for _, method := range fs {
		// Create request type
		var params []string
		requestObject := make(map[string]string)
		for _, v := range method.Params {
			if v.ImportType == "context.Context" {
				continue
			}

			params = append(params, v.Name+" "+v.ImportType)

			exportedName := ExportiseName(v.Name)
			requestObject[exportedName] = v.Name
		}

		// Create response type
		var responseList []string
		var results []string
		var returnList []string
		for _, v := range method.Results {
			returnList = append(returnList, v.Name)

			// Do not include errors in the response type but do in the results
			if v.ImportType == "error" {
				results = append(results, "err error")
				continue
			}

			results = append(results, v.Name+" "+v.ImportType)
			responseList = append(responseList, ExportiseName(v.Name))
		}

		// Add client methods to fulfill API spec
		h := templates.HttpClient{
			Method:         method.Name,
			Service:        logical.Name,
			RequestType:    "server." + method.Name + "Request",
			ResponseType:   "server." + method.Name + "Response",
			Params:         params,
			ResponseParams: responseList,
			Request:        requestObject,
			Results:        results,
			Return:         returnList,
		}

		err = h.AddTo(file)
		if err != nil {
			return errors.Wrap(err, "")
		}
	}

	APIEnforcer := templates.Statement{
		Value: "var _ " + logical.Name + "." + logical.API.InterfaceName + " = (*Client)(nil)",
	}

	err = APIEnforcer.AddTo(file)
	if err != nil {
		return errors.Wrap(err, "")
	}

	return nil
}

func CreateLogicalClientImpl(c *config.Config, logical config.Logical, fs []reader.FunctionSignature) error {
	err := CreateDirIfNotExists(logical.Name + "/client")
	if err != nil {
		return errors.Wrap(err, "")
	}

	err = CreateDirIfNotExists(logical.Name + "/client/logical")
	if err != nil {
		return errors.Wrap(err, "")
	}

	filePath := logical.Name + "/client/logical/client_gen.go"
	err = CreateFileIfNotExists(filePath, templates.FileConfig{
		&templates.PackageHeader{Name: "logical"},
	})
	if err != nil {
		return errors.Wrap(err, "")
	}

	// Reset server file
	err = os.Truncate(filePath, 0)
	if err != nil {
		return errors.Wrap(err, "")
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return errors.Wrap(err, "")
	}

	packageHeader := templates.PackageHeader{
		Name: "logical",
	}

	err = packageHeader.AddTo(file)
	if err != nil {
		return errors.Wrap(err, "")
	}

	// Add required imports to file
	imps := templates.Imports{
		Values: []string{
			"context",
			templates.SingleLineSpace.String(),
			c.Module + "/" + c.Service.Name + "/" + logical.Name,
			c.Module + "/" + c.Service.Name + "/" + logical.Name + "/" + "server",
		},
	}
	err = imps.AddTo(file)
	if err != nil {
		return errors.Wrap(err, "")
	}

	// Add client type
	client := templates.LogicalClientType{
		API: strings.Join([]string{logical.Name, logical.API.InterfaceName}, "."),
	}

	err = client.AddTo(file)
	if err != nil {
		return errors.Wrap(err, "")
	}

	for _, method := range fs {
		var params []string
		var inlineParams []string
		for _, v := range method.Params {
			if v.ImportType == "context.Context" {
				continue
			}

			params = append(params, v.Name+" "+v.ImportType)
			inlineParams = append(inlineParams, v.Name)
		}

		var results []string
		for _, v := range method.Results {
			results = append(results, v.ImportType)
		}

		logicalMethod := &templates.LogicalClientTemplate{
			Method:       method.Name,
			Params:       params,
			InlineParams: inlineParams,
			Results:      results,
		}

		err = logicalMethod.AddTo(file)
		if err != nil {
			return errors.Wrap(err, "")
		}
	}

	APIEnforcer := templates.Statement{
		Value: "var _ " + logical.Name + "." + logical.API.InterfaceName + " = (*Client)(nil)",
	}

	err = APIEnforcer.AddTo(file)
	if err != nil {
		return errors.Wrap(err, "")
	}

	return nil
}

func ExportiseName(s string) string {
	if s == "" {
		return ""
	}
	breakdown := strings.Split(s, "")
	breakdown[0] = strings.ToUpper(breakdown[0])
	return strings.Join(breakdown, "")
}
