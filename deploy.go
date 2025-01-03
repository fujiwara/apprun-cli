package cli

import (
	"context"
	"errors"
	"log/slog"

	"github.com/sacloud/apprun-api-go"
)

type DeployOption struct {
	Application string `arg:"" help:"Name of the definition file to deploy" required:""`
}

func (c *CLI) runDeploy(ctx context.Context) error {
	opt := c.Deploy
	app, err := LoadApplication(ctx, opt.Application)
	if err != nil {
		return err
	}
	slog.Info("deploying", "app", app.Name)
	info, _, err := c.getApplicationByName(ctx, app.Name)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return err
		}
		slog.Info("creating", "app", app.Name)
		return c.createApplication(ctx, app)
	}
	slog.Info("updating", "app", app.Name)
	return c.updateApplication(ctx, v(info.Id), app)
}

func (c *CLI) createApplication(ctx context.Context, app *Application) error {
	op := apprun.NewApplicationOp(c.client)
	created, err := op.Create(ctx, app)
	if err != nil {
		return err
	}
	slog.Info("created", "id", v(created.Id))
	return nil
}

func (c *CLI) updateApplication(ctx context.Context, id string, app *Application) error {
	op := apprun.NewApplicationOp(c.client)
	updated, err := op.Update(ctx, id, toUpdateV1Application(app))
	if err != nil {
		return err
	}
	slog.Info("updated", "id", v(updated.Id))
	return nil
}
