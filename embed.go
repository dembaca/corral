// Package corral embeds repository files the binary ships to agents.
package corral

import "embed"

// HelperSkill holds the corral-helper skill and the docs its references point to. The
// skill's reference symlinks are not embeddable, so the docs are embedded from docs/.
//
//go:embed plugins/corral-helper/skills/corral/SKILL.md plugins/corral-helper/skills/corral/references/agent.md docs/explanation/*.md docs/how-to/*.md docs/reference/*.md
var HelperSkill embed.FS
