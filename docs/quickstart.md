# Quickstart

Start Codex in the current project:

```bash
myharness --harness codex
```

Start Claude Code:

```bash
myharness --harness claude
```

Attach/detach sessions are persistent Byobu/tmux sessions. Detach with
`Ctrl+a d`, then re-run the same command to attach.

Aliases:

```bash
myCodex
myClaude
myOpenCode
myHermes
```

Use private project state:

```bash
myharness --harness codex --private-env
```

Add a bind mount:

```bash
myharness --harness codex -v ./cache:/mnt/cache
```
