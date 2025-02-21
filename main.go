package cli

import (
	"context"

	"github.com/google/go-jsonnet"
	"github.com/sacloud/apprun-api-go"
)

var Version string

func New(ctx context.Context) (*CLI, error) {
	c := &CLI{
		client: &apprun.Client{},
		vm:     jsonnet.MakeVM(),
	}
	return c, nil
}
