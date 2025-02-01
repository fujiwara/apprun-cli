package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"github.com/sacloud/apprun-api-go"
	v1 "github.com/sacloud/apprun-api-go/apis/v1"
)

type Application = v1.PostApplicationBody

type ApplicationInfo = v1.HandlerListApplicationsData

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

func toUpdateV1Application(app *Application) *v1.PatchApplicationBody {
	b, err := json.Marshal(app)
	if err != nil {
		panic(err)
	}
	var v v1.PatchApplicationBody
	if err := json.Unmarshal(b, &v); err != nil {
		panic(err)
	}
	v.AllTrafficAvailable = ptr(true) // TODO: configurable
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

func LoadApplication(ctx context.Context, name string) (*Application, error) {
	vm := jsonnet.MakeVM()
	nativeFuncs := DefaultJsonnetNativeFuncs()
	for _, f := range nativeFuncs {
		vm.NativeFunction(f)
	}
	jsonStr, err := vm.EvaluateFile(name)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate jsonnet file: %s", err)
	}
	app := &Application{}
	if err := json.Unmarshal([]byte(jsonStr), app); err != nil {
		return nil, fmt.Errorf("failed to unmarshal jsonnet result: %s", err)
	}
	return app, nil
}

func DefaultJsonnetNativeFuncs() []*jsonnet.NativeFunction {
	return []*jsonnet.NativeFunction{
		{
			Name:   "env",
			Params: []ast.Identifier{"name", "default"},
			Func: func(args []any) (any, error) {
				key, ok := args[0].(string)
				if !ok {
					return nil, fmt.Errorf("env: name must be a string")
				}
				if v := os.Getenv(key); v != "" {
					return v, nil
				}
				return args[1], nil
			},
		},
		{
			Name:   "must_env",
			Params: []ast.Identifier{"name"},
			Func: func(args []any) (any, error) {
				key, ok := args[0].(string)
				if !ok {
					return nil, fmt.Errorf("must_env: name must be a string")
				}
				if v, ok := os.LookupEnv(key); ok {
					return v, nil
				}
				return nil, fmt.Errorf("must_env: %s is not set", key)
			},
		},
	}
}

func (c *CLI) getApplicationByName(ctx context.Context, name string) (*ApplicationInfo, *Application, error) {
	for data, err := range c.allApplications(ctx) {
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list applications: %s", err)
		}
		if v(data.Name) != name {
			continue
		}
		op := apprun.NewApplicationOp(c.client)
		id := v(data.Id)
		v1app, err := op.Read(ctx, id)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read application: %s", err)
		}
		app := fromV1Application(v1app)
		return data, app, nil
	}
	return nil, nil, ErrNotFound
}
