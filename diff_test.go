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
		Components: []v1.CreateApplicationBodyComponentsItem{
			{
				Name:      "test",
				MaxCPU:    "0.5",
				MaxMemory: "1Gi",
				DeploySource: v1.CreateApplicationBodyComponentsItemDeploySource{
					ContainerRegistry: v1.NewOptCreateApplicationBodyComponentsItemDeploySourceContainerRegistry(
						v1.CreateApplicationBodyComponentsItemDeploySourceContainerRegistry{
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
