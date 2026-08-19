package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"ikemen-studio/pkg/config"
	"ikemen-studio/pkg/engine"
	"ikemen-studio/pkg/project"
)

// App struct
type App struct {
	ctx             context.Context
	downloadsMu     sync.Mutex
	activeDownloads map[string]context.CancelFunc
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		activeDownloads: make(map[string]context.CancelFunc),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// --- Settings APIs ---

// GetSettings loads user preferences
func (a *App) GetSettings() (*config.Settings, error) {
	return config.LoadSettings()
}

// UpdateSettings saves updated preferences
func (a *App) UpdateSettings(s *config.Settings) error {
	return config.SaveSettings(s)
}

// SelectDirectoryDialog opens a native folder selection dialog
func (a *App) SelectDirectoryDialog(title string) (string, error) {
	if title == "" {
		title = "Select Directory"
	}
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
	})
}

// --- Engine APIs ---

// FetchAvailableEngines queries GitHub for Ikemen GO releases
func (a *App) FetchAvailableEngines() ([]engine.ReleaseInfo, error) {
	return engine.FetchReleases()
}

// GetInstalledEngines returns all locally installed engines in configured directory
func (a *App) GetInstalledEngines() ([]engine.InstalledEngine, error) {
	cfg, err := config.LoadSettings()
	if err != nil {
		return nil, err
	}
	return engine.ListInstalledEngines(cfg.EnginesDir)
}

// DownloadEngine downloads and extracts an engine release by version tag
func (a *App) DownloadEngine(tag string) error {
	cfg, err := config.LoadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	releases, err := engine.FetchReleases()
	if err != nil {
		return fmt.Errorf("failed to fetch releases: %w", err)
	}

	var targetRelease *engine.ReleaseInfo
	for i := range releases {
		if strings.EqualFold(releases[i].Tag, tag) {
			targetRelease = &releases[i]
			break
		}
	}

	if targetRelease == nil {
		return fmt.Errorf("release tag %s not found", tag)
	}

	asset, err := engine.FindBestAsset(*targetRelease)
	if err != nil {
		return fmt.Errorf("could not find compatible binary for tag %s: %w", tag, err)
	}

	downloadCtx, cancel := context.WithCancel(context.Background())

	a.downloadsMu.Lock()
	if existingCancel, exists := a.activeDownloads[tag]; exists {
		existingCancel()
	}
	a.activeDownloads[tag] = cancel
	a.downloadsMu.Unlock()

	// Run download in background goroutine with progress events emitted to frontend
	go func() {
		defer func() {
			a.downloadsMu.Lock()
			delete(a.activeDownloads, tag)
			a.downloadsMu.Unlock()
		}()

		err := engine.DownloadAndExtractEngine(downloadCtx, *asset, tag, cfg.EnginesDir, func(progress engine.DownloadProgress) {
			runtime.EventsEmit(a.ctx, "engine-download-progress", progress)
		})

		if err != nil {
			if errors.Is(err, context.Canceled) {
				runtime.EventsEmit(a.ctx, "engine-download-progress", engine.DownloadProgress{
					Version: tag,
					Status:  "cancelled",
				})
			} else {
				runtime.EventsEmit(a.ctx, "engine-download-progress", engine.DownloadProgress{
					Version: tag,
					Status:  "error",
					Error:   err.Error(),
				})
			}
		}
	}()

	return nil
}

// CancelDownload cancels an active engine download
func (a *App) CancelDownload(tag string) error {
	a.downloadsMu.Lock()
	cancel, exists := a.activeDownloads[tag]
	if exists {
		cancel()
		delete(a.activeDownloads, tag)
	}
	a.downloadsMu.Unlock()

	runtime.EventsEmit(a.ctx, "engine-download-progress", engine.DownloadProgress{
		Version: tag,
		Status:  "cancelled",
	})
	return nil
}

// DeleteEngine removes an engine from the local cache
func (a *App) DeleteEngine(version string) error {
	cfg, err := config.LoadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}
	return engine.DeleteEngine(cfg.EnginesDir, version)
}

// --- Project APIs ---

// CreateProject scaffolds a new Ikemen GO project
func (a *App) CreateProject(name, targetDir, engineVersion, author string) (*project.ProjectManifest, error) {
	cfg, err := config.LoadSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}

	var enginePath string
	if engineVersion != "" {
		enginePath = filepath.Join(cfg.EnginesDir, engineVersion)
	}

	channel := "stable"
	if strings.Contains(strings.ToLower(engineVersion), "nightly") || strings.Contains(strings.ToLower(engineVersion), "pre") {
		channel = "nightly"
	}

	manifest, err := project.Scaffold(project.ScaffoldOptions{
		Name:          name,
		TargetDir:     targetDir,
		EngineVersion: engineVersion,
		EngineChannel: channel,
		EnginePath:    enginePath,
		Author:        author,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scaffold project: %w", err)
	}

	// Add to recent projects
	a.addToRecentProjects(targetDir, cfg)

	return manifest, nil
}

