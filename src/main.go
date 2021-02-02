package main

import (
	"errors"
	"flag"
	"gomicro/config"
	"gomicro/templates"
	"io/ioutil"
	"os"
)

var configPath = flag.String("config", "../example/config.yaml", "The location of the GoMicro config yaml file")
var outputPath = flag.String("output", "../myRepo", "The directory to generate the services or update them")

// main scenario 1:
// Nothing exists and a boilerplate must be built
//
// main scenario 2:
// The GoMicro framework exists but needs dependency
// injection and http api impl generation.
//
// Breakdown:
// - HTTP api generation will happen first
//
// - HTTP api server and client impl will be generated
// 	using the api file configured.
//
// - Dependency injection will use the
// 	dependants client package and requires the dependant
//	to have an API.
//
// The GoMicro framework runs in three parts:
// 1. Generate all the necessary files to match the framework schema
// 2. Generate the local and http implementation of the api
// 3. Inject dependencies into the services
//
// API Rules
// All methods must have a context.Context in the first position
// All methods must always return an error
// All types must have exported fields or they will be excluded
// All parameters must have variables not just type declaration
// Comments will be written in the implementation file from the interface

func main() {
	flag.Parse()

	// get the current location that its being executed from
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// change directory to its being executed from
	err = os.Chdir(wd)
	if err != nil {
		panic(err)
	}

	c, err := config.ParseConfig(*configPath)
	if err != nil {
		panic(err)
	}

	err = CreateFrameworkWithFillInStrategy(*outputPath, c)
	if err != nil {
		panic(err)
	}

	err = WireUpHttpClientServer(*outputPath, c)
	if err != nil {
		panic(err)
	}
}

func CreateFileIfNotExists(path string, fc templates.FileConfig) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = ioutil.WriteFile(path, []byte{}, os.ModePerm)
		if err != nil {
			return errors.New("unable to create file")
		}

		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			return err
		}

		for _, adder := range fc {
			err := adder.AddTo(f)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func CreateDirIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return errors.New("unable to create service directory")
		}
	}

	return nil
}
