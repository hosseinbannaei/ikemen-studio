# Ikemen GO Studio - Phase 1: Engine Manager, Project Scaffolder & Launcher

## Overview
Phase 1 delivers the foundational desktop workflow for Ikemen GO creators:
1. **Engine Version Manager**: Discover, download, extract, cache, and delete official Ikemen GO releases (stable & nightly) from GitHub.
2. **Project Scaffolding & Management**: Create new clean Ikemen GO projects, initialize standard directory structure (`chars/`, `stages/`, `data/`, `font/`, `sound/`), copy base system assets from cached engine, and track projects via `ikemen-project.json`.
3. **Game Launcher**: Execute cached Ikemen GO binaries with the project root as the working directory.
4. **Configurable Settings**: Custom storage directory for downloaded engine binaries, stored persistently in user config.
5. **Modern Desktop UI (Svelte + TypeScript + Tailwind CSS)**: Sleek, lightweight interface for managing engines, creating/opening projects, changing settings, and launching games.

*(Note: Roster Grid Visual Editor and `select.def` parser are intentionally deferred to Phase 2 as requested).*

---

## User Review Required

> [!IMPORTANT]
> **Go & Wails Versioning**: We will use `go1.25.14` (installed via `gvm`) and install `wails v2` CLI (`github.com/wailsapp/wails/v2/cmd/wails@latest`).
>
> **Engine Storage Defaults**:
> - Linux: `~/.local/share/ikemen-studio/engines`
> - Windows: `%LOCALAPPDATA%\ikemen-studio\engines`
> - macOS: `~/Library/Application Support/ikemen-studio/engines`
>
> **Config Storage Defaults**:
> - Linux: `~/.config/ikemen-studio/settings.json`
> - Windows: `%APPDATA%\ikemen-studio\settings.json`
> - macOS: `~/Library/Application Support/ikemen-studio/settings.json`

---

## Proposed Changes

### 1. Build & Project Configuration

#### [NEW] [wails.json](file:///home/wrench/Work/Projects/Personal/Crossplatform/ikemen-studio/wails.json)
- Wails application configuration specifying app name, frontend build commands, Go bindings output, and window defaults.

#### [NEW] [go.mod](file:///home/wrench/Work/Projects/Personal/Crossplatform/ikemen-studio/go.mod)
- Go module definition `ikemen-studio` with dependencies (`github.com/wailsapp/wails/v2`, etc.).

#### [NEW] [main.go](file:///home/wrench/Work/Projects/Personal/Crossplatform/ikemen-studio/main.go)
- Application entry point initializing Wails runtime with embedded frontend assets and backend bindings.

#### [NEW] [app.go](file:///home/wrench/Work/Projects/Personal/Crossplatform/ikemen-studio/app.go)
- Wails backend service struct exposing all Go APIs to the Svelte frontend via auto-generated TypeScript bindings:
  - Settings APIs (`GetSettings`, `UpdateSettings`, `SelectDirectoryDialog`)
  - Engine APIs (`FetchAvailableEngines`, `GetInstalledEngines`, `DownloadEngine`, `DeleteEngine`)
  - Project APIs (`CreateProject`, `OpenProject`, `GetRecentProjects`, `SelectProjectDirectoryDialog`)
  - Runner APIs (`LaunchProject`, `GetRunningStatus`, `KillGame`)

#### [NEW] [README.md](file:///home/wrench/Work/Projects/Personal/Crossplatform/ikemen-studio/README.md)
- Project introduction, architecture overview, AI transparency statement, build instructions, and MIT license.

---

### 2. Go Backend Packages (`pkg/`)

#### [NEW] [pkg/config/settings.go](file:///home/wrench/Work/Projects/Personal/Crossplatform/ikemen-studio/pkg/config/settings.go)
- `Settings` struct: `EnginesDir`, `Theme`, `RecentProjects`, `DefaultChannel`.
- Default path resolution (XDG on Linux / AppData on Windows / Library on macOS).
- Persistent JSON read/write operations.

#### [NEW] [pkg/engine/downloader.go](file:///home/wrench/Work/Projects/Personal/Crossplatform/ikemen-studio/pkg/engine/downloader.go)
- GitHub API client fetching `ikemen-engine/Ikemen-GO` releases.
- Asset matching logic (detecting OS: Linux `x86_64`/`arm64`, Windows `x64`, macOS `darwin-amd64`/`arm64`).
- Streamed download with progress events emitted to Wails frontend.
- Decompressor supporting `.tar.gz` and `.zip` archives.
- Setting executable permissions (`chmod +x`) on engine binary (`Ikemen_GO` / `Ikemen_GO.exe`).

#### [NEW] [pkg/engine/manager.go](file:///home/wrench/Work/Projects/Personal/Crossplatform/ikemen-studio/pkg/engine/manager.go)
- Scanning installed engines in configured `EnginesDir`.
- Detecting binary path, version tags, and engine integrity.
- Deleting engine versions safely.

