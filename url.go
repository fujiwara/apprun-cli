package cli

import (
	"context"
	"fmt"
)

func (c *CLI) runURL(ctx context.Context) error {
	app, err := c.LoadApplication(ctx, c.Application)
	if err != nil {
		return fmt.Errorf("failed to load application: %w", err)
	}
	info, _, err := c.getApplicationByName(ctx, app.Name)
	if err != nil {
		return fmt.Errorf("failed to get application: %w", err)
	}
	fmt.Println(info.PublicUrl)
	return nil
}
