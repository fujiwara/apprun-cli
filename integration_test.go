package cli_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	cli "github.com/fujiwara/apprun-cli"
	"github.com/google/go-cmp/cmp"
	"github.com/sacloud/sakumock/apprun"
)

// newMockCLI returns a CLI connected to a sakumock AppRun server.
func newMockCLI(t *testing.T, ctx context.Context) (*cli.CLI, *apprun.Server) {
	t.Helper()
	srv := apprun.NewTestServer(apprun.Config{})
	t.Cleanup(srv.Close)
	t.Setenv("SAKURA_ENDPOINTS_APPRUN_SHARED", srv.TestURL())
	t.Setenv("SAKURA_ACCESS_TOKEN", "dummy")
	t.Setenv("SAKURA_ACCESS_TOKEN_SECRET", "dummy")
	t.Setenv("SAKURA_RATE_LIMIT", "1000") // saclient limits requests to 5 req/s by default

	c, err := cli.New(ctx)
	if err != nil {
		t.Fatalf("cli.New() = %v, want nil", err)
	}
	if err := c.SetupVM(ctx); err != nil {
		t.Fatalf("c.SetupVM() = %v, want nil", err)
	}
	c.Deploy.AllTraffic = true
	return c, srv
}

type testSecret struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

// writeDefinition writes an application definition file with the secrets and returns its path.
func writeDefinition(t *testing.T, secrets []testSecret) string {
	t.Helper()
	def := map[string]any{
		"name":            "integration-test",
		"port":            8080,
		"min_scale":       0,
		"max_scale":       1,
		"timeout_seconds": 60,
		"components": []map[string]any{
			{
				"name":       "app",
				"max_cpu":    "0.5",
				"max_memory": "1Gi",
				"deploy_source": map[string]any{
					"container_registry": map[string]any{
						"image":    "example.sakuracr.jp/app:latest",
						"server":   "example.sakuracr.jp",
						"username": "apprun",
						"password": "password",
					},
				},
				"env":    []map[string]string{{"key": "FOO", "value": "bar"}},
				"secret": secrets,
				"probe": map[string]any{
					"http_get": map[string]any{"path": "/", "port": 8080},
				},
			},
		},
		"packet_filter": map[string]any{
			"is_enabled": true,
			"settings": []map[string]any{
				{"from_ip": "192.0.2.0", "from_ip_prefix_length": 24},
			},
		},
	}
	b, err := json.Marshal(def)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "app.json")
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// secretKeys returns the secret keys of the first component of the deployed application.
func secretKeys(t *testing.T, ctx context.Context, c *cli.CLI) []string {
	t.Helper()
	_, remote, err := c.GetApplicationByName(ctx, "integration-test")
	if err != nil {
		t.Fatalf("GetApplicationByName() = %v, want nil", err)
	}
	keys := []string{}
	for _, s := range remote.Components[0].Secret.Value {
		if s.Value.Set {
			t.Errorf("secret %q must not have a value in the application read from API", s.Key)
		}
		keys = append(keys, s.Key)
	}
	return keys
}

// assertNoDiff asserts that the definition file has no differences from the deployed application.
func assertNoDiff(t *testing.T, ctx context.Context, c *cli.CLI) {
	t.Helper()
	local, err := c.LoadApplication(ctx, c.Application)
	if err != nil {
		t.Fatalf("LoadApplication() = %v, want nil", err)
	}
	_, remote, err := c.GetApplicationByName(ctx, local.Name)
	if err != nil {
		t.Fatalf("GetApplicationByName() = %v, want nil", err)
	}
	diff, err := cli.DiffApplications("remote", remote, "local", local, nil)
	if err != nil {
		t.Fatalf("DiffApplications() = %v, want nil", err)
	}
	if diff != "" {
		t.Errorf("unexpected diff after deploy:\n%s", diff)
	}
}

func TestDeployWithSecret(t *testing.T) {
	ctx := t.Context()
	c, srv := newMockCLI(t, ctx)

	// create
	c.Application = writeDefinition(t, []testSecret{{Key: "SECRET_A", Value: "a"}})
	if err := c.RunDeploy(ctx); err != nil {
		t.Fatalf("RunDeploy() to create = %v, want nil", err)
	}
	if diff := cmp.Diff([]string{"SECRET_A"}, secretKeys(t, ctx, c)); diff != "" {
		t.Errorf("secret keys mismatch after create (-want +got):\n%s", diff)
	}
	assertNoDiff(t, ctx, c)

	// update: keep SECRET_A as is (without value) and add SECRET_B
	c.Application = writeDefinition(t, []testSecret{{Key: "SECRET_A"}, {Key: "SECRET_B", Value: "b"}})
	if err := c.RunDeploy(ctx); err != nil {
		t.Fatalf("RunDeploy() to update = %v, want nil", err)
	}
	if diff := cmp.Diff([]string{"SECRET_A", "SECRET_B"}, secretKeys(t, ctx, c)); diff != "" {
		t.Errorf("secret keys mismatch after update (-want +got):\n%s", diff)
	}
	assertNoDiff(t, ctx, c)

	// update: a secret without value that has never been stored must be rejected
	c.Application = writeDefinition(t, []testSecret{{Key: "SECRET_UNKNOWN"}})
	if err := c.RunDeploy(ctx); err == nil {
		t.Error("RunDeploy() with an unknown secret without value = nil, want error")
	}

	if v := srv.SpecViolations(); len(v) > 0 {
		t.Errorf("OpenAPI spec violations: %+v", v)
	}
}

func TestCreateWithSecretWithoutValue(t *testing.T) {
	ctx := t.Context()
	c, _ := newMockCLI(t, ctx)

	c.Application = writeDefinition(t, []testSecret{{Key: "SECRET_A"}})
	if err := c.RunDeploy(ctx); err == nil {
		t.Fatal("RunDeploy() to create with a secret without value = nil, want error")
	}
	if _, _, err := c.GetApplicationByName(ctx, "integration-test"); err == nil {
		t.Error("the application must not be created")
	}
}
