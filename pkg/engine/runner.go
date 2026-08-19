package engine

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
)

type ProcessManager struct {
	mu        sync.Mutex
	processes map[string]*exec.Cmd
}

var (
	globalRunner = &ProcessManager{
		processes: make(map[string]*exec.Cmd),
	}
)

func GetProcessManager() *ProcessManager {
	return globalRunner
}

// Launch starts the engine executable with projectDir as current working directory
func (pm *ProcessManager) Launch(executablePath, projectDir string, onExit func(error)) error {
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

	cmd := exec.Command(executablePath)
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start engine process: %w", err)
	}

	pm.processes[projectDir] = cmd

	go func() {
		err := cmd.Wait()

		pm.mu.Lock()
		delete(pm.processes, projectDir)
		pm.mu.Unlock()

		if onExit != nil {
			onExit(err)
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
	pm.mu.Unlock()

	if !exists || cmd.Process == nil {
		return fmt.Errorf("no running process for project %s", projectDir)
	}

	return cmd.Process.Kill()
}
