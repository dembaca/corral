// Package opencode is the adapter for the opencode coding agent (github.com/anomalyco/opencode,
// v2 >= 2.1.0). opencode ships no sandbox of its own; policy enforcement rides a plugin that the
// launcher activates per session through OPENCODE_CONFIG_CONTENT.
package opencode

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/go-corral/corral/internal/agents/spec"
	"github.com/go-corral/corral/internal/health"
)

type agent struct{}

func New() spec.Agent { return agent{} }

var _ spec.Agent = agent{}

const configContentEnv = "OPENCODE_CONFIG_CONTENT"

func (agent) Name() string       { return "opencode" }
func (agent) Binaries() []string { return []string{"opencode"} }

func xdgDir(home string, host map[string]string, env, fallback string) string {
	if d := strings.TrimSpace(host[env]); d != "" {
		return filepath.Join(d, "opencode")
	}
	return filepath.Join(home, fallback, "opencode")
}

func (agent) ConfigDir(home string, host map[string]string) string {
	return xdgDir(home, host, "XDG_CONFIG_HOME", ".config")
}

func (agent) Launch() spec.Launch {
	return spec.Launch{
		ExtensionAsset: bridgeSource,
		ExtensionName:  "opencode-policy/index.js",
		ExtensionEnv:   activationEnv,
		Args:           launchArgs,
	}
}

// activationEnv loads the policy plugin directory and adds an `ask` rule for every tool. The
// plugin turns each ask into allow after its own check; when the plugin fails to load (opencode
// only warns), the rule makes opencode prompt, or reject in `opencode run`, instead of running
// tools unchecked. It also writes the corral-helper skill into the plugin directory, which the
// sandbox mounts read-only, and loads it; a write failure only drops the skill.
func activationEnv(extensionPath string) map[string]string {
	dir := filepath.Dir(extensionPath)
	cfg := map[string]any{
		"plugins":     []string{dir},
		"permissions": []map[string]string{{"action": "*", "resource": "*", "effect": "ask"}},
	}
	if skills, err := materializeSkill(dir); err == nil {
		cfg["skills"] = []string{skills}
	}
	content, _ := json.Marshal(cfg)
	return map[string]string{configContentEnv: string(content)}
}

// Subcommands that accept --standalone. Without it opencode connects to its background service,
// which runs outside the sandbox or, started inside it, outlives the session.
var standaloneSubcommands = []string{"run", "mini", "models", "stats"}

// Subcommands that run locally without a server.
var localSubcommands = []string{"auth", "debug", "mcp", "plugin"}

// launchArgs pins every server-backed launch to a private --standalone server and rejects the
// flags that would bypass the sandbox or the permission fallback.
func launchArgs(args []string) ([]string, error) {
	for _, a := range args {
		name, _, _ := strings.Cut(a, "=")
		switch name {
		case "--server":
			return nil, fmt.Errorf("opencode --server connects to a server outside the sandbox; corral runs a private --standalone server")
		case "--auto":
			return nil, fmt.Errorf("opencode --auto approves the permission prompts corral uses as its fallback when the policy plugin fails to load")
		}
	}
	if len(args) == 0 || strings.HasPrefix(args[0], "-") || !isSubcommand(args[0]) {
		return withStandalone(args, 0), nil
	}
	switch {
	case slices.Contains(standaloneSubcommands, args[0]):
		return withStandalone(args, 1), nil
	case slices.Contains(localSubcommands, args[0]):
		return args, nil
	}
	return nil, fmt.Errorf("opencode %s is not supported by corral run: it starts or reaches a server outside the session's private one", args[0])
}

func withStandalone(args []string, at int) []string {
	if slices.Contains(args, "--standalone") {
		return args
	}
	out := slices.Clone(args[:at])
	out = append(out, "--standalone")
	return append(out, args[at:]...)
}

// isSubcommand separates a subcommand from the TUI's optional <directory> positional.
func isSubcommand(arg string) bool {
	return slices.Contains(standaloneSubcommands, arg) || slices.Contains(localSubcommands, arg) || !strings.ContainsAny(arg, "/.~")
}

func (a agent) ConfigPaths() []spec.ConfigPath {
	return []spec.ConfigPath{
		{Path: "$HOME/.local/share/opencode", Description: "opencode data: auth, sessions database, logs", Writeable: true, Optional: true},
		{Path: "$HOME/.local/state/opencode", Description: "opencode state", Writeable: true, Optional: true},
		{Path: "$HOME/.cache/opencode", Description: "opencode cache: plugin packages, LSP binaries", Writeable: true, Optional: true},
	}
}

func (agent) ProtectedPaths(configDir string) []string { return nil }

func (agent) Footprint() spec.Footprint {
	return spec.Footprint{
		Dir:       "opencode",
		DirReason: "opencode's config directory",
		ProtectedFiles: map[string]string{
			"opencode.json":  "opencode's global config",
			"opencode.jsonc": "opencode's global config",
			"service.json":   "opencode's background service config",
		},
		ProtectedDirs: map[string]string{
			"plugin":  "opencode's global plugin directory",
			"plugins": "opencode's global plugin directory",
		},
	}
}

func (agent) ReservedEnv() []string {
	return []string{configContentEnv}
}

func (a agent) Doctor(in spec.StatusInput) []health.Check {
	if len(a.Launch().ExtensionAsset) == 0 {
		return []health.Check{{State: health.Fail, Label: "policy plugin", Value: "missing",
			Reason: "the embedded policy plugin is empty (corral build problem)"}}
	}
	return []health.Check{{Label: "policy plugin", Value: fmt.Sprintf("embedded (%d bytes)", len(a.Launch().ExtensionAsset))}}
}

func (a agent) Sync(in spec.SyncInput) (spec.SyncReport, error) {
	return spec.SyncReport{Messages: []string{fmt.Sprintf(
		"%s enforces policy via a plugin activated at `corral run` (%s) — no settings file to sync.",
		a.Name(), configContentEnv)}}, nil
}

func (agent) LaunchWarnings(in spec.StatusInput) []health.Check { return nil }
