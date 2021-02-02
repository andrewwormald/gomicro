package main

import (
	"gomicro/config"
	"gomicro/reader"
	"gomicro/templates"
	"os"
	"strings"
)

// WireUpDependencies attempts to setup and inject inter-service dependencies
func WireUpDependencies() {}

// WireUpHttpClientServer takes an interface as an http API and generates a connected http client and server
func WireUpHttpClientServer(path string, c *config.Config) error {
	err := os.Chdir(path + "/" + c.Service.Name)
	if err != nil {
		return err
	}

	for _, logical := range c.Service.Logicals {
		if logical.API.FileName == "" {
			continue
		}

		path := logical.Name + "/" + logical.API.FileName
		fs, err := reader.ReadAPI(logical.Name, path)
		if err != nil {
			return err
		}

		err = CreateServerHandlers(c, logical, fs)
		if err != nil {
			return err
		}

		err = CreateHttpClientImpl(c, logical, fs)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateServerHandlers(c *config.Config, logical config.Logical, fs []reader.FunctionSignature) error {
	serverFilePath := logical.Name + "/server/server.go"
	// Reset server file
	err := os.Truncate(serverFilePath, 0)
	if err != nil {
		return err
	}

	serverFile, err := os.OpenFile(serverFilePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}

	packageHeader := templates.PackageHeader{
		Name: "server",
	}

	err = packageHeader.AddTo(serverFile)
	if err != nil {
		return err
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
		return err
	}

	for _, method := range fs {

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
			return err
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
			return err
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
			return err
		}
	}

	return nil
}

func CreateHttpClientImpl(c *config.Config, logical config.Logical, fs []reader.FunctionSignature) error {
	filePath := logical.Name + "/client/client.go"
	// Reset server file
	err := os.Truncate(filePath, 0)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}

	packageHeader := templates.PackageHeader{
		Name: "client",
	}

	err = packageHeader.AddTo(file)
	if err != nil {
		return err
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
		return err
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

		// Add client type
		client := templates.Struct{
			Name: "HttpClient",
			Fields: map[string]string{
				"cl":      "*http.Client",
				"address": "string",
			},
		}

		err = client.AddTo(file)
		if err != nil {
			return err
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
			return err
		}
	}

	APIEnforcer := templates.Statement{
		Value: "var _ " + logical.Name + "." + logical.API.InterfaceName + " = (*HttpClient)(nil)",
	}

	err = APIEnforcer.AddTo(file)
	if err != nil {
		return err
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
