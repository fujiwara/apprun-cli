package cli

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Songmu/prompter"
	"github.com/sacloud/apprun-api-go"
)

type DeleteOption struct {
	Force bool `help:"Force delete without confirmation"`
}

func (c *CLI) runDelete(ctx context.Context) error {
	opt := c.Delete
	app, err := c.LoadApplication(ctx, c.Application)
	if err != nil {
		return fmt.Errorf("failed to load application: %w", err)
	}
	info, _, err := c.getApplicationByName(ctx, app.Name)
	if err != nil {
		return fmt.Errorf("failed to get application: %w", err)
	}
	if opt.Force || prompter.YN(fmt.Sprintf("Do you really want to delete %s (id:%s)?", app.Name, info.Id), false) {
		return c.deleteApplication(ctx, info.Id)
	} else {
		slog.Info("canceled")
	}
	return nil
}

func (c *CLI) deleteApplication(ctx context.Context, id string) error {
	op := apprun.NewApplicationOp(c.client)
	err := op.Delete(ctx, id)
	if err != nil {
		return err
	}
	slog.Info("deleted", "id", id)
	return nil
}
