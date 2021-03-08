package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	
	"golang.org/x/net/context/ctxhttp"
	
	"andrewwormald/apollo/users"
	"andrewwormald/apollo/users/server"
)

func New(serverAddress string) users.API {
	return &Client{
		Address: serverAddress,
		HttpClient: &http.Client{},
	}
}

type Client struct {
	Address string
	HttpClient *http.Client
}

func (c * Client) Ping(ctx context.Context) (err error) {
	req := server.PingRequest {
	}

	b, err := json.Marshal(req)
	if err != nil {
		return 
	}

	uniquePath := "/users/Ping"
	buf := bytes.NewBuffer(b)
	httpResp, err := ctxhttp.Post(ctx, c.HttpClient, c.Address + uniquePath, "application/json", buf)
	if err != nil {
		return 
	}

	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return 
	}

	var resp server.PingResponse
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return 
	}

	return nil
}
var _ users.API = (*Client)(nil)
