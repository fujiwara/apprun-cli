package cli

import (
	"context"

	"github.com/sacloud/apprun-api-go"
)

func New(ctx context.Context) (*CLI, error) {
	c := &CLI{
		client: &apprun.Client{},
	}
	return c, nil
}
