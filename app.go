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

// RemoveRecentProject removes a project from the recent history
func (a *App) RemoveRecentProject(projectDir string) error {
	return config.RemoveRecentProject(projectDir)
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

	a.addToRecentProjects(targetDir, cfg)
	return manifest, nil
}

// DetectExistingGame analyzes an existing game folder
func (a *App) DetectExistingGame(dir string) (*project.ExistingGameInspection, error) {
	return project.DetectExistingGame(dir)
}

// ImportExistingGame copies an existing MUGEN/Ikemen folder into a new managed studio project
func (a *App) ImportExistingGame(srcDir, targetDir, name, engineVersion, author string) (*project.ProjectManifest, error) {
	return a.ImportExistingGameWithOptions(project.ImportOptions{
		SourceDir:     srcDir,
		TargetDir:     targetDir,
		ProjectName:   name,
		EngineVersion: engineVersion,
		Author:        author,
		Mode:          "rebuild",
		IncludeChars:  true,
		IncludeStages: true,
		IncludeSound:  true,
		IncludeFonts:  true,
		IncludeRoster: true,
	})
}

// ImportExistingGameWithOptions imports legacy game with custom mode and checklist
func (a *App) ImportExistingGameWithOptions(opts project.ImportOptions) (*project.ProjectManifest, error) {
	cfg, err := config.LoadSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}

	if opts.EngineVersion != "" {
		opts.EnginePath = filepath.Join(cfg.EnginesDir, opts.EngineVersion)
	}

	channel := "stable"
	if strings.Contains(strings.ToLower(opts.EngineVersion), "nightly") || strings.Contains(strings.ToLower(opts.EngineVersion), "pre") {
		channel = "nightly"
	}
	opts.EngineChannel = channel

	manifest, err := project.ImportExistingGame(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to import game: %w", err)
	}

	a.addToRecentProjects(opts.TargetDir, cfg)
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

// SwitchProjectEngine migrates project to a different engine version with backups
func (a *App) SwitchProjectEngine(projectDir, newVersion string) error {
	cfg, err := config.LoadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	newEngineDir := filepath.Join(cfg.EnginesDir, newVersion)
	if _, err := os.Stat(newEngineDir); os.IsNotExist(err) {
		return fmt.Errorf("engine version %s is not downloaded", newVersion)
	}

	return project.SwitchEngine(projectDir, newEngineDir, newVersion)
}

// GetEngineBackups returns backups created during engine switches
func (a *App) GetEngineBackups(projectDir string) ([]project.EngineBackupInfo, error) {
	return project.GetEngineBackups(projectDir)
}

// RollbackProjectEngine restores a previous runtime backup
func (a *App) RollbackProjectEngine(projectDir, backupID string) error {
	return project.RollbackEngine(projectDir, backupID)
}

// VerifyAndRepairProject validates core game assets against the configured engine and repairs missing files
func (a *App) VerifyAndRepairProject(projectDir string) (*project.VerificationReport, error) {
	return a.VerifyAndRepairProjectWithMode(projectDir, false)
}

// VerifyAndRepairProjectWithMode validates core assets and optionally updates core system data scripts (common1.cns.zss, etc.)
func (a *App) VerifyAndRepairProjectWithMode(projectDir string, updateCoreSystem bool) (*project.VerificationReport, error) {
	manifest, err := project.LoadManifest(projectDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load project manifest: %w", err)
	}

	cfg, err := config.LoadSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}

	engineDir := filepath.Join(cfg.EnginesDir, manifest.Engine.Version)
	return project.VerifyAndRepairProjectWithMode(engineDir, projectDir, updateCoreSystem)
}

// OpenProjectLogsFolder opens the project's save/logs directory in the OS file explorer
func (a *App) OpenProjectLogsFolder(projectDir string) error {
	logsDir := filepath.Join(projectDir, "save", "logs")
	_ = os.MkdirAll(logsDir, 0755)
	return a.OpenFolderInExplorer(logsDir)
}

// GetGameConfig loads save/config.ini from project
func (a *App) GetGameConfig(projectDir string) (map[string]string, error) {
	return config.LoadGameConfig(projectDir)
}

// SaveGameConfig updates save/config.ini in project
func (a *App) SaveGameConfig(projectDir string, updates map[string]string) error {
	return config.SaveGameConfig(projectDir, updates)
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
	return a.LaunchProjectWithOptions(projectDir, []string{})
}

// LaunchProjectWithOptions runs the game with custom command-line arguments
func (a *App) LaunchProjectWithOptions(projectDir string, args []string) error {
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
	err = runner.Launch(execPath, projectDir, args, func(exitErr error, userTerminated bool, outputTail string, logFilePath string) {
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
