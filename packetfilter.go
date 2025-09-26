package cli

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sacloud/apprun-api-go"
	v1 "github.com/sacloud/apprun-api-go/apis/v1"
)

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

func (c *CLI) updatePacketFilter(ctx context.Context, appID string, pf v1.PatchPacketFilter) error {
	if pf.IsEnabled == nil {
		pf.IsEnabled = ptr(false)
	}
	if pf.Settings == nil {
		pf.Settings = &[]v1.PacketFilterSetting{}
	}
	slog.Debug("updating packet filter", "patch", toJSON(pf))

	pfOp := apprun.NewPacketFilterOp(c.client)
	if res, err := pfOp.Update(ctx, appID, &pf); err != nil {
		return fmt.Errorf("failed to update packet filter: %s", err)
	} else {
		slog.Info("updated packet filter", "result", toJSON(res))
	}
	return nil
}
