# Business Rules: tmux-pilot

## CLI Argument Processing

### Flag Parsing Rules
1. **Manual parsing only** - No external flag libraries. Zero deps except golang.org/x/term.
2. **Help flag precedence** - `--help` or `-h` prints usage and exits with code 0, regardless of other flags.
3. **Version flag precedence** - `--version` or `-v` prints version and exits with code 0, after help precedence.
4. **Socket flags are exclusive** - Cannot specify both `-S` and `-L`. Exit with error code 1.
5. **Socket path validation** - `-S` argument must be non-empty. `-L` argument must be non-empty.
6. **List mode exclusivity** - `--list`/`-l` bypasses TUI entirely. Print TSV and exit 0.

### Flag Format Rules  
```
tmux-pilot [flags]

Flags:
  -h, --help         Show usage and exit
  -v, --version      Show version and exit  
  -l, --list         Print sessions as TSV and exit
  -S <socket-path>   Use custom tmux socket path
  -L <socket-name>   Use named tmux socket
      --no-color     Disable ANSI colors
```

### Argument Validation
- Unknown flags → error code 1
- Missing values for `-S`/`-L` → error code 1  
- Extra positional arguments → error code 1

## Exit Code Rules

### Exit Code 0 (Success)
- User completed action successfully (switch, new, rename, kill, detach)
- User quit without action (q, esc)
- Help or version flag used
- List mode completed successfully

### Exit Code 1 (Error)  
- tmux not found in PATH
- tmux server not running and command failed
- Invalid session name for operations
- Socket path/name doesn't exist
- Terminal setup failed
- Argument parsing errors

### Exit Code 130 (SIGINT)
- User pressed Ctrl-C
- SIGINT received
- Terminal state must be restored before exit

## Color Mode Rules

### Color Disabled When:
1. `NO_COLOR` environment variable is set (to any value, including empty string)
2. `--no-color` flag is specified
3. Output is not a TTY (pipe/redirect detection)

### Color Enabled When:
- None of the above conditions are true
- stdout is a TTY

### ANSI Color Codes (when enabled)
- Accent: 205 (bright magenta)
- Text: 252 (light gray) 
- Dim: 243 (dark gray)
- Border: 240 (darker gray)
- Green: 46 (bright green)
- Yellow: 214 (orange-yellow)
- Background: 237 (dark background)

## TUI Behavior Rules

### Terminal Control
1. **Never use alt screen** - No `\e[?1049h/l` sequences
2. **Raw mode required** - stdin must be in raw mode for key processing  
3. **Single I/O streams** - stdin for input, stdout for output. No `/dev/tty` access.
4. **Graceful restoration** - Original terminal state restored on ANY exit path

### Input Processing Rules  
```
List Mode:
  ↑/k     Move cursor up (bounded at 0)
  ↓/j     Move cursor down (bounded at session count - 1)  
  Enter   Switch to selected session
  n       Enter create mode
  r       Enter rename mode (if sessions exist)
  x       Enter kill confirmation (if sessions exist)
  d       Detach current client  
  q/Esc   Quit without action

Create Mode:
  [char]  Append to session name
  Bksp    Remove last character
  Enter   Create session (if name non-empty and unique)
  Esc     Cancel, return to list mode

Rename Mode:  
  [char]  Append to new name
  Bksp    Remove last character  
  Enter   Rename session (if name non-empty and unique)
  Esc     Cancel, return to list mode

Kill Confirmation:
  y/Enter Kill session
  n/q/Esc Cancel, return to list mode
```

### Session Name Validation
1. **Non-empty rule** - Session names cannot be empty strings
2. **Uniqueness rule** - New/renamed sessions cannot have duplicate names
3. **Character restrictions** - Follow tmux session name rules (no special chars that break tmux)

### Visual Layout Rules
1. **Minimum width** - 50 columns minimum, or terminal width - 4 (for border padding)
2. **Border style** - Rounded border characters: ╭─╮│ ╰─╯
3. **Selected item** - Background highlight + cursor indicator  
4. **Status indicators** - ● for attached sessions, ○ for detached
5. **Information display** - `name    N windows   attached/detached`

### Error Display Rules
1. **Warning messages** - Show inline warnings for duplicate names, validation errors
2. **No flash/beep** - Silent error indication via text only
3. **Persistent warnings** - Warnings persist until user takes corrective action

## tmux Integration Rules

### Socket Passing
1. **Thread through all commands** - Socket options (-S/-L) must be passed to every tmux exec
2. **Command precedence** - Socket flags appear before tmux subcommand: `tmux -S /path list-sessions`
3. **Validation** - Socket paths/names validated before passing to tmux

### Session Operations
```
Switch/Attach:
  Inside tmux:  tmux switch-client -t <name>  
  Outside tmux: tmux attach-session -t <name>

New Session:
  tmux new-session -d -s <name>
  Then switch/attach to it

Rename:
  tmux rename-session -t <old> <new>

Kill:
  tmux kill-session -t <name>  

Detach:
  tmux detach-client
```

### Error Handling Rules
1. **No tmux server** - Handle gracefully, show "No sessions" if no server running
2. **Session not found** - Propagate tmux error messages to user  
3. **Permission denied** - Show clear error about socket permissions
4. **Invalid session names** - Validate before passing to tmux to avoid command injection

## List Mode Output Format

### TSV Format Rules
```
<name>\t<window_count>\t<status>\n
```

Where:
- `name`: Session name (no escaping, assume clean)
- `window_count`: Integer number of windows  
- `status`: Either "attached" or "detached"

### Output Examples
```
main	3	attached
api-server	1	detached  
notes	2	detached
```

### Edge Cases
- Empty session list: No output (empty stdout)
- No tmux server: No output, exit 0 (consistent with tmux list-sessions behavior)