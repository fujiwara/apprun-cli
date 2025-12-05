package cli

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/alecthomas/kong"
	armed "github.com/fujiwara/jsonnet-armed"
	"github.com/sacloud/apprun-api-go"
)

type CLI struct {
	List ListOption `cmd:"" help:"List applications"`

	Init InitOption `cmd:"" help:"Initialize files from existing application"`

	Deploy   DeployOption   `cmd:"" help:"Deploy an application"`
	Diff     DiffOption     `cmd:"" help:"Show diff of applications"`
	Render   RenderOption   `cmd:"" help:"Render application"`
	Status   StatusOption   `cmd:"" help:"Show status of applications"`
	Delete   DeleteOption   `cmd:"" help:"Delete the application"`
	Versions VersionsOption `cmd:"" help:"Manage versions of application"`
	Traffics TrafficsOption `cmd:"" help:"Manage traffics of application"`
	User     UserOption     `cmd:"" help:"Manage apprun user"`
	URL      struct{}       `cmd:"" help:"Show application public URL"`

	Debug       bool             `help:"Enable debug mode" env:"DEBUG"`
	Application string           `name:"app" help:"Name of the application definition file" env:"APPRUN_CLI_APP"`
	TFState     string           `name:"tfstate" help:"URL to terraform.tfstate" env:"APPRUN_CLI_TFSTATE"`
	Version     kong.VersionFlag `short:"v" help:"Show version and exit."`

	client *apprun.Client
	loader *armed.CLI
}

func (c *CLI) Run(ctx context.Context) error {
	k := kong.Parse(c, kong.Vars{"version": fmt.Sprintf("apprun-cli %s", Version)})
	var err error
	if c.Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else if k.Command() == "url" {
		// suppress info logs for url command
		slog.SetLogLoggerLevel(slog.LevelWarn)
	}
	if err := c.setupVM(ctx); err != nil {
		return err
	}

	switch k.Command() {
	case "list":
		err = c.runList(ctx)
	case "init":
		err = c.runInit(ctx)
	case "deploy":
		err = c.runDeploy(ctx)
	case "diff":
		err = c.runDiff(ctx)
	case "render":
		err = c.runRender(ctx)
	case "status":
		err = c.runStatus(ctx)
	case "delete":
		err = c.runDelete(ctx)
	case "user <operation>":
		err = c.runUser(ctx)
	case "versions":
		err = c.runVersions(ctx)
	case "traffics":
		err = c.runTraffics(ctx)
	case "url":
		err = c.runURL(ctx)
	default:
		err = fmt.Errorf("unknown command: %s", k.Command())
	}
	return err
}
