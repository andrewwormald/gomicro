package logical

import (
	"context"
	
	"andrewwormald/apollo/users"
	"andrewwormald/apollo/users/server"
)

func New() users.API {
	return &Client{
		ServerImpl: &server.Server{},
	}
}

type Client struct {
	ServerImpl users.API
}

func (cl *Client) Ping(ctx context.Context) (error) {
	return cl.ServerImpl.Ping(ctx)
}
var _ users.API = (*Client)(nil)
