package opencode

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	dir := filepath.Join(t.TempDir(), "opencode-policy")
	env := activationEnv(filepath.Join(dir, "index.js"))
	var cfg struct {
		Plugins     []string            `json:"plugins"`
		Permissions []map[string]string `json:"permissions"`
		Skills      []string            `json:"skills"`
	}
	if err := json.Unmarshal([]byte(env[configContentEnv]), &cfg); err != nil {
		t.Fatalf("%s is not JSON: %v", configContentEnv, err)
	}
	if !slices.Equal(cfg.Skills, []string{filepath.Join(dir, "skills")}) {
		t.Errorf("skills = %q, want the skills directory beside the plugin", cfg.Skills)
	}
	if !slices.Equal(cfg.Plugins, []string{dir}) {
		t.Errorf("plugins = %q, want the extension directory", cfg.Plugins)
	}
	if len(cfg.Permissions) != 1 || cfg.Permissions[0]["effect"] != "ask" || cfg.Permissions[0]["action"] != "*" {
		t.Errorf("permissions = %v, want one catch-all ask rule", cfg.Permissions)
	}
}

func TestMaterializeSkillMirrorsReferenceLayout(t *testing.T) {
	root, err := materializeSkill(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"SKILL.md", "references/agent.md", "references/how-to/install.md", "references/reference/config.md"} {
		if _, err := os.Stat(filepath.Join(root, "corral", rel)); err != nil {
			t.Errorf("skill file %s: %v", rel, err)
		}
	}
}

// Every reference symlink in the repository must have an embedded counterpart.
func TestSkillFilesCoverReferenceSymlinks(t *testing.T) {
	files, err := skillFiles()
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join("..", "..", "..", filepath.FromSlash(skillSrc))
	err = filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if _, ok := files[filepath.ToSlash(rel)]; !ok {
			t.Errorf("%s is not embedded", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestBridgeIsEmbedded(t *testing.T) {
	if !strings.Contains(string(bridgeSource), `ctx.tool.hook("execute.before"`) {
		t.Error("embedded bridge does not register execute.before")
	}
}
