package cli

import (
	"context"
	"fmt"
)

type StatusOption struct {
	Application string `arg:"" help:"Name of the definition file to status" required:""`
}

func (c *CLI) runStatus(ctx context.Context) error {
	opt := c.Status
	app, err := LoadApplication(ctx, opt.Application)
	if err != nil {
		return fmt.Errorf("failed to load application: %w", err)
	}
	info, _, err := c.getApplicationByName(ctx, app.Name)
	if err != nil {
		return fmt.Errorf("failed to get application: %w", err)
	}
	fmt.Println(toJSONIndent(info))
	return nil
}
