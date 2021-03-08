package dependencies

import (
)

// Injector defines the getter methods for obtaining the Dependency attributes
// which are generally client configurations.
type Injector interface {
}

// Dependencies is the instance that will hold all dependency configurations such as db connections,
// http clients, logical clients etc.
type Dependencies struct {
}


var _ Injector = (*Dependencies)(nil)
