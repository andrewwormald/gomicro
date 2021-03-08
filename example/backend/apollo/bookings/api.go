package bookings

import (
	"context"
)

type Client interface {
	Ping(ctx context.Context) error
}
