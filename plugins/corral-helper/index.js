/**
 * corral-helper for opencode.
 *
 * The opencode entry point of this plugin directory: it registers the skills under skills/
 * through opencode's skill API, so Claude Code and opencode load the same SKILL.md files and
 * references. Plain JavaScript with Node built-ins only (opencode >= 2.1.0, Promise plugin API).
 */
import { existsSync, readdirSync, readFileSync, symlinkSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const SKILLS_DIR = path.join(path.dirname(fileURLToPath(import.meta.url)), "skills");
const DOC_SECTIONS = ["explanation", "how-to", "reference"];

// restoreReferences links references/<section> to the repository's docs/<section> when the
// per-file symlinks are missing: package installs from Git drop symlinks but keep docs/.
function restoreReferences(skillDir) {
  const refs = path.join(skillDir, "references");
  const root = path.resolve(SKILLS_DIR, "..", "..", "..");
  for (const section of DOC_SECTIONS) {
    const link = path.join(refs, section);
    const target = path.join(root, "docs", section);
    if (existsSync(link) || !existsSync(target)) continue;
    try {
      symlinkSync(path.relative(refs, target), link, "dir");
    } catch (e) {}
  }
}

// parseSkill reads the top-level frontmatter keys SKILL.md uses: plain `key: value` and folded
// (`>-`) or literal (`|`) block scalars.
function parseSkill(text) {
  const m = /^---\r?\n([\s\S]*?)\r?\n---\r?\n?([\s\S]*)$/.exec(text);
  if (!m) return { data: {}, content: text };
  const data = {};
  const lines = m[1].split(/\r?\n/);
  for (let i = 0; i < lines.length; i++) {
    const kv = /^([A-Za-z0-9_-]+):\s*(.*)$/.exec(lines[i]);
    if (!kv) continue;
    let value = kv[2].trim();
    if (/^[>|][-+]?$/.test(value)) {
      const block = [];
      while (i + 1 < lines.length && (/^\s/.test(lines[i + 1]) || lines[i + 1] === "")) block.push(lines[++i].trim());
      value = block.join(value.startsWith(">") ? " " : "\n").trim();
    } else {
      value = value.replace(/^(["'])(.*)\1$/, "$2");
    }
    data[kv[1]] = value;
  }
  return { data, content: m[2] };
}

function loadSkills() {
  const skills = [];
  for (const entry of readdirSync(SKILLS_DIR, { withFileTypes: true })) {
    if (!entry.isDirectory()) continue;
    const dir = path.join(SKILLS_DIR, entry.name);
    const file = path.join(dir, "SKILL.md");
    let text;
    try {
      text = readFileSync(file, "utf8");
    } catch (e) {
      continue;
    }
    restoreReferences(dir);
    const { data, content } = parseSkill(text);
    const skill = { id: entry.name, name: data.name || entry.name, path: file, content };
    if (data.description) skill.description = data.description;
    skills.push(skill);
  }
  return skills;
}

export default {
  id: "corral-helper",
  setup: async (ctx) => {
    const skills = loadSkills();
    await ctx.skill.transform((editor) => {
      for (const skill of skills) {
        if (!editor.get(skill.id)) editor.add(skill);
      }
    });
  },
};
