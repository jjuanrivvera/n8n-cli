// Package n8ncli embeds the agent-skill assets (SKILL.md + references) into the
// binary so `n8nctl skills install` can write them into an AI agent's skills
// directory. The same files under skills/n8nctl-cli/ are what
// `npx skills add jjuanrivvera/n8n-cli` and the Claude Code plugin consume, so
// there is a single canonical copy with no drift.
package n8ncli

import (
	"embed"
	"io/fs"
)

//go:embed skills/n8nctl-cli
var embedded embed.FS

// SkillName is the directory the skill installs into within an agent's skills dir.
const SkillName = "n8nctl-cli"

// SkillFS is rooted at the skill directory, so it contains SKILL.md and
// references/ at its top level.
var SkillFS = mustSub(embedded, "skills/"+SkillName)

func mustSub(f embed.FS, dir string) fs.FS {
	sub, err := fs.Sub(f, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
