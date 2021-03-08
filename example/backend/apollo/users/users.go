package users

import (
	"andrewwormald/apollo/users/dependencies"
	"andrewwormald/apollo/users/server"
)

func Run(d *dependencies.Dependencies) error {
	serverImpl := &server.Server{
		Dependencies: d,
	}
	server.RegisterHandlers(serverImpl)

	return nil
}
