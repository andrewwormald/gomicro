package client

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

type HttpClient struct {
	address string
	cl *http.Client
}

func (hc * HttpClient) Set(ctx context.Context, u users.User) (userID int64, err error) {
	req := server.SetRequest {
		U: u,
	}

	b, err := json.Marshal(req)
	if err != nil {
		return userID, err
	}

	uniquePath := "/users/Set" 
	buf := bytes.NewBuffer(b)
	httpResp, err := ctxhttp.Post(ctx, hc.cl, hc.address + uniquePath, "application/json", buf)
	if err != nil {
		return userID, err
	}

	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return userID, err
	}

	var resp server.SetResponse
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return userID, err
	}

	return resp.UserID, nil
}

var _ users.API = (*HttpClient)(nil)
