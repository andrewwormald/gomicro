package users

import (
	"context"
)

type User struct {
	ID int64
	Email string
}

// API is cool
type API interface {
	Set(ctx context.Context, u User) (userID int64, err error)
}
