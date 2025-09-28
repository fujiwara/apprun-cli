package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/go-jsonnet"
	"github.com/sacloud/apprun-api-go"
	v1 "github.com/sacloud/apprun-api-go/apis/v1"
)

// Application represents an application definition
// This is combined struct of v1.PostApplicationBody and v1.PatchPacketFilter
type Application struct {
	// same as v1.PostApplicationBody
	Components     []v1.PostApplicationBodyComponent `json:"components"`
	MaxScale       int                               `json:"max_scale"`
	MinScale       int                               `json:"min_scale"`
	Name           string                            `json:"name"`
	Port           int                               `json:"port"`
	TimeoutSeconds int                               `json:"timeout_seconds"`

	PacketFilter v1.PatchPacketFilter `json:"packet_filter,omitempty"`
}

type ApplicationInfo = v1.HandlerListApplicationsData

// PostApplicationBody returns v1.PostApplicationBody representation of Application
func (app *Application) PostApplicationBody() *v1.PostApplicationBody {
	return &v1.PostApplicationBody{
		Components:     app.Components,
		MaxScale:       app.MaxScale,
		MinScale:       app.MinScale,
		Name:           app.Name,
		Port:           app.Port,
		TimeoutSeconds: app.TimeoutSeconds,
	}
}

func fromV1Application(v *v1.Application) *Application {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	var app Application
	if err := json.Unmarshal(b, &app); err != nil {
		panic(err)
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
	v.AllTrafficAvailable = ptr(allTraffic)
	slog.Debug("toUpdateV1Application", "body", toJSON(v))
	return &v
}

func toJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func toJSONIndent(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
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

func v[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}
	return *v
}

func ptr[T any](v T) *T {
	return &v
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
		id := data.Id
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
