package cli_test

import (
	"context"
	"testing"

	cli "github.com/fujiwara/apprun-cli"
	"github.com/google/go-cmp/cmp"
	v1 "github.com/sacloud/sacloud-sdk-go/api/apprun/apis/v1"
)

var testApplication = &cli.Application{
	MaxScale:       2,
	MinScale:       1,
	Name:           "test",
	Port:           80,
	TimeoutSeconds: 10,
	Components: []v1.PatchApplicationBodyComponentsItem{
		{
			Name: "test",
			DeploySource: v1.PatchApplicationBodyComponentsItemDeploySource{
				ContainerRegistry: v1.NewOptPatchApplicationBodyComponentsItemDeploySourceContainerRegistry(
					v1.PatchApplicationBodyComponentsItemDeploySourceContainerRegistry{
						Username: v1.NewOptNilString("apprun"),
						Password: v1.NewOptNilString("password"),
						Server:   v1.NewOptNilString("example.sakuracr.jp"),
						Image:    "example.sakuracr.jp/debian:latest",
						Action:   v1.NewOptNilContainerRegistryAction(v1.ContainerRegistryActionNew), // default
					},
				),
			},
			Env: v1.NewOptNilRequestEnv(
				v1.RequestEnv{
					{
						Key:   "FOO",
						Value: "BAR",
					},
				},
			),
			Secret: v1.NewOptNilPatchApplicationBodyComponentsItemSecretItemArray(
				[]v1.PatchApplicationBodyComponentsItemSecretItem{
					{
						Key:   "SECRET_FOO",
						Value: v1.NewOptString("secret"),
					},
					{
						Key: "SECRET_KEEP",
					},
				},
			),
			MaxCPU:    "0.5",
			MaxMemory: "1Gi",
			Probe: v1.NewOptNilPatchApplicationBodyComponentsItemProbe(
				v1.PatchApplicationBodyComponentsItemProbe{
					HTTPGet: v1.NewOptNilPatchApplicationBodyComponentsItemProbeHTTPGet(
						v1.PatchApplicationBodyComponentsItemProbeHTTPGet{
							Headers: []v1.PatchApplicationBodyComponentsItemProbeHTTPGetHeadersItem{
								{
									Name:  v1.NewOptString("X-Test"),
									Value: v1.NewOptString("test"),
								},
							},
							Path: "/",
							Port: 80,
						},
					),
				},
			),
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
				Components: []v1.PatchApplicationBodyComponentsItem{
					{
						Name:      "test",
						MaxCPU:    v1.PatchApplicationBodyComponentsItemMaxCPU(tt.cpu),
						MaxMemory: v1.PatchApplicationBodyComponentsItemMaxMemory(tt.memory),
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

func TestCreateApplicationBody(t *testing.T) {
	newApp := func(secrets ...v1.PatchApplicationBodyComponentsItemSecretItem) *cli.Application {
		return &cli.Application{
			Name: "test",
			Components: []v1.PatchApplicationBodyComponentsItem{
				{
					Name:      "test",
					MaxCPU:    "0.5",
					MaxMemory: "1Gi",
					Secret:    v1.NewOptNilPatchApplicationBodyComponentsItemSecretItemArray(secrets),
				},
			},
		}
	}

	t.Run("secret with value", func(t *testing.T) {
		app := newApp(v1.PatchApplicationBodyComponentsItemSecretItem{Key: "FOO", Value: v1.NewOptString("secret")})
		body, err := app.CreateApplicationBody()
		if err != nil {
			t.Fatalf("CreateApplicationBody() = %v, want nil", err)
		}
		want := []v1.CreateApplicationBodyComponentsItemSecretItem{{Key: "FOO", Value: "secret"}}
		if diff := cmp.Diff(want, body.Components[0].Secret.Value); diff != "" {
			t.Errorf("secret mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("secret without value", func(t *testing.T) {
		app := newApp(v1.PatchApplicationBodyComponentsItemSecretItem{Key: "FOO"})
		if _, err := app.CreateApplicationBody(); err == nil {
			t.Error("CreateApplicationBody() = nil, want error")
		}
	})
}

func TestFromV1ApplicationWithSecret(t *testing.T) {
	remote := &v1.HandlerReadApplication{
		Name: "test",
		Components: []v1.HandlerReadApplicationComponentsItem{
			{
				Name:      "test",
				MaxCPU:    "0.5",
				MaxMemory: "1Gi",
				DeploySource: v1.HandlerReadApplicationComponentsItemDeploySource{
					ContainerRegistry: v1.NewOptHandlerReadApplicationComponentsItemDeploySourceContainerRegistry(
						v1.HandlerReadApplicationComponentsItemDeploySourceContainerRegistry{
							Image: "example.sakuracr.jp/debian:latest",
						},
					),
				},
				Secret: v1.ResponseSecret{{Key: "FOO"}},
			},
		},
	}
	app := cli.FromV1Application(remote)
	c := app.Components[0]
	want := []v1.PatchApplicationBodyComponentsItemSecretItem{{Key: "FOO"}}
	if diff := cmp.Diff(want, c.Secret.Value); diff != "" {
		t.Errorf("secret mismatch (-want +got):\n%s", diff)
	}
	if c.DeploySource.ContainerRegistry.Value.Action.Set {
		t.Error("container_registry.action must not be set for the application read from API")
	}
}
