package cli

import (
	"context"
	"fmt"

	"github.com/sacloud/apprun-api-go"
	v1 "github.com/sacloud/apprun-api-go/apis/v1"
)

type PacketFilterOption struct {
}

func (c *CLI) runPacketFilter(ctx context.Context) error {
	return nil
}

func (c *CLI) getPacketFilter(ctx context.Context, appID string) (*v1.PatchPacketFilter, error) {
	// packet filter
	pfOp := apprun.NewPacketFilterOp(c.client)
	pf, err := pfOp.Read(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet filter: %s", err)
	}
	return &v1.PatchPacketFilter{
		IsEnabled: &pf.IsEnabled,
		Settings:  &pf.Settings,
	}, nil
}
