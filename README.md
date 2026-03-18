# tmux-pilot

A minimal TUI for managing tmux sessions. No commands to memorize.

[![CI](https://github.com/blockful/tmux-pilot/actions/workflows/ci.yml/badge.svg)](https://github.com/blockful/tmux-pilot/actions/workflows/ci.yml)

```
$ tp

╭──────────────────────────────────────────╮
│          tmux sessions                   │
│                                          │
│  ● main            3 windows  attached   │
│  ○ api-server      1 window   detached   │
│  ○ notes           2 windows  detached   │
│                                          │
╰──────────────────────────────────────────╯
↑/k ↓/j: navigate  Enter: attach  n: new
r: rename  x: kill  q/Esc: quit
tip: Ctrl-b d to detach from tmux
```

`tp` is a shortcut for `tmux-pilot`, installed automatically.

## Why

tmux is essential for running long-lived processes on remote servers — coding agents, builds, anything you want to survive closing your laptop. But the commands are hard to remember. `tmux ls`, `tmux attach -t`, `tmux kill-session -t`... every time.

tmux-pilot gives you a simple picker. No commands to memorize. Works reliably over SSH.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/blockful/tmux-pilot/main/install.sh | bash
```

This installs `tmux-pilot` and the `tp` shortcut to `~/.local/bin`.

### Other methods

<details>
<summary>Homebrew</summary>

```bash
brew install blockful/tap/tmux-pilot
```
</details>

<details>
<summary>Go</summary>

```bash
go install github.com/blockful/tmux-pilot/cmd@latest
```
</details>

<details>
<summary>Download binary</summary>

Grab the latest binary from [Releases](https://github.com/blockful/tmux-pilot/releases).
</details>

## Update

```bash
curl -fsSL https://raw.githubusercontent.com/blockful/tmux-pilot/main/install.sh | bash
```

Same command as install. It detects your current version and skips if already up to date.

## Uninstall

```bash
curl -fsSL https://raw.githubusercontent.com/blockful/tmux-pilot/main/install.sh | bash -s -- --uninstall
```

## Usage

```bash
tp
```

That's it. You see your sessions, you pick one.

### Keybindings

| Key | Action |
|-----|--------|
| `↑`/`k` | Navigate up |
| `↓`/`j` | Navigate down |
| `Enter` | Switch to session |
| `n` | Create new session |
| `r` | Rename session |
| `x` | Kill session (stays in picker) |
| `q`/`Esc` | Quit |

**tip:** `Ctrl-b d` to detach from tmux (this is a tmux command, not tmux-pilot)

### Flags

| Flag | Description |
|------|-------------|
| `-h`, `--help` | Show help |
| `-v`, `--version` | Show version |
| `-l`, `--list` | Print sessions as TSV (pipe-friendly) |
| `-S <path>` | Custom tmux socket path |
| `-L <name>` | Named tmux socket |
| `--no-color` | Disable colors (also respects `NO_COLOR` env) |

### Examples

```bash
$ tp                        # open the picker
$ tp --list                 # non-interactive, TSV output
$ tp -S /tmp/custom.sock    # use a specific tmux socket
$ tp --version              # check installed version
```

## How it works

- ~1000 lines of Go
- One dependency: `golang.org/x/term` (Go extended stdlib)
- Raw ANSI rendering — no TUI framework
- In-place drawing — no alt screen escape sequences
- Works cleanly through SSH → tmux layers

## License

MIT
