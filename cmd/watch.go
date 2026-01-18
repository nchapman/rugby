package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/nchapman/rugby/internal/builder"
	"github.com/nchapman/rugby/internal/watcher"
)

var (
	// Header styles
	watchHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("14")).
				Bold(true)
	watchFileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Bold(true)
	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true)

	// Status styles
	buildingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11"))
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true)
	failStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)
	outputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))

	// Info styles
	changeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12"))
	timeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
)

var watchCmd = &cobra.Command{
	Use:   "watch <file.rg> [-- args...]",
	Short: "Watch for changes and re-run",
	Long: `Watches for file changes and automatically recompiles and runs the program.

Pass arguments to the program after --:
  rugby watch main.rg -- --port 8080

The console is cleared between runs for a fresh view.
Press Ctrl+C to stop watching.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		file := args[0]

		if err := ValidateRgFile(file); err != nil {
			return err
		}

		// Arguments after -- go to the program
		var programArgs []string
		if dashIndex := cmd.ArgsLenAtDash(); dashIndex >= 0 {
			programArgs = args[dashIndex:]
		}

		return watchFile(file, programArgs)
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
}

// watchFile sets up file watching and runs the file on changes.
func watchFile(file string, args []string) error {
	project, err := builder.FindProjectFrom(file)
	if err != nil {
		return err
	}

	w, err := watcher.New(project.Root)
	if err != nil {
		return err
	}
	defer func() { _ = w.Close() }()

	// Watch the project root for .rg and .go files
	if err := w.AddPath(project.Root); err != nil {
		return fmt.Errorf("failed to watch directory: %w", err)
	}

	// Setup context for clean shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)
	go func() {
		<-sigChan
		cancel()
	}()

	// Start the watch loop
	events, errors := w.Watch(ctx)

	// Create runner to manage subprocess
	runner := newProcessRunner()

	// Initial run
	watcher.ClearScreen()
	printWatchHeader(file)
	runner.runOnce(project, file, args)

	// Watch loop
	for {
		select {
		case <-ctx.Done():
			runner.stop()
			fmt.Println()
			fmt.Println(watchHeaderStyle.Render("Stopped watching."))
			return nil

		case event, ok := <-events:
			if !ok {
				return nil
			}
			// Skip changes in .rugby directory
			if strings.Contains(event.Path, ".rugby") {
				continue
			}

			watcher.ClearScreen()
			printWatchHeader(file)

			relPath, _ := filepath.Rel(project.Root, event.Path)
			fmt.Printf("%s %s\n\n",
				changeStyle.Render("Changed:"),
				relPath)

			runner.runOnce(project, file, args)

		case watchErr, ok := <-errors:
			if !ok {
				return nil
			}
			logger.Error("watch error", "err", watchErr)
		}
	}
}

// printWatchHeader prints the watch mode header.
func printWatchHeader(file string) {
	fmt.Printf("%s %s\n",
		watchHeaderStyle.Render("watching"),
		watchFileStyle.Render(file))
	fmt.Println(hintStyle.Render("Press Ctrl+C to stop"))
	fmt.Println(separatorStyle.Render(strings.Repeat("─", 40)))
	fmt.Println()
}

// buildSpinner manages an animated spinner during builds.
type buildSpinner struct {
	spinner spinner.Model
	stop    chan struct{}
	done    chan struct{}
}

func newBuildSpinner() *buildSpinner {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = buildingStyle
	return &buildSpinner{
		spinner: s,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
}

func (s *buildSpinner) start(message string) {
	go func() {
		defer close(s.done)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-s.stop:
				// Clear the spinner line
				fmt.Print("\r\033[K")
				return
			case <-ticker.C:
				s.spinner, _ = s.spinner.Update(s.spinner.Tick())
				fmt.Printf("\r%s %s", s.spinner.View(), buildingStyle.Render(message))
			}
		}
	}()
}

func (s *buildSpinner) finish() {
	close(s.stop)
	<-s.done
}

// processRunner manages the running subprocess with proper synchronization.
type processRunner struct {
	mu  sync.Mutex
	cmd *exec.Cmd
}

func newProcessRunner() *processRunner {
	return &processRunner{}
}

// stop terminates the current process if running.
func (r *processRunner) stop() {
	r.mu.Lock()
	cmd := r.cmd
	r.mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return
	}

	// Try graceful shutdown first - signal the process group if available
	pid := cmd.Process.Pid
	pgid, pgidErr := syscall.Getpgid(pid)
	if pgidErr == nil && pgid > 0 {
		// Signal the entire process group
		_ = syscall.Kill(-pgid, syscall.SIGTERM)
	} else {
		// Fall back to signaling just the process
		_ = cmd.Process.Signal(syscall.SIGTERM)
	}

	// Wait for process to exit with timeout
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		// Force kill if still running
		if pgidErr == nil && pgid > 0 {
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		} else {
			_ = cmd.Process.Kill()
		}
		<-done
	}

	r.mu.Lock()
	r.cmd = nil
	r.mu.Unlock()
}

// runOnce compiles and runs the file, stopping any previous process.
func (r *processRunner) runOnce(project *builder.Project, file string, args []string) {
	r.stop()

	startTime := time.Now()

	b := builder.New(project, builder.WithVerbose(verbose), builder.WithColorMode(getColorMode()))

	// Start spinner
	spin := newBuildSpinner()
	spin.start("Building...")

	// Compile
	result, err := b.Compile([]string{file})
	if err != nil {
		spin.finish()
		fmt.Println(failStyle.Render("Build failed"))
		fmt.Println()
		fmt.Println(err)
		printWaitingForChanges()
		return
	}

	// Build binary
	base := filepath.Base(file)
	binName := strings.TrimSuffix(base, filepath.Ext(base))
	binPath := project.BinPath(binName)

	if err := b.GoBuild(result.GenFiles, binPath); err != nil {
		spin.finish()
		fmt.Println(failStyle.Render("Build failed"))
		fmt.Println()
		fmt.Println(err)
		printWaitingForChanges()
		return
	}

	spin.finish()

	elapsed := time.Since(startTime)
	fmt.Printf("%s %s\n\n",
		successStyle.Render("Built"),
		timeStyle.Render(fmt.Sprintf("in %dms", elapsed.Milliseconds())))

	// Run
	fmt.Println(outputStyle.Render("─── Output ───"))
	fmt.Println()

	cmd := exec.Command(binPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set process group so we can terminate all child processes
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		fmt.Println(failStyle.Render("Run error:"), err)
		return
	}

	r.mu.Lock()
	r.cmd = cmd
	r.mu.Unlock()

	// Wait in background - we don't block here
	go func() {
		waitErr := cmd.Wait()
		if waitErr != nil {
			if exitErr, ok := waitErr.(*exec.ExitError); ok {
				fmt.Println()
				fmt.Println(failStyle.Render(fmt.Sprintf("Exited with code %d", exitErr.ExitCode())))
			}
		} else {
			fmt.Println()
			fmt.Println(successStyle.Render("Done"))
		}
		printWaitingForChanges()
	}()
}

func printWaitingForChanges() {
	fmt.Println()
	fmt.Println(hintStyle.Render("Waiting for changes..."))
}
