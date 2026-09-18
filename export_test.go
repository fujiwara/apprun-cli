package cli

import "context"

func (c *CLI) SetupVM(ctx context.Context) error {
	return c.setupVM(ctx)
}

var FromV1Application = fromV1Application

var DiffApplications = diffApplications

func (c *CLI) RunDeploy(ctx context.Context) error {
	return c.runDeploy(ctx)
}

func (c *CLI) GetApplicationByName(ctx context.Context, name string) (*ApplicationInfo, *Application, error) {
	return c.getApplicationByName(ctx, name)
}
