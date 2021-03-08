// GoMicro expects a struct type called Server to exist and is dependant on it.
// You may edit this file and will not be overwritten.
package server

import (
	"andrewwormald/apollo/bookings"
	"andrewwormald/apollo/bookings/dependencies"
)

type Server struct {
	Dependencies dependencies.Dependencies
}

var _ bookings.Client = (*Server)(nil)
