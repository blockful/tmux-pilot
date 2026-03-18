# Architecture

## Overview

tmux-pilot is a terminal picker for tmux sessions. ~1000 lines of Go, one dependency (`golang.org/x/term`).

## Project Structure

```
cmd/main.go              Entry point: flag parsing, orchestration
internal/tmux/            tmux interaction layer
  ├── client.go           All tmux commands (list, switch, kill, etc.)
  ├── client_test.go      Tests
  └── types.go            Session struct, ClientOptions
internal/tui/             Terminal UI (raw ANSI, no framework)
  ├── terminal.go         Raw mode, signal handling
  ├── input.go            Keystroke reading, escape sequence parsing
  ├── render.go           ANSI output, colors, screen drawing
  ├── picker.go           Main loop, state machine, event handling
  ├── input_test.go
  ├── picker_test.go
  └── render_test.go
docs/                     This documentation
install.sh                curl installer / updater / uninstaller
```

## Data Flow

```
User runs `tp`
  │
  ├─ Parse CLI flags (manual, no library)
  ├─ tmux.ListSessions() → []Session
  │
  ├─ [--list mode] → print TSV, exit
  │
  ├─ [interactive mode]
  │    ├─ terminal.EnterRawMode()
  │    ├─ Loop:
  │    │    ├─ renderer.RenderUI() → draw in-place
  │    │    ├─ input.ReadKey() → wait for keystroke
  │    │    └─ state.handleKey() → update state or execute tmux command
  │    │         ├─ kill/rename/create → execute inline, refresh list, stay in picker
  │    │         └─ switch/quit → exit loop
  │    ├─ renderer.Cleanup() → erase drawn lines
  │    └─ terminal.Restore() → restore original terminal state
  │
  └─ execute action (switch-client or nothing)
```

## Key Design Decisions

### No TUI framework

Previous version used BubbleTea. It broke inside tmux over SSH — alt screen leaks, input races, phantom newlines. Replaced with raw ANSI escape sequences. We control every byte.

See: [decisions/001-drop-bubbletea-for-raw-ansi.md](decisions/001-drop-bubbletea-for-raw-ansi.md)

### No alt screen

Alt screen (`\e[?1049h`) saves/restores the terminal buffer. Through SSH → tmux layers, the restore sequence often leaks, leaving blank screens. Instead we draw in-place and erase our lines on exit.

### Inline execution

Kill, rename, and create don't exit the picker. They execute the tmux command, refresh the session list, and stay open. Only "switch to session" and "quit" exit the TUI.

### Raw mode input

We read from `os.Stdin` in raw mode (`golang.org/x/term`). No opening `/dev/tty` separately — that races with tmux's PTY. One input source, one output destination.

### `\r\n` not `\n`

In raw terminal mode, `\n` moves down but doesn't return to column 1. All output uses `\r\n` to avoid staircase rendering.

### Socket threading

All tmux functions accept `ClientOptions` with optional `-S`/`-L` socket config. This threads through every `exec.Command("tmux", ...)` call, enabling isolated tmux servers for both user workflows and testing.

## Terminal Rendering

The renderer tracks how many lines it drew last frame (`lastHeight`). On re-render:
1. Move cursor up `lastHeight` lines
2. Clear and redraw each line
3. If the new frame is shorter, clear leftover lines

On exit, `Cleanup()` erases all drawn lines and restores cursor position.

## Color

Colors use ANSI 256-color codes. Disabled when:
- `NO_COLOR` env var is set (any value)
- `--no-color` flag
- stdout is not a TTY

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success or user quit |
| 1 | Error (bad args, tmux failure) |
| 130 | SIGINT / Ctrl-C |

## Testing

- **Unit tests**: key parsing, ANSI output, state machine transitions, color modes
- **Integration tests**: spin up isolated tmux servers (`-S /tmp/test.sock`), create/kill/rename real sessions, verify outcomes
- No mocks for tmux — tests use real tmux with isolated sockets
