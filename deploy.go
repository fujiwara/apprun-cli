package cli

import (
	"context"
	"errors"
	"log/slog"

	"github.com/sacloud/apprun-api-go"
)

type DeployOption struct {
	AllTraffic bool `help:"Shift all traffic for the deployed version (default:true)" default:"true" negatable:""`
}

func (c *CLI) runDeploy(ctx context.Context) error {
	opt := c.Deploy
	app, err := LoadApplication(ctx, c.Application)
	if err != nil {
		return err
	}
	slog.Info("deploying", "app", app.Name, "allTraffic", opt.AllTraffic)
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
	updated, err := op.Update(ctx, id, toUpdateV1Application(app, c.Deploy.AllTraffic))
	if err != nil {
		return err
	}
	slog.Info("updated", "id", v(updated.Id))
	return nil
}
