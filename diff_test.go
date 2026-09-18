package cli_test

import (
	"strings"
	"testing"

	cli "github.com/fujiwara/apprun-cli"
	v1 "github.com/sacloud/sacloud-sdk-go/api/apprun/apis/v1"
)

func newDiffTestApp(name string, password string) *cli.Application {
	return &cli.Application{
		Name: name,
		Components: []v1.PatchApplicationBodyComponentsItem{
			{
				Name:      "test",
				MaxCPU:    "0.5",
				MaxMemory: "1Gi",
				DeploySource: v1.PatchApplicationBodyComponentsItemDeploySource{
					ContainerRegistry: v1.NewOptPatchApplicationBodyComponentsItemDeploySourceContainerRegistry(
						v1.PatchApplicationBodyComponentsItemDeploySourceContainerRegistry{
							Image:    "example.sakuracr.jp/debian:latest",
							Password: v1.NewOptNilString(password),
						},
					),
				},
			},
		},
	}
}

func TestDiffApplicationsIgnore(t *testing.T) {
	remote := newDiffTestApp("remote", "")
	local := newDiffTestApp("local", "password")

	diff, err := cli.DiffApplications("remote", remote, "local", local, nil)
	if err != nil {
		t.Fatalf("DiffApplications() = %v, want nil", err)
	}
	if !strings.Contains(diff, `"name"`) {
		t.Errorf("diff must contain name:\n%s", diff)
	}
	if strings.Contains(diff, "password") {
		t.Errorf("password must be ignored by default:\n%s", diff)
	}

	diff, err = cli.DiffApplications("remote", remote, "local", local, []string{".name"})
	if err != nil {
		t.Fatalf("DiffApplications() with ignore = %v, want nil", err)
	}
	if diff != "" {
		t.Errorf("unexpected diff:\n%s", diff)
	}
}

func TestDiffApplicationsSecret(t *testing.T) {
	type secrets = []v1.PatchApplicationBodyComponentsItemSecretItem
	newApp := func(s secrets) *cli.Application {
		app := newDiffTestApp("test", "")
		if s != nil {
			app.Components[0].Secret = v1.NewOptNilPatchApplicationBodyComponentsItemSecretItemArray(s)
		}
		return app
	}
	tests := []struct {
		name     string
		remote   secrets
		local    secrets
		wantDiff string // substring expected in the diff. empty means no diff
	}{
		{"no secrets", nil, nil, ""},
		{
			"value is ignored",
			secrets{{Key: "FOO"}},
			secrets{{Key: "FOO", Value: v1.NewOptString("secret-value")}},
			"",
		},
		{
			"key is compared",
			secrets{{Key: "FOO"}},
			secrets{{Key: "BAR", Value: v1.NewOptString("secret-value")}},
			"BAR",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff, err := cli.DiffApplications("remote", newApp(tt.remote), "local", newApp(tt.local), nil)
			if err != nil {
				t.Fatalf("DiffApplications() = %v, want nil", err)
			}
			if tt.wantDiff == "" && diff != "" {
				t.Errorf("unexpected diff:\n%s", diff)
			}
			if !strings.Contains(diff, tt.wantDiff) {
				t.Errorf("diff does not contain %q:\n%s", tt.wantDiff, diff)
			}
			if strings.Contains(diff, "secret-value") {
				t.Errorf("secret value leaked in diff:\n%s", diff)
			}
		})
	}
}
