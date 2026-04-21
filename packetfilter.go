package cli

import (
	"context"
	"fmt"
	"log/slog"

	apprun "github.com/sacloud/apprun-api-go"
	v1 "github.com/sacloud/apprun-api-go/apis/v1"
)

func (c *CLI) getPacketFilter(ctx context.Context, appID string) (*v1.PatchPacketFilter, error) {
	// packet filter
	pfOp := apprun.NewPacketFilterOp(c.client)
	pf, err := pfOp.Read(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet filter: %s", err)
	}
	settings := make([]v1.PatchPacketFilterSettingsItem, 0, len(pf.Settings))
	for _, s := range pf.Settings {
		settings = append(settings, v1.PatchPacketFilterSettingsItem{
			FromIP:             s.FromIP,
			FromIPPrefixLength: s.FromIPPrefixLength,
		})
	}
	return &v1.PatchPacketFilter{
		IsEnabled: v1.NewOptBool(pf.IsEnabled),
		Settings:  settings,
	}, nil
}

func (c *CLI) updatePacketFilter(ctx context.Context, appID string, pf v1.PatchPacketFilter) error {
	if !pf.IsEnabled.Set {
		pf.IsEnabled = v1.NewOptBool(false)
	}
	if pf.Settings == nil {
		pf.Settings = []v1.PatchPacketFilterSettingsItem{}
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
