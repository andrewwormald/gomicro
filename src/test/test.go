package users

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context/ctxhttp"
)

// API is cool
type API interface {
	Set(ctx context.Context, name, surname string) error
}

// Rules
// All methods must have a context.Context
// All methods must at least return an error
// All types must have exported fields or they will be excluded
// Comments will be written in the implementation file from the interface

type implPinger struct {
	cl *http.Client
	address string
}

// Create a request structure based on the list of params provided
// this will exclude the context
type SetRequest struct {
	Name string
	Surname string
}

func (i *implPinger) Set(ctx context.Context, name, surname string) error {
	req := SetRequest{
		Name:   name,
		Surname: surname,
	}
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	uniquePath := "/ping"
	buf := bytes.NewBuffer(b)
	_, err = ctxhttp.Post(ctx, i.cl, i.address + uniquePath, "application/json", buf)
	return err
}

// Middleware implementation
// Generation is entirely for the middle ware. The user just needs to write a implementation of the api and pass it in
//func RegisterServer(address string, api API) {
//	http.HandleFunc("/ping", HandleSet(api))
//}

func HandleSet(api API) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusBadRequest)
		}

		var req SetRequest
		err = json.Unmarshal(body, &req)
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusBadRequest)
		}

		err = api.Set(r.Context(), req.Name, req.Surname)
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}



// User implementation
// this can be auto generated for the first time to save the user some time but wont do it once the file already exists
// to ensure we do not overwrite anything. Later on we can parse the file with AST to read what is implemented and update
// with a mock function maybe but that seems like a lot of work and little gain right now.

type Server struct {}

func (s *Server) Ping(ctx context.Context) error {
	return nil
}

func (s *Server) Set(ctx context.Context, name, surname string) error {
	return nil
}

var _ API = (*Server)(nil)













