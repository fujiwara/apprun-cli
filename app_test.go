package cli_test

import (
	"context"
	"testing"

	cli "github.com/fujiwara/apprun-cli"
	"github.com/google/go-cmp/cmp"
	v1 "github.com/sacloud/apprun-api-go/apis/v1"
)

func ptr[T any](v T) *T {
	return &v
}

var testApplication = &cli.Application{
	MaxScale:       2,
	MinScale:       1,
	Name:           "test",
	Port:           80,
	TimeoutSeconds: 10,
	Components: []v1.PostApplicationBodyComponent{
		{
			Name: "test",
			DeploySource: v1.PostApplicationBodyComponentDeploySource{
				ContainerRegistry: &v1.PostApplicationBodyComponentDeploySourceContainerRegistry{
					Username: ptr("apprun"),
					Password: ptr("password"),
					Server:   ptr("example.sakuracr.jp"),
					Image:    "example.sakuracr.jp/debian:latest",
				},
			},
			Env: &[]v1.PostApplicationBodyComponentEnv{
				{
					Key:   ptr("FOO"),
					Value: ptr("BAR"),
				},
			},
			MaxCpu:    "0.5",
			MaxMemory: "1Gi",
			Probe: &v1.PostApplicationBodyComponentProbe{
				HttpGet: &v1.PostApplicationBodyComponentProbeHttpGet{
					Headers: &[]v1.PostApplicationBodyComponentProbeHttpGetHeader{
						{
							Name:  ptr("X-Test"),
							Value: ptr("test"),
						},
					},
					Path: "/",
					Port: 80,
				},
			},
		},
	},
}

func newCLI(t *testing.T, ctx context.Context) *cli.CLI {
	c, err := cli.New(ctx)
	if err != nil {
		t.Fatalf("cli.New() = %v, want nil", err)
	}
	c.TFState = "testdata/terraform.tfstate"
	if err := c.SetupVM(ctx); err != nil {
		t.Fatalf("c.SetupVM() = %v, want nil", err)
	}
	return c
}

func TestLoadApplication(t *testing.T) {
	ctx := t.Context()
	t.Setenv("REGISTRY_PASSWORD", "password")
	for _, p := range []string{"testdata/app.json", "testdata/app.jsonnet"} {
		c := newCLI(t, ctx)
		app, err := c.LoadApplication(ctx, p)
		if err != nil {
			t.Errorf("c.LoadApplication(%s) = %v, want nil", p, err)
		}
		if diff := cmp.Diff(app, testApplication); diff != "" {
			t.Errorf("c.LoadApplication(%s) mismatch (-want +got):\n%s", p, diff)
		}
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cpu     string
		memory  string
		wantErr bool
	}{
		{"valid 0.5/1Gi", "0.5", "1Gi", false},
		{"valid 1/2Gi", "1", "2Gi", false},
		{"valid 2/4Gi", "2", "4Gi", false},
		{"invalid cpu", "0.1", "1Gi", true},
		{"invalid memory", "0.5", "512Mi", true},
		{"both invalid", "0.3", "256Mi", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &cli.Application{
				Name:           "test",
				Port:           80,
				TimeoutSeconds: 10,
				MinScale:       1,
				MaxScale:       2,
				Components: []v1.PostApplicationBodyComponent{
					{
						Name:      "test",
						MaxCpu:    v1.PostApplicationBodyComponentMaxCpu(tt.cpu),
						MaxMemory: v1.PostApplicationBodyComponentMaxMemory(tt.memory),
					},
				},
			}
			err := app.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
