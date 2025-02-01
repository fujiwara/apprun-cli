package cli

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/alecthomas/kong"
	"github.com/sacloud/apprun-api-go"
)

type CLI struct {
	Init     InitOption     `cmd:"" help:"Initialize files from existing application"`
	Deploy   DeployOption   `cmd:"" help:"Deploy an application"`
	List     ListOption     `cmd:"" help:"List applications"`
	Diff     DiffOption     `cmd:"" help:"Show diff of applications"`
	Render   RenderOption   `cmd:"" help:"Render application"`
	Status   StatusOption   `cmd:"" help:"Show status of applications"`
	Delete   DeleteOption   `cmd:"" help:"Delete the application"`
	Versions VersionsOption `cmd:"" help:"Show versions of application"`
	User     UserOption     `cmd:"" help:"Manage apprun user"`

	Debug bool `help:"Enable debug mode" env:"DEBUG"`

	client *apprun.Client
}

func (c *CLI) Run(ctx context.Context) error {
	k := kong.Parse(c)
	var err error
	if c.Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	switch k.Command() {
	case "init":
		err = c.runInit(ctx)
	case "deploy <application>":
		err = c.runDeploy(ctx)
	case "diff <application>":
		err = c.runDiff(ctx)
	case "render <application>":
		err = c.runRender(ctx)
	case "status <application>":
		err = c.runStatus(ctx)
	case "delete <application>":
		err = c.runDelete(ctx)
	case "list":
		err = c.runList(ctx)
	case "user <operation>":
		err = c.runUser(ctx)
	case "versions <application>":
		err = c.runVersions(ctx)
	default:
		err = fmt.Errorf("unknown command: %s", k.Command())
	}
	return err
}
