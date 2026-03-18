# ADR 001: Drop BubbleTea for Raw ANSI Implementation

## Status
Accepted

## Context
tmux-pilot currently uses the Charm BubbleTea framework for TUI rendering. In nested tmux environments (Mac → SSH → Linux → tmux → tmux-pilot), this causes critical reliability issues:

1. **Alt screen leakage**: `tea.WithAltScreen()` escape sequences (`\e[?1049h/l`) leak through multiple PTY layers, causing blank screens and phantom newlines after exit
2. **Input races**: `tea.WithInputTTY()` opens `/dev/tty` directly, racing with tmux's own PTY input handling, resulting in dropped keystrokes  
3. **Framework overhead**: BubbleTea is 14 transitive dependencies for a simple session picker
4. **Complex debugging**: Framework abstractions make PTY issues harder to diagnose

The user environment is: `Mac → SSH → Linux → tmux → shell → tmux-pilot`

## Options Considered

### Option 1: Fix BubbleTea Issues
**Approach**: Disable alt screen, configure input handling differently
**Pros**: 
- Minimal code changes
- Leverage existing framework features
**Cons**:
- Still 14 dependencies for simple functionality
- Limited control over escape sequence generation
- May hit other PTY issues in framework
- Framework designed for alt screen mode

### Option 2: Switch to Different TUI Framework  
**Approach**: Replace BubbleTea with tcell, tview, or similar
**Pros**:
- Potentially better PTY compatibility
- More established frameworks
**Cons**:
- Dependency bloat remains
- Same fundamental issues with framework abstractions
- Still complex for a simple picker

### Option 3: Raw ANSI + golang.org/x/term (CHOSEN)
**Approach**: Direct terminal control with minimal dependencies
**Pros**:
- Complete control over every escape sequence
- Single dependency: golang.org/x/term (part of Go's extended stdlib)
- In-place rendering, no alt screen needed
- Direct stdin/stdout, no separate TTY access
- Easier to debug and reason about
- Matches project philosophy: minimal, reliable tools
**Cons**:
- More code to write and maintain
- No framework conveniences  
- Need to handle terminal edge cases manually

## Decision

**Option 3: Raw ANSI + golang.org/x/term**

We will completely rewrite `internal/tui/` to use direct ANSI escape sequences with `golang.org/x/term` for raw mode control.

## Rationale

1. **Reliability first**: The current PTY issues are blocking real usage. Framework abstractions introduce too many variables in the PTY stack.

2. **Dependency minimalism**: A session picker doesn't justify 14 transitive dependencies. One focused dependency (golang.org/x/term) is appropriate.

3. **Control and debuggability**: Direct escape sequence control means we know exactly what's being sent to the terminal. PTY issues become traceable.

4. **Appropriate complexity**: The required TUI functionality (list, navigate, select, input) maps well to direct terminal control. No complex layouts or widgets needed.

5. **Long-term maintainability**: Raw ANSI is stable (decades old). golang.org/x/term is part of Go's extended standard library.

## Implementation Approach

### Phase 1: Preserve Existing Interface
- Keep `tui.Run(sessions []Session) (Action, error)` interface
- Maintain all existing functionality and keybindings
- Pass all current tests

### Phase 2: Raw Terminal Control
```go
// Terminal state management
type Terminal struct {
    originalState *term.State
    fd int
}

func (t *Terminal) EnterRawMode() error
func (t *Terminal) Restore() error  
func (t *Terminal) ReadKey() (Key, error)
```

### Phase 3: Direct ANSI Rendering
```go
// Direct stdout writes with ANSI codes
func MoveCursor(x, y int)     // \e[<y>;<x>H
func ClearLine()              // \e[2K
func SetColor(code int)       // \e[38;5;<code>m
func HideCursor()             // \e[?25l
func ShowCursor()             // \e[?25h
```

### Phase 4: Signal Handling
```go
// Always restore terminal state
func setupSignalHandlers(term *Terminal) {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        term.Restore()
        os.Exit(130) // Standard SIGINT exit code
    }()
}
```

## Risks and Mitigation

### Risk: Terminal Compatibility
**Mitigation**: Use standard ANSI escape sequences supported by all modern terminals. Test on xterm, tmux, screen.

### Risk: Increased Code Complexity
**Mitigation**: 
- Well-defined component boundaries (Terminal, Renderer, InputHandler, StateMachine)
- Comprehensive test coverage including edge cases
- Clear documentation of ANSI sequences used

### Risk: Missing Framework Features
**Mitigation**: We only need basic functionality that maps well to ANSI control. No complex widgets or layouts required.

## Success Metrics

1. **Reliability**: Zero escape sequence leakage in nested tmux environments
2. **Performance**: <50ms startup time, responsive input processing  
3. **Dependencies**: Single external dependency (golang.org/x/term)
4. **Compatibility**: Works in all tested terminal environments
5. **Maintainability**: Code is clear, well-tested, and easy to debug

## Implementation Notes

- Original terminal state restoration is critical - use defer and signal handlers
- Color mode must respect NO_COLOR environment variable  
- Input processing must handle all terminal escape sequence variations
- Memory usage should be bounded and minimal
- Integration tests with isolated tmux sockets (`-S /tmp/test-socket`)