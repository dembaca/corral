package opencode

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

func TestLaunchArgsPinsStandalone(t *testing.T) {
	for _, tc := range []struct {
		in, want []string
	}{
		{nil, []string{"--standalone"}},
		{[]string{"./src"}, []string{"--standalone", "./src"}},
		{[]string{"--continue"}, []string{"--standalone", "--continue"}},
		{[]string{"run", "fix the tests"}, []string{"run", "--standalone", "fix the tests"}},
		{[]string{"run", "--standalone", "hi"}, []string{"run", "--standalone", "hi"}},
		{[]string{"models"}, []string{"models", "--standalone"}},
		{[]string{"auth", "list"}, []string{"auth", "list"}},
	} {
		got, err := launchArgs(tc.in)
		if err != nil {
			t.Errorf("launchArgs(%q): %v", tc.in, err)
			continue
		}
		if !slices.Equal(got, tc.want) {
			t.Errorf("launchArgs(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestLaunchArgsRejects(t *testing.T) {
	for _, in := range [][]string{
		{"--server", "http://127.0.0.1:49374"},
		{"run", "--server=http://127.0.0.1:49374", "hi"},
		{"--auto"},
		{"run", "--auto", "hi"},
		{"serve"},
		{"service", "start"},
		{"acp"},
	} {
		if got, err := launchArgs(in); err == nil {
			t.Errorf("launchArgs(%q) = %q, want an error", in, got)
		}
	}
}

func TestActivationEnvLoadsPluginDirWithAskFallback(t *testing.T) {
	env := activationEnv("/home/u/.cache/corral/opencode-policy/index.js")
	var cfg struct {
		Plugins     []string            `json:"plugins"`
		Permissions []map[string]string `json:"permissions"`
	}
	if err := json.Unmarshal([]byte(env[configContentEnv]), &cfg); err != nil {
		t.Fatalf("%s is not JSON: %v", configContentEnv, err)
	}
	if !slices.Equal(cfg.Plugins, []string{"/home/u/.cache/corral/opencode-policy"}) {
		t.Errorf("plugins = %q, want the extension directory", cfg.Plugins)
	}
	if len(cfg.Permissions) != 1 || cfg.Permissions[0]["effect"] != "ask" || cfg.Permissions[0]["action"] != "*" {
		t.Errorf("permissions = %v, want one catch-all ask rule", cfg.Permissions)
	}
}

func TestBridgeIsEmbedded(t *testing.T) {
	if !strings.Contains(string(bridgeSource), `ctx.tool.hook("execute.before"`) {
		t.Error("embedded bridge does not register execute.before")
	}
}
