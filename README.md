# tmux-pilot

A minimal TUI for managing tmux sessions. Pick → execute → done.

[![CI](https://github.com/blockful/tmux-pilot/actions/workflows/ci.yml/badge.svg)](https://github.com/blockful/tmux-pilot/actions/workflows/ci.yml)

```
┌─ tmux-pilot ──────────────────────────────┐
│                                           │
│  ● main         3 windows   attached      │
│  ○ api-server   1 window    detached      │
│  ○ notes        2 windows   detached      │
│                                           │
│  [enter] switch  [n] new  [r] rename      │
│  [x] kill  [d] detach  [q] quit           │
│                                           │
│  tip: Ctrl-b d to detach                  │
└───────────────────────────────────────────┘
```

## Install

```bash
go install github.com/blockful/tmux-pilot/cmd@latest
```

Or download from [Releases](https://github.com/blockful/tmux-pilot/releases).

## Setup

Add to `~/.tmux.conf`:

```bash
bind s display-popup -E -w 60% -h 50% "tmux-pilot"
```

Optional alias in `~/.bashrc` or `~/.zshrc`:

```bash
alias tp="tmux-pilot"
```

## How it works

tmux-pilot is a thin picker. It shows your sessions, you choose an action, it exits and runs the tmux command:

| Key | Runs |
|-----|------|
| `enter` | `tmux switch-client -t <name>` |
| `n` | `tmux new-session -d -s <name> && tmux switch-client -t <name>` |
| `r` | `tmux rename-session -t <old> <new>` |
| `x` | `tmux kill-session -t <name>` |
| `d` | `tmux detach-client` |
| `q`/`esc` | exit |

Navigation: `↑`/`↓` or `j`/`k`

## License

MIT
