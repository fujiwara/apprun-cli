package cli

import (
	"context"
	"fmt"
)

type StatusOption struct {
}

func (c *CLI) runStatus(ctx context.Context) error {
	app, err := c.LoadApplication(ctx, c.Application)
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
