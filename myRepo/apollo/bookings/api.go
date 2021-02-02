package bookings

import (
	"context"
)

type API interface {
	Ping(ctx context.Context) error
}
