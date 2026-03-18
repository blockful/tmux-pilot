# Architecture: Raw ANSI TUI Rewrite

## Overview

Complete rewrite of `internal/tui/` to replace BubbleTea framework with raw ANSI escape sequences and `golang.org/x/term` for terminal control. Zero external dependencies except `golang.org/x/term`.

## Problem Statement

Current BubbleTea implementation has reliability issues in nested tmux environments:
- Alt screen escape sequences leak through PTY layers over SSH
- `tea.WithInputTTY()` opens `/dev/tty` separately, racing with tmux's PTY handling
- Framework overhead for a simple session picker
- Dependency bloat (14 transitive dependencies)

## Solution Architecture

### Core Principles
1. **Single I/O stream**: Read stdin directly, write to stdout. No separate `/dev/tty` access.
2. **In-place rendering**: No alt screen mode. Draw in current terminal, clean up on exit.
3. **Direct ANSI control**: Manual escape sequences for cursor movement, colors, line clearing.
4. **Graceful cleanup**: Always restore terminal state, even on panic/signals.

### Component Boundaries

```
cmd/main.go
├── Parse CLI flags (manual parsing)
├── Handle --list mode (bypass TUI)
└── Create TUI → Execute actions

internal/tmux/
├── ClientOptions struct (socket config)
├── All functions accept ClientOptions
└── Thread socket flags through exec.Command

internal/tui/ (COMPLETE REWRITE)
├── Terminal: raw mode + ANSI control
├── Renderer: direct stdout writes  
├── InputHandler: stdin key processing
└── StateMachine: mode transitions
```

### New CLI Interface

```bash
# Basic usage
tmux-pilot

# Custom sockets (threaded to tmux commands)
tmux-pilot -S /tmp/custom-socket
tmux-pilot -L session-name

# Non-interactive TSV output  
tmux-pilot --list
# Output: name\twindows\tattached/detached

# Flags
tmux-pilot --help        # usage + keybindings + examples
tmux-pilot --version     # version info
tmux-pilot --no-color    # disable colors
NO_COLOR=1 tmux-pilot    # also disables colors
```

### Exit Codes
- `0`: Success or user quit
- `1`: Error (tmux not found, invalid args, etc)
- `130`: SIGINT (Ctrl-C)

## Data Flow

### Interactive Mode
```
1. Parse flags → ClientOptions
2. tmux.ListSessions(opts) → []Session
3. tui.Run(sessions, colorMode) → Action  
4. execute(action, opts) → tmux commands
```

### List Mode (--list)
```
1. Parse flags → ClientOptions  
2. tmux.ListSessions(opts) → []Session
3. Print TSV to stdout
4. Exit 0
```

## Implementation Plan

### 1. Create ClientOptions in internal/tmux/
```go
type ClientOptions struct {
    SocketPath string // -S flag
    SocketName string // -L flag  
}

func (o ClientOptions) Args() []string {
    var args []string
    if o.SocketPath != "" {
        args = append(args, "-S", o.SocketPath)
    }
    if o.SocketName != "" {
        args = append(args, "-L", o.SocketName)  
    }
    return args
}
```

### 2. Update all tmux functions
```go
func ListSessions(opts ClientOptions) ([]Session, error)
func SwitchOrAttach(name string, opts ClientOptions) error
// etc...
```

### 3. Rewrite internal/tui/ completely
```go
// terminal.go - raw mode control
type Terminal struct {
    originalState *term.State
    width, height int
}

func (t *Terminal) EnterRawMode() error
func (t *Terminal) Restore() error
func (t *Terminal) Size() (int, int)

// renderer.go - ANSI output  
type Renderer struct {
    colorMode ColorMode
}

func (r *Renderer) MoveCursor(x, y int)
func (r *Renderer) ClearLine()  
func (r *Renderer) SetColor(code int)

// input.go - key processing
func ReadKey() (Key, error)

// tui.go - main TUI controller
func Run(sessions []Session, opts TUIOptions) (Action, error)
```

### 4. Manual flag parsing in cmd/main.go
```go
func parseArgs(args []string) (Config, error) {
    // Manual parsing, no flag library
    // Return struct with all options
}
```

### 5. Update tests
- Unit tests for new TUI components
- Integration tests with `-S /tmp/test-socket`  
- Signal handling tests
- Color mode tests

## Security Considerations

- Terminal state restoration is critical - any panic must restore
- Input validation on session names (tmux injection prevention)
- Bounded memory usage (session list size)

## Performance Requirements

- Startup time: <50ms on commodity hardware
- Memory usage: <5MB resident  
- Responsive input: <16ms key processing (60fps equivalent)

## Compatibility

- Terminals: xterm, tmux, screen, modern terminal emulators
- OS: Linux, macOS, BSDs (anywhere golang.org/x/term works)
- tmux: 2.0+ (modern versions)