// OpenProject loads a project from its directory
func (a *App) OpenProject(projectDir string) (*project.ProjectManifest, error) {
	manifest, err := project.LoadManifest(projectDir)
	if err != nil {
		return nil, err
	}

	cfg, _ := config.LoadSettings()
	if cfg != nil {
		a.addToRecentProjects(projectDir, cfg)
	}

	return manifest, nil
}

// GetRecentProjects returns the list of recently opened projects
func (a *App) GetRecentProjects() ([]string, error) {
	cfg, err := config.LoadSettings()
	if err != nil {
		return nil, err
	}
	return cfg.RecentProjects, nil
}

// SelectProjectDirectoryDialog opens directory picker for project selection
func (a *App) SelectProjectDirectoryDialog() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Ikemen Project Directory",
	})
}

// VerifyAndRepairProject validates core game assets against the configured engine and repairs missing files
func (a *App) VerifyAndRepairProject(projectDir string) (*project.VerificationReport, error) {
	manifest, err := project.LoadManifest(projectDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load project manifest: %w", err)
	}

	cfg, err := config.LoadSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}

	engineDir := filepath.Join(cfg.EnginesDir, manifest.Engine.Version)
	return project.VerifyAndRepairProject(engineDir, projectDir)
}

// OpenProjectLogsFolder opens the project's save/logs directory in the OS file explorer
func (a *App) OpenProjectLogsFolder(projectDir string) error {
	logsDir := filepath.Join(projectDir, "save", "logs")
	_ = os.MkdirAll(logsDir, 0755)
	return a.OpenFolderInExplorer(logsDir)
}

func (a *App) addToRecentProjects(projectDir string, cfg *config.Settings) {
	var updated []string
	updated = append(updated, projectDir)
	for _, p := range cfg.RecentProjects {
		if !strings.EqualFold(p, projectDir) {
			updated = append(updated, p)
		}
		if len(updated) >= 10 {
			break
		}
	}
	cfg.RecentProjects = updated
	_ = config.SaveSettings(cfg)
}

// --- Runner APIs ---

// LaunchProject runs the game binary associated with the project's engine
func (a *App) LaunchProject(projectDir string) error {
	manifest, err := project.LoadManifest(projectDir)
	if err != nil {
		return fmt.Errorf("failed to load project manifest: %w", err)
	}

	cfg, err := config.LoadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	engineDir := filepath.Join(cfg.EnginesDir, manifest.Engine.Version)
	execPath := engine.FindEngineExecutable(engineDir)
	if execPath == "" {
		return fmt.Errorf("could not find Ikemen GO executable for engine version %s (path: %s)", manifest.Engine.Version, engineDir)
	}

	runner := engine.GetProcessManager()
	err = runner.Launch(execPath, projectDir, func(exitErr error, userTerminated bool, outputTail string, logFilePath string) {
		isRealCrash := exitErr != nil && !userTerminated

		if isRealCrash {
			errSummary := engine.ExtractErrorSummary(outputTail)
			runtime.EventsEmit(a.ctx, "game-crashed", map[string]interface{}{
				"projectDir":   projectDir,
				"errorSummary": errSummary,
				"logFilePath":  logFilePath,
				"canRepair":    true,
			})
		}

		runtime.EventsEmit(a.ctx, "game-stopped", map[string]interface{}{
			"projectDir":     projectDir,
			"error":          isRealCrash,
			"userTerminated": userTerminated,
		})
	})

	if err != nil {
		return err
	}

	runtime.EventsEmit(a.ctx, "game-started", projectDir)
	return nil
}

// IsProjectRunning returns whether a game process is active for the given project
func (a *App) IsProjectRunning(projectDir string) bool {
	return engine.GetProcessManager().IsRunning(projectDir)
}

// StopProject terminates a running game process for the given project
func (a *App) StopProject(projectDir string) error {
	return engine.GetProcessManager().Stop(projectDir)
}

// --- OS Integration APIs ---

// OpenFolderInExplorer opens the OS file explorer at the specified directory
func (a *App) OpenFolderInExplorer(folderPath string) error {
	var cmd *exec.Cmd
	switch goruntime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", folderPath)
	case "darwin":
		cmd = exec.Command("open", folderPath)
	default: // linux, bsd
		cmd = exec.Command("xdg-open", folderPath)
	}
	return cmd.Start()
}
