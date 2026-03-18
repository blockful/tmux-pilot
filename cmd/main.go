package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/blockful/tmux-pilot/internal/tmux"
	"github.com/blockful/tmux-pilot/internal/tui"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// Config holds parsed command line configuration.
type Config struct {
	Help      bool
	Version   bool
	List      bool
	NoColor   bool
	ClientOpts tmux.ClientOptions
}

func main() {
	config, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	// Handle help flag
	if config.Help {
		printHelp()
		os.Exit(0)
	}

	// Handle version flag
	if config.Version {
		fmt.Printf("tmux-pilot %s (%s, %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Fetch sessions
	sessions, err := tmux.ListSessions(config.ClientOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Handle list mode
	if config.List {
		printSessionsTSV(sessions)
		os.Exit(0)
	}

	// Run TUI
	action, err := tui.Run(sessions, !config.NoColor, config.ClientOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Execute action
	if err := execute(action, config.ClientOpts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// parseArgs manually parses command line arguments.
func parseArgs(args []string) (Config, error) {
	config := Config{}
	
	for i := 0; i < len(args); i++ {
		arg := args[i]
		
		switch arg {
		case "-h", "--help":
			config.Help = true
		case "-v", "--version":
			config.Version = true
		case "-l", "--list":
			config.List = true
		case "--no-color":
			config.NoColor = true
		case "-S":
			if i+1 >= len(args) {
				return Config{}, fmt.Errorf("flag -S requires an argument")
			}
			if config.ClientOpts.SocketName != "" {
				return Config{}, fmt.Errorf("cannot specify both -S and -L flags")
			}
			config.ClientOpts.SocketPath = args[i+1]
			if config.ClientOpts.SocketPath == "" {
				return Config{}, fmt.Errorf("socket path cannot be empty")
			}
			i++ // skip next argument
		case "-L":
			if i+1 >= len(args) {
				return Config{}, fmt.Errorf("flag -L requires an argument")
			}
			if config.ClientOpts.SocketPath != "" {
				return Config{}, fmt.Errorf("cannot specify both -S and -L flags")
			}
			config.ClientOpts.SocketName = args[i+1]
			if config.ClientOpts.SocketName == "" {
				return Config{}, fmt.Errorf("socket name cannot be empty")
			}
			i++ // skip next argument
		default:
			if strings.HasPrefix(arg, "-") {
				return Config{}, fmt.Errorf("unknown flag: %s", arg)
			}
			return Config{}, fmt.Errorf("unexpected argument: %s", arg)
		}
	}
	
	return config, nil
}

// printHelp displays usage information.
func printHelp() {
	fmt.Print(`tmux-pilot - Interactive tmux session manager

USAGE:
    tmux-pilot [flags]

FLAGS:
    -h, --help         Show this help and exit
    -v, --version      Show version and exit
    -l, --list         Print sessions as TSV and exit
    -S <socket-path>   Use custom tmux socket path
    -L <socket-name>   Use named tmux socket
        --no-color     Disable ANSI colors

KEYBINDINGS:
    ↑/k ↓/j    Navigate sessions
    Enter      Switch to selected session
    n          Create new session
    r          Rename selected session
    x          Kill selected session (stays in picker)
    q/Esc      Quit without action

    tip: Ctrl-b d to detach from tmux

EXAMPLES:
    tmux-pilot                     # Basic usage
    tmux-pilot -S /tmp/custom      # Custom socket path
    tmux-pilot -L dev              # Named socket
    tmux-pilot --list              # Non-interactive TSV output
    NO_COLOR=1 tmux-pilot          # Disable colors via env

ENVIRONMENT:
    NO_COLOR    Set to any value to disable colors
`)
}

// printSessionsTSV prints sessions in TSV format.
func printSessionsTSV(sessions []tmux.Session) {
	for _, session := range sessions {
		status := "detached"
		if session.Attached {
			status = "attached"
		}
		fmt.Printf("%s\t%d\t%s\n", session.Name, session.WindowCount, status)
	}
}

// execute runs the action chosen by the user.
// Only "switch" exits the TUI — kill, rename, and create are handled
// inline while the picker stays open.
func execute(a tui.Action, opts tmux.ClientOptions) error {
	if a.Kind == "switch" {
		return tmux.SwitchOrAttach(a.Target, opts)
	}
	return nil // user quit without action
}