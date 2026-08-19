package engine

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ProcessManager struct {
	mu             sync.Mutex
	processes      map[string]*exec.Cmd
	userTerminated map[string]bool
}

var (
	globalRunner = &ProcessManager{
		processes:      make(map[string]*exec.Cmd),
		userTerminated: make(map[string]bool),
	}
)

func GetProcessManager() *ProcessManager {
	return globalRunner
}

type circularBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (c *circularBuffer) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.buf.Len() > 64*1024 {
		// Keep last 32KB
		data := c.buf.Bytes()
		c.buf.Reset()
		c.buf.Write(data[len(data)-32*1024:])
	}
	return c.buf.Write(p)
}

func (c *circularBuffer) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.String()
}

// Launch starts the engine executable with projectDir as current working directory and captures logs
func (pm *ProcessManager) Launch(executablePath, projectDir string, onExit func(err error, userTerminated bool, outputTail string, logFilePath string)) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, err := os.Stat(executablePath); os.IsNotExist(err) {
		return fmt.Errorf("executable not found at %s", executablePath)
	}

	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return fmt.Errorf("project directory not found at %s", projectDir)
	}

	// Check if already running for this project
	if cmd, exists := pm.processes[projectDir]; exists && cmd.Process != nil {
		return fmt.Errorf("game is already running for this project (PID: %d)", cmd.Process.Pid)
	}

	// Create project log file
	logsDir := filepath.Join(projectDir, "save", "logs")
	_ = os.MkdirAll(logsDir, 0755)
	logFilePath := filepath.Join(logsDir, "ikemen-latest.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		// Fallback to non-file logging if log file can't be created
		logFile = nil
	}

	recentOut := &circularBuffer{}

	var stdoutWriter io.Writer = os.Stdout
	var stderrWriter io.Writer = os.Stderr

	if logFile != nil {
		header := fmt.Sprintf("=== Ikemen GO Session Log - %s ===\nExecutable: %s\nProject: %s\n\n", time.Now().Format(time.RFC3339), executablePath, projectDir)
		_, _ = logFile.WriteString(header)
		stdoutWriter = io.MultiWriter(os.Stdout, logFile, recentOut)
		stderrWriter = io.MultiWriter(os.Stderr, logFile, recentOut)
	} else {
		stdoutWriter = io.MultiWriter(os.Stdout, recentOut)
		stderrWriter = io.MultiWriter(os.Stderr, recentOut)
	}

	cmd := exec.Command(executablePath)
	cmd.Dir = projectDir
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter

	if err := cmd.Start(); err != nil {
		if logFile != nil {
			logFile.Close()
		}
		return fmt.Errorf("failed to start engine process: %w", err)
	}

	pm.processes[projectDir] = cmd
	pm.userTerminated[projectDir] = false

	go func() {
		err := cmd.Wait()
		if logFile != nil {
			_ = logFile.Sync()
			_ = logFile.Close()
		}

		pm.mu.Lock()
		wasKilled := pm.userTerminated[projectDir]
		delete(pm.processes, projectDir)
		delete(pm.userTerminated, projectDir)
		pm.mu.Unlock()

		output := recentOut.String()
		if onExit != nil {
			onExit(err, wasKilled, output, logFilePath)
		}
	}()

	return nil
}

// IsRunning checks whether an engine instance is active for the given project
func (pm *ProcessManager) IsRunning(projectDir string) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	_, exists := pm.processes[projectDir]
	return exists
}

// Stop terminates the running engine instance for the given project
func (pm *ProcessManager) Stop(projectDir string) error {
	pm.mu.Lock()
	cmd, exists := pm.processes[projectDir]
	if exists {
		pm.userTerminated[projectDir] = true
	}
	pm.mu.Unlock()

	if !exists || cmd.Process == nil {
		return fmt.Errorf("no running process for project %s", projectDir)
	}

	return cmd.Process.Kill()
}

// ExtractErrorSummary extracts the most relevant error/panic line from raw log output
func ExtractErrorSummary(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var errorLines []string

	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		lower := strings.ToLower(trimmed)
		if strings.Contains(lower, "panic:") ||
			strings.Contains(lower, "lua open failed:") ||
			strings.Contains(lower, "error:") ||
			strings.Contains(lower, "failed to open") ||
			strings.Contains(lower, "fatal:") ||
			strings.Contains(lower, "runtime error:") {
			errorLines = append(errorLines, trimmed)
		}
	}

	if len(errorLines) > 0 {
		return strings.Join(errorLines, "\n")
	}

	// Fallback to the last 3 non-empty lines
	var lastLines []string
	for i := len(lines) - 1; i >= 0 && len(lastLines) < 3; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			lastLines = append([]string{strings.TrimSpace(lines[i])}, lastLines...)
		}
	}

	if len(lastLines) > 0 {
		return strings.Join(lastLines, "\n")
	}

	return "Unknown engine error. Check session log for details."
}
