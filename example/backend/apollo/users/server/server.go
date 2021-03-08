// GoMicro expects a struct type called Server to exist and is dependant on it.
// You may edit this file and will not be overwritten.
package server

import (
	"andrewwormald/apollo/users"
	"andrewwormald/apollo/users/dependencies"
)

type Server struct {
	Dependencies dependencies.Dependencies
}

var _ users.API = (*Server)(nil)
