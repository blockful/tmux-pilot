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

## Usage

```bash
tmux-pilot [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `-h`, `--help` | Show usage and exit |
| `-v`, `--version` | Show version and exit |
| `-l`, `--list` | Print sessions as TSV and exit |
| `-S <path>` | Use custom tmux socket path |
| `-L <name>` | Use named tmux socket |
| `--no-color` | Disable ANSI colors |

### Examples

```bash
# Basic usage
tmux-pilot

# Custom socket path
tmux-pilot -S /tmp/custom-socket

# Named socket
tmux-pilot -L development

# Non-interactive TSV output
tmux-pilot --list

# Disable colors
tmux-pilot --no-color
NO_COLOR=1 tmux-pilot
```

### List Mode Output

With `--list`, sessions are printed as tab-separated values:

```
main	3	attached
api-server	1	detached
notes	2	detached
```

## Setup

Add to `~/.tmux.conf`:

```bash
bind s display-popup -E -w 60% -h 50% "tmux-pilot"
```

Optional alias in `~/.bashrc` or `~/.zshrc`:

```bash
alias tp="tmux-pilot"
```

## Keybindings

| Key | Action | Command |
|-----|--------|---------|
| `↑`/`k` | Navigate up | - |
| `↓`/`j` | Navigate down | - |
| `Enter` | Switch/attach to session | `tmux switch-client -t <name>` |
| `n` | Create new session | `tmux new-session -d -s <name>` |
| `r` | Rename session | `tmux rename-session -t <old> <new>` |
| `x` | Kill session | `tmux kill-session -t <name>` |
| `d` | Detach client | `tmux detach-client` |
| `q`/`Esc` | Quit without action | - |

## Features

- **Zero dependencies** except `golang.org/x/term`
- **Raw ANSI rendering** for maximum compatibility
- **Socket support** for isolated tmux instances
- **Color control** respects `NO_COLOR` environment variable
- **In-place rendering** without alt screen mode
- **Signal handling** for graceful cleanup

## License

MIT
