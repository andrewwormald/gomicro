package logical

import (
	"context"
	
	"andrewwormald/apollo/bookings"
	"andrewwormald/apollo/bookings/server"
)

func New() bookings.Client {
	return &Client{
		ServerImpl: &server.Server{},
	}
}

type Client struct {
	ServerImpl bookings.Client
}

func (cl *Client) Ping(ctx context.Context) (error) {
	return cl.ServerImpl.Ping(ctx)
}
var _ bookings.Client = (*Client)(nil)
