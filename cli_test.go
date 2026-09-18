package cli_test

import (
	"strings"
	"testing"

	"github.com/alecthomas/kong"
	cli "github.com/fujiwara/apprun-cli"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr string
	}{
		// kong v1.16+ accepts no command if CLI has a method named Run()
		{"no command", []string{}, "", `expected one of "list", "init"`},
		{"flags only", []string{"--debug"}, "", `expected one of "list", "init"`},
		{"unknown command", []string{"unknown"}, "", "unexpected argument"},
		{"list", []string{"list"}, "list", ""},
		{"deploy", []string{"deploy", "--app", "app.jsonnet"}, "deploy", ""},
		{"user", []string{"user", "read"}, "user <operation>", ""},
		{"user without operation", []string{"user"}, "", "expected"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := kong.New(&cli.CLI{}, kong.Vars{"version": "test"})
			if err != nil {
				t.Fatalf("kong.New() = %v, want nil", err)
			}
			k, err := parser.Parse(tt.args)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Parse(%v) error = %v, want error containing %q", tt.args, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%v) = %v, want nil", tt.args, err)
			}
			if got := k.Command(); got != tt.want {
				t.Errorf("Parse(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}
