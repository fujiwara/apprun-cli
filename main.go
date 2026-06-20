package cli

import (
	"context"
	"fmt"

	armed "github.com/fujiwara/jsonnet-armed"
	apprun "github.com/sacloud/sacloud-sdk-go/api/apprun"
	"github.com/sacloud/sacloud-sdk-go/common/saclient"
)

func New(ctx context.Context) (*CLI, error) {
	var sc saclient.Client
	if err := sc.Populate(); err != nil {
		return nil, fmt.Errorf("failed to populate sakura cloud client: %w", err)
	}
	client, err := apprun.NewClient(&sc)
	if err != nil {
		return nil, fmt.Errorf("failed to build apprun client: %w", err)
	}
	c := &CLI{
		client: client,
		loader: &armed.CLI{},
	}
	return c, nil
}
