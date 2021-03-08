package bookings

import (
	"andrewwormald/apollo/bookings/dependencies"
	"andrewwormald/apollo/bookings/server"
)

func Run(d *dependencies.Dependencies) error {
	serverImpl := &server.Server{
		Dependencies: d,
	}
	server.RegisterHandlers(serverImpl)

	return nil
}
