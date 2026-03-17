# tmux-pilot

A TUI for managing tmux sessions. One keybinding to see, create, rename, switch, and kill sessions.

[![CI](https://github.com/blockful/tmux-pilot/actions/workflows/ci.yml/badge.svg)](https://github.com/blockful/tmux-pilot/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

```
┌─ tmux-pilot ──────────────────────────────┐
│                                           │
│  ● main         3 windows   attached      │
│  ○ api-server   1 window    detached      │
│  ○ notes        2 windows   detached      │
│  ○ scratch      1 window    detached      │
│                                           │
│  [enter] switch  [n] new  [r] rename      │
│  [x] kill  [q] quit                       │
│                                           │
│  tip: Ctrl-b d to detach from tmux        │
└───────────────────────────────────────────┘
```

## Install

### Go

```bash
go install github.com/blockful/tmux-pilot/cmd@latest
```

### Download binary

Grab the latest from [Releases](https://github.com/blockful/tmux-pilot/releases) for your platform (Linux/macOS, amd64/arm64).

### Build from source

```bash
git clone https://github.com/blockful/tmux-pilot.git
cd tmux-pilot
go build -o tmux-pilot ./cmd
```

## Setup

Add to `~/.tmux.conf` for a popup overlay on `prefix + s`:

```bash
bind s display-popup -E -w 60% -h 50% "tmux-pilot"
```

Then reload: `tmux source-file ~/.tmux.conf`

Also works standalone outside tmux — it will `attach-session` instead of `switch-client`.

## Keybindings

| Key | Action |
|-----|--------|
| `↑`/`↓` `j`/`k` | Navigate |
| `enter` | Switch to session |
| `n` | New session |
| `r` | Rename session |
| `x` | Kill session (with confirmation) |
| `q` / `esc` | Quit |

## How it works

tmux-pilot shells out to the `tmux` binary for all operations. No library dependencies on tmux internals. When running inside tmux it uses `switch-client`; outside it uses `attach-session`.

Built with [BubbleTea](https://github.com/charmbracelet/bubbletea) + [LipGloss](https://github.com/charmbracelet/lipgloss).

## License

MIT
