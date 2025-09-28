package cli

import (
	"context"

	armed "github.com/fujiwara/jsonnet-armed"
	"github.com/sacloud/apprun-api-go"
)

func New(ctx context.Context) (*CLI, error) {
	c := &CLI{
		client: &apprun.Client{},
		loader: &armed.CLI{},
	}
	return c, nil
}
