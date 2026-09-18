package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/go-jsonnet"
	apprun "github.com/sacloud/sacloud-sdk-go/api/apprun"
	v1 "github.com/sacloud/sacloud-sdk-go/api/apprun/apis/v1"
	"slices"
	"strings"
)

// Application represents an application definition
// This is combined struct of v1.CreateApplicationBody and v1.PatchPacketFilterBody
type Application struct {
	// Components uses the patch type because it is a superset of the create type:
	// secret values are optional (the API never returns them, and omitting a value
	// on update keeps the one stored in the latest version).
	Components     []v1.PatchApplicationBodyComponentsItem `json:"components"`
	MaxScale       int                                     `json:"max_scale"`
	MinScale       int                                     `json:"min_scale"`
	Name           string                                  `json:"name"`
	Port           int                                     `json:"port"`
	TimeoutSeconds int                                     `json:"timeout_seconds"`

	PacketFilter v1.PatchPacketFilterBody `json:"packet_filter"`
}

type ApplicationInfo = v1.HandlerListApplicationsDataItem

// CreateApplicationBody returns v1.CreateApplicationBody representation of Application
func (app *Application) CreateApplicationBody() (*v1.CreateApplicationBody, error) {
	for _, c := range app.Components {
		for _, s := range c.Secret.Value {
			if !s.Value.Set {
				return nil, fmt.Errorf("component %q: secret %q requires a value to create an application", c.Name, s.Key)
			}
		}
	}
	b, err := json.Marshal(app.Components)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal components: %w", err)
	}
	var components []v1.CreateApplicationBodyComponentsItem
	if err := json.Unmarshal(b, &components); err != nil {
		return nil, fmt.Errorf("failed to convert components: %w", err)
	}
	return &v1.CreateApplicationBody{
		Components:     components,
		MaxScale:       app.MaxScale,
		MinScale:       app.MinScale,
		Name:           app.Name,
		Port:           app.Port,
		TimeoutSeconds: app.TimeoutSeconds,
	}, nil
}

func fromV1Application(v *v1.HandlerReadApplication) *Application {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	var app Application
	if err := json.Unmarshal(b, &app); err != nil {
		panic(err)
	}
	for i := range app.Components {
		// action is a request-only field. The decoder fills it with the default value, so reset it.
		if cr := &app.Components[i].DeploySource.ContainerRegistry; cr.Set {
			cr.Value.Action.Reset()
		}
	}
	return &app
}

func toUpdateV1Application(app *Application, allTraffic bool) *v1.PatchApplicationBody {
	b, err := json.Marshal(app)
	if err != nil {
		panic(err)
	}
	var v v1.PatchApplicationBody
	if err := json.Unmarshal(b, &v); err != nil {
		panic(err)
	}
	v.TimeoutSeconds = v1.NewOptInt(app.TimeoutSeconds)
	v.Port = v1.NewOptInt(app.Port)
	v.MinScale = v1.NewOptInt(app.MinScale)
	v.MaxScale = v1.NewOptInt(app.MaxScale)
	v.AllTrafficAvailable = v1.NewOptBool(allTraffic)
	slog.Debug("toUpdateV1Application", "body", toJSON(app), "allTraffic", allTraffic)
	return &v
}

func toJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("<json marshal error: %s>", err)
	}
	return string(b)
}

func toJSONIndent(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("<json marshal error: %s>", err)
	}
	return string(b)
}

func toMap(v any) map[string]any {
	m := make(map[string]any)
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(b, &m); err != nil {
		panic(err)
	}
	return m
}

func (app *Application) Validate() error {
	var errs []string
	for _, c := range app.Components {
		if !slices.Contains(apprun.ApplicationMaxCPUs, string(c.MaxCPU)) {
			errs = append(errs, fmt.Sprintf("component %q: invalid max_cpu %q (valid values: %s)", c.Name, c.MaxCPU, strings.Join(apprun.ApplicationMaxCPUs, ", ")))
		}
		if !slices.Contains(apprun.ApplicationMaxMemories, string(c.MaxMemory)) {
			errs = append(errs, fmt.Sprintf("component %q: invalid max_memory %q (valid values: %s)", c.Name, c.MaxMemory, strings.Join(apprun.ApplicationMaxMemories, ", ")))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("validation failed:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}

func (c *CLI) LoadApplication(ctx context.Context, name string) (*Application, error) {
	if name == "" {
		return nil, fmt.Errorf("application name is required. use --app flag or set APPRUN_CLI_APP environment variable")
	}
	slog.Info("loading application", "file", name)

	var buf bytes.Buffer
	c.loader.SetWriter(&buf)
	c.loader.Filename = name
	if err := c.loader.Run(ctx); err != nil {
		return nil, fmt.Errorf("failed to evaluate jsonnet file: %s", err)
	}
	app := &Application{}
	if err := json.Unmarshal(buf.Bytes(), app); err != nil {
		return nil, fmt.Errorf("failed to unmarshal jsonnet result: %s", err)
	}
	if err := app.Validate(); err != nil {
		return nil, err
	}
	return app, nil
}

func DefaultJsonnetNativeFuncs() []*jsonnet.NativeFunction {
	return []*jsonnet.NativeFunction{}
}

func (c *CLI) getApplicationByName(ctx context.Context, name string) (*ApplicationInfo, *Application, error) {
	for data, err := range c.allApplications(ctx) {
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list applications: %s", err)
		}
		if data.Name != name {
			continue
		}
		op := apprun.NewApplicationOp(c.client)
		id := data.ID
		v1app, err := op.Read(ctx, id)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read application: %s", err)
		}
		app := fromV1Application(v1app)

		if pf, err := c.getPacketFilter(ctx, id); err != nil {
			return nil, nil, fmt.Errorf("failed to get packet filter: %s", err)
		} else if pf != nil {
			app.PacketFilter = *pf
		}

		return data, app, nil
	}
	return nil, nil, ErrNotFound
}