#### [NEW] [pkg/engine/runner.go](file:///home/wrench/Work/Projects/Personal/Crossplatform/ikemen-studio/pkg/engine/runner.go)
- Executing engine process with `cmd.Dir = projectPath`.
- Process tracking (PID, start/exit detection, graceful termination).

#### [NEW] [pkg/project/manifest.go](file:///home/wrench/Work/Projects/Personal/Crossplatform/ikemen-studio/pkg/project/manifest.go)
- `ProjectManifest` schema (`ikemen-project.json`):
  ```json
  {
    "name": "Project Name",
    "version": "0.1.0",
    "engine": {
      "version": "v0.99.0",
      "channel": "stable"
    },
    "created_at": "2026-08-19T14:00:00Z",
    "author": "Developer"
  }
  ```
- Validation, serialization, and deserialization.

#### [NEW] [pkg/project/scaffolder.go](file:///home/wrench/Work/Projects/Personal/Crossplatform/ikemen-studio/pkg/project/scaffolder.go)
- New project scaffolding:
  - Creates directories: `chars/`, `stages/`, `data/`, `font/`, `sound/`.
  - Copies base engine system data (e.g. `data/system.def`, default scripts) from the specified cached engine version.
  - Generates initial `ikemen-project.json` and default empty `data/select.def`.

---

### 3. Frontend Architecture (`frontend/`)

#### [NEW] [frontend/package.json](file:///home/wrench/Work/Projects/Personal/Crossplatform/ikemen-studio/frontend/package.json) & [vite.config.ts](file:///home/wrench/Work/Projects/Personal/Crossplatform/ikemen-studio/frontend/vite.config.ts)
- Svelte + TypeScript + Vite + Tailwind CSS + Lucide Icons.

#### [NEW] [frontend/src/App.svelte](file:///home/wrench/Work/Projects/Personal/Crossplatform/ikemen-studio/frontend/src/App.svelte)
- Main application shell:
  - **Top Navigation Bar**: Active project title, selected engine badge, "Play / Launch" action button with live running state indicator, Settings button, Engine Manager button.
  - **Main View Area**: Swappable views (Project Dashboard, Engine Manager, Settings Modal, New Project Wizard).

#### [NEW] [frontend/src/lib/components/](file:///home/wrench/Work/Projects/Personal/Crossplatform/ikemen-studio/frontend/src/lib/components/)
- `Header.svelte`: Quick launch bar, active project & engine version badge, status indicators.
- `EngineManager.svelte`: Tab/modal listing Available GitHub Releases (Stable/Nightly tags, dates, changelog links) vs Installed Versions, Download buttons with live progress bar, and Delete buttons.
- `ProjectDashboard.svelte`: Project details card (name, author, engine version, path), quick links to open project folder in OS file explorer, launch game button.
- `NewProjectModal.svelte`: Step-by-step creation modal (Project name, destination directory picker, engine version dropdown, create button).
- `SettingsModal.svelte`: Custom Engine Directory path picker, theme options, default engine channel toggle.
- `Toast.svelte` / `Notification.svelte`: Non-intrusive alerts for download completions, launch errors, and file actions.

#### [NEW] [frontend/src/lib/stores/](file:///home/wrench/Work/Projects/Personal/Crossplatform/ikemen-studio/frontend/src/lib/stores/)
- `projectStore.ts`: Current active project state, recent projects list.
- `engineStore.ts`: Available and installed engine versions, active downloads & progress map.
- `settingsStore.ts`: App settings & directories.

---

## Verification Plan

### Automated Unit Tests
- `go test ./pkg/project/...`: Test manifest creation, parsing, validation, and directory scaffolding.
- `go test ./pkg/config/...`: Test settings loading, fallback defaults, and updates.
- `go test ./pkg/engine/...`: Test asset filename matcher (Linux/Win/macOS patterns) and engine directory scanner.

### Manual Verification
1. **Engine Manager**:
   - Fetch release list from GitHub API and verify stable/nightly tags display properly.
   - Trigger download of an Ikemen GO release (e.g. `v0.99.0` or latest).
   - Observe live progress bar, extraction into the configured engines directory, and permission setup.
   - Verify installed version appears in the list with binary path and can be deleted if requested.
2. **Settings**:
   - Change engine directory to a custom folder, verify directory creation and persistence across app restarts.
3. **Project Scaffolding**:
   - Create a new project via the UI wizard selecting the downloaded engine.
   - Inspect created folders (`chars/`, `stages/`, `data/`, etc.) and `ikemen-project.json`.
4. **Game Launcher**:
   - Click "Play" on the created project.
   - Verify Ikemen GO window launches with the project directory as working directory.
   - Verify Studio UI shows "Running" status and handles game exit cleanly.
