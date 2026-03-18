# Business Rules

## Session Picker

- The picker shows all tmux sessions with name, window count, and attached/detached status
- `●` (green) = attached, `○` (dim) = detached
- Navigation wraps at boundaries (cursor stays at first/last item)
- Empty session list shows "No tmux sessions running"

## Actions That Stay in Picker

These execute the tmux command and refresh the session list without exiting:

| Action | Key | Behavior |
|--------|-----|----------|
| Create | `n` → type name → `Enter` | Creates detached session, refreshes list |
| Rename | `r` → edit name → `Enter` | Renames session, refreshes list |
| Kill | `x` → `y` to confirm | Kills session, refreshes list |

On error, a warning is shown and the picker stays open.

## Actions That Exit

| Action | Key | Behavior |
|--------|-----|----------|
| Switch | `Enter` | Exits picker, runs `tmux switch-client` or `attach-session` |
| Quit | `q` / `Esc` / `Ctrl-C` | Exits picker, no action |

## Session Names

- Cannot be empty
- Cannot duplicate an existing session name (warning shown)
- Rename pre-fills the current name

## CLI Flags

- `-S` and `-L` are mutually exclusive (error if both specified)
- `-S`/`-L` values cannot be empty
- Unknown flags → error, exit 1
- `--help` and `--version` take precedence over other flags
- `--list` bypasses the TUI entirely, prints TSV to stdout

## List Mode Output

Tab-separated: `name\twindow_count\tattached|detached`

```
main	3	attached
api	1	detached
```

Empty session list = no output, exit 0.

## Color

Disabled when any of:
- `NO_COLOR` env var is set (to any value, including empty)
- `--no-color` flag is passed
- stdout is not a TTY (piped/redirected)

## Terminal Behavior

- Never use alt screen (`\e[?1049h/l`)
- Always restore terminal state on exit, even on Ctrl-C or SIGTERM
- Use `\r\n` for newlines (raw mode doesn't auto-CR)
- Hide cursor during render, show on exit
