---
goal: "Implement live RCON command execution for ARK servers"
condition: "POST /servers/:id/rcon/execute returns command output, WS /ws/rcon/:serverId streams terminal I/O, RCONConsole xterm.js panel renders on server detail page"
state: "completed"
current_task: 0
notes: |
  All 10 tasks from the 2026-07-24-rcon-integration.md plan are complete plus
  four extras: backend WS hardening (size/rate/timeouts), Go HTTP handler
  tests via an injectable executor + nil-safe audit logger, a pt-BR locale
  file (AGENTS.md claimed support but only en/zh existed), and a vitest setup
  with extractedpure RCON input parser unit tests.
  Verification:
    - go build ./... + go vet ./... + go test ./... : all green
      (controllers/rcon now has 5 passing tests)
    - npm run build: passes (/servers/[id] route 86.2kB)
    - npm test (vitest): 19/19 pass
    - npm run lint: no new warnings in new files
  Commits on top of 7aaf0ad (the last commit from the prior session):
    4f8eae8  feat: register RCON WebSocket route and auth-protect WS endpoints
    56ee85b  feat: add live RCON console (xterm.js) on server detail page
    e521140  feat: harden RCON WS handler and add HTTP handler tests
    2853d99  feat: add pt-BR locale and vitest setup with input parsing tests
