/**
 * corral-helper for opencode.
 *
 * Registers the skills under skills/ through opencode's skill API, so Claude Code and
 * opencode load the same SKILL.md files and references (opencode >= 2.1.0).
 */
import { existsSync, readdirSync, readFileSync, symlinkSync } from "node:fs";
import path from "node:path";

const SKILLS_DIR = path.join(import.meta.dirname, "skills");
const DOCS_DIR = path.join(import.meta.dirname, "..", "..", "docs");

// Installs from Git drop the reference symlinks but keep docs/, so link each missing
// references/<section> to its docs/<section>.
function restoreReferences(refs) {
  for (const section of ["explanation", "how-to", "reference"]) {
    const link = path.join(refs, section);
    const target = path.join(DOCS_DIR, section);
    if (!existsSync(link) && existsSync(target)) symlinkSync(path.relative(refs, target), link, "dir");
  }
}

function loadSkill(id) {
  const file = path.join(SKILLS_DIR, id, "SKILL.md");
  const [, frontmatter, content] = /^---\n([\s\S]*?)\n---\n([\s\S]*)$/.exec(readFileSync(file, "utf8"));
  const { name, description } = Bun.YAML.parse(frontmatter);
  try {
    restoreReferences(path.join(SKILLS_DIR, id, "references"));
  } catch {}
  return { id, name, description, path: file, content };
}

export default {
  id: "corral-helper",
  setup: async (ctx) => {
    const skills = readdirSync(SKILLS_DIR).map(loadSkill);
    await ctx.skill.transform((editor) => {
      for (const skill of skills) if (!editor.get(skill.id)) editor.add(skill);
    });
  },
};
