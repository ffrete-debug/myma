# Project Memory

## Project: ark-commander (myma)
ARK Server Commander - Web UI for managing ARK: Survival Evolved servers in Docker.
- Backend: Go 1.24 + Gin + GORM + SQLite
- Frontend: Next.js 15 + React 19 + TypeScript + Tailwind CSS 4
- Auth: JWT + bcrypt
- Infra: Docker + Docker Compose

## Skills & Methodology
- Using [Superpowers](https://github.com/obra/superpowers) skills framework: brainstorming, writing-plans, subagent-driven-development, TDD, systematic-debugging, request-code-review
- Using [Anthropics frontend-design skill](https://github.com/anthropics/skills/blob/main/skills/frontend-design/SKILL.md) for UI work
- OpenCode plugin: `superpowers@git+https://github.com/obra/superpowers.git`

## Workflow
1. Brainstorm → Design spec → Writing-plans → TDD implementation → Code review → Merge
2. Always use skills before acting (brainstorming for new features, systematic-debugging for bugs)

## Python Tooling
- `poetry init` run with: name=myma, deps=pyyaml, requests, python-dotenv
- pyproject.toml created at `/workspaces/myma/pyproject.toml`

## Reference Repos Cloned
- superpowers: `/workspaces/myma/skills/superpowers/`
- anthropics-skills: `/workspaces/myma/skills/anthropics-skills/` (frontend-design skill at `skills/frontend-design/SKILL.md`)
- ark-commander reference: `/tmp/ark-commander-ref/` (cloned for source reference)

## History
- Initial session: workspace set up with ark-commander codebase (Go backend + Next.js frontend)
- Cloned superpowers skills framework as submodule at `skills/superpowers/`
- Cloned anthropics/skills as submodule at `skills/anthropics-skills/` (frontend-design skill at `skills/anthropics-skills/skills/frontend-design/SKILL.md`)
- Configured opencode.json with `superpowers@git+https://github.com/obra/superpowers.git` plugin
- Ran `poetry init` with name=myma, deps=pyyaml, requests, python-dotenv → `pyproject.toml` created
- Created MEMORY.md for cross-session memory tracking
- Created CLAUDE.md from superpowers for agent instructions
- Created .gitignore for Go + Node.js build artifacts
- Committed all setup files (2 commits)
- Next session: user should use `skill` tool to load brainstorming before starting any new feature work

## Completed Tasks
1. **Player Management** (task #1) — All priority tasks completed
2. **Real-time WebSocket Status** (task #2) — WebSocket server status updates in UI
3. **RCON Integration** — Full RCON console + HTTP/WS endpoints
4. **Backup Restore** — Full restore logic with `RestoreVolume` method