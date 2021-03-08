package dependencies

import (
	"andrewwormald/apollo/users"
	"database/sql"
)

// Injector defines the getter methods for obtaining the Dependency attributes
// which are generally client configurations.
type Injector interface {
	Users() users.API
	MainDB() *sql.DB
}

// Dependencies is the instance that will hold all dependency configurations such as db connections,
// http clients, logical clients etc.
type Dependencies struct {
	users users.API
	maindb *sql.DB
}

func (c *Dependencies) Users() users.API {
	return c.users
}

func (c *Dependencies) MainDB() *sql.DB {
	return c.maindb
}


var _ Injector = (*Dependencies)(nil)
