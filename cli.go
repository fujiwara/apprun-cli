package cli

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/alecthomas/kong"
	"github.com/google/go-jsonnet"
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

	Debug           bool              `help:"Enable debug mode" env:"DEBUG"`
	Application     string            `name:"app" help:"Name of the application definition file" env:"APPRUN_CLI_APP"`
	TFState         *string           `name:"tfstate" help:"URL to terraform.tfstate" env:"APPRUN_CLI_TFSTATE"`
	PrefixedTFState map[string]string `name:"prefixed-tfstate" help:"key value pair of the prefix for template function name and URL to terraform.tfstate" env:"APPRUN_CLI_PREFIXED_TFSTATE"`

	client *apprun.Client
	vm     *jsonnet.VM
}

func (c *CLI) Run(ctx context.Context) error {
	k := kong.Parse(c)
	var err error
	if c.Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
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
	case "user":
		err = c.runUser(ctx)
	case "versions":
		err = c.runVersions(ctx)
	case "traffics":
		err = c.runTraffics(ctx)
	default:
		err = fmt.Errorf("unknown command: %s", k.Command())
	}
	return err
}
