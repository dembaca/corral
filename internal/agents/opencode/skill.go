package opencode

import (
	"bytes"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-corral/corral"
)

const skillSrc = "plugins/corral-helper/skills/corral"

// skillFiles maps the corral-helper skill to its installed layout: SKILL.md and agent.md as
// they are, and every doc under references/<section>/, where the repository's symlinks point.
func skillFiles() (map[string][]byte, error) {
	out := map[string][]byte{}
	err := fs.WalkDir(corral.HelperSkill, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(corral.HelperSkill, p)
		if err != nil {
			return err
		}
		if rel, ok := strings.CutPrefix(p, skillSrc+"/"); ok {
			out[rel] = data
		} else if rel, ok := strings.CutPrefix(p, "docs/"); ok {
			out[path.Join("references", rel)] = data
		}
		return nil
	})
	return out, err
}

// materializeSkill writes the skill to <dir>/skills/corral and returns <dir>/skills. Only files
// whose content changed are rewritten.
func materializeSkill(dir string) (string, error) {
	files, err := skillFiles()
	if err != nil {
		return "", err
	}
	root := filepath.Join(dir, "skills")
	for rel, data := range files {
		p := filepath.Join(root, "corral", filepath.FromSlash(rel))
		if cur, err := os.ReadFile(p); err == nil && bytes.Equal(cur, data) {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
			return "", err
		}
		if err := os.WriteFile(p, data, 0o644); err != nil {
			return "", err
		}
	}
	return root, nil
}
