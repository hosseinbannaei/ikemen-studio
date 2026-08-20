# Ikemen Studio - Complete Feature & Test Specification List

This document provides a comprehensive test list of all features within **Ikemen Studio**, describing how each feature is designed to work, expected inputs/outputs, and validation criteria.

---

## Table of Contents
1. [Project Management & Workspaces](#1-project-management--workspaces)
2. [Legacy Game Import & Migration](#2-legacy-game-import--migration)
3. [Engine Management & Version Control](#3-engine-management--version-control)
4. [Game Runner & Launch Modes](#4-game-runner--launch-modes)
5. [Maintenance & Repair Hub](#5-maintenance--repair-hub)
6. [Game Configuration & Config Doctor](#6-game-configuration--config-doctor)
7. [System Settings & UI Themes](#7-system-settings--ui-themes)
8. [Error Handling & Crash Diagnostics](#8-error-handling--crash-diagnostics)

---

## 1. Project Management & Workspaces

### 1.1 Create New Project
- **Purpose**: Scaffolds a clean, standardized Ikemen GO project directory structure tied to an installed engine version.
- **How it works**:
  - User specifies project name, author name, target directory, and engine version.
  - Generates directories: `chars/`, `stages/`, `sound/`, `font/`, `data/`, `save/`, `external/`, `lib/`.
  - Generates `project.json` manifest with project metadata and engine version/channel.
  - Automatically copies baseline files from the selected installed engine.
- **Expected Outcome**: Project opens immediately in the workspace view, appears in the "Recent Projects" list, and launches without missing-asset errors.

### 1.2 Open Existing Project
- **Purpose**: Opens a previously created or imported Ikemen Studio project.
- **How it works**:
  - Native OS directory picker or clicking a card in the Recent Projects list.
  - Validates `project.json` manifest.
  - If `project.json` is missing but valid game folders (`chars/`, `data/`) are present, prompts the user with the Import/Migration wizard.
- **Expected Outcome**: Project workspace loads active engine version, directory shortcuts, and recent backups.

### 1.3 Recent Projects History
- **Purpose**: Quick access to recent projects with quick-launch and removal options.
- **How it works**:
  - Tracks the last 10 opened project paths in `settings.json`.
  - Provides a remove button (removes from list without deleting files on disk).
- **Expected Outcome**: Instant navigation to project workspaces.

### 1.4 Directory Quick Shortcuts
- **Purpose**: Quick OS file manager access to individual asset subfolders.
- **How it works**:
  - Workspace cards for **Characters** (`chars/`), **Stages** (`stages/`), **System Data** (`data/`), **Fonts** (`font/`), **Sound & Music** (`sound/`), and **Roster** (`data/select.def`).
  - Clicking opens the native OS explorer (`explorer.exe` on Windows, `open` on macOS, `xdg-open` on Linux).
- **Expected Outcome**: Native file manager opens to the exact requested subfolder.

---

## 2. Legacy Game Import & Migration

### 2.1 Game Detection & Inspection
- **Purpose**: Analyzes legacy M.U.G.E.N / Ikemen folders to detect assets, engine version compatibility, and missing components.
- **How it works**:
  - Scans for `data/system.def`, `data/select.def`, `chars/`, `stages/`, and executable binaries.
  - Determines whether the project is 0.99 legacy, 0.98, or modern Ikemen GO.
- **Expected Outcome**: Displays detected fighter count, stage count, and compatibility status.

### 2.2 Granular Import Wizard
- **Purpose**: Safely imports legacy projects into a clean Studio workspace.
- **How it works**:
  - Provides checkboxes for: Stock & Custom Characters, Stages, Sound/BGM, Fonts, and `select.def` roster.
  - Replaces legacy 0.99 binaries and Lua VM with modern engine runtime.
  - Generates a fresh `project.json` and creates a backup of the source configuration.
- **Expected Outcome**: Legacy game runs on the modern engine without breaking custom fighters.

---

## 3. Engine Management & Version Control

### 3.1 Online Release Fetcher
- **Purpose**: Queries GitHub API for official Ikemen GO releases.
- **How it works**:
  - Fetches tagged releases and nightlies from `ikemen-engine/Ikemen-GO`.
  - Filters by OS architecture (Windows x64, macOS ARM/Intel, Linux x64).
- **Expected Outcome**: Shows list of available versions with changelogs and release dates.

### 3.2 Background Engine Downloader
- **Purpose**: Downloads and extracts engine binaries without freezing the UI.
- **How it works**:
  - Streams `.zip` or `.tar.gz` archive to local cache.
  - Real-time progress bar (percentage, download speed, total size).
  - Extracts into `AppData/Local/ikemen-studio/engines/<version>/`.
  - Includes cancelation support.
- **Expected Outcome**: Downloaded engine is marked "Installed" and ready for immediate project assignment.

### 3.3 Engine Version Switcher & Migration
- **Purpose**: Upgrades or downgrades a project's engine runtime version.
- **How it works**:
  - Creates a timestamped backup in `save/backups/engine_<timestamp>/`.
  - Replaces `external/`, `lib/`, and core system scripts with the target engine files.
  - Updates `project.json` engine version reference.
- **Expected Outcome**: Project runs on the newly selected engine version.

### 3.4 Automated Engine Rollback System
- **Purpose**: One-click rollback if an engine upgrade causes unexpected issues.
- **How it works**:
  - Lists historical backups stored in the project.
  - Restores previous `external/`, `lib/`, and scripts from the selected backup point.
- **Expected Outcome**: Project is restored to its exact previous state.

---

## 4. Game Runner & Launch Modes

### 4.1 Normal Game Launch
- **Purpose**: Standard arcade/menu launch.
- **Command Line**: `Ikemen_GO.exe` (no extra args)
- **Expected Outcome**: Game opens to title screen / attract mode.

### 4.2 Direct Sparring & Quick Versus Launch
- **Purpose**: Bypasses the title screen and character select to immediately test two characters on a stage.
- **Command Line**: `Ikemen_GO.exe -p1 <char> -p2 <char> -s <stage>`
- **Expected Outcome**: Game instantly enters the fight between P1 and P2 on the designated stage.

### 4.3 Training & Practice Mode Launch
- **Purpose**: Starts a practice session with dummy AI and infinite timer.
- **Command Line**: `Ikemen_GO.exe -p1 <char> -p2 <dummyChar> -p2.ai 0 -time -1 -s <stage>`
- **Expected Outcome**: Game enters match with infinite round timer and stationary P2 dummy.

### 4.4 Developer & Debug Mode Launch
- **Purpose**: Launches the game with debugging HUD and developer tools enabled.
- **Command Line**: `Ikemen_GO.exe -debug -maxpowermode`
- **Expected Outcome**: Game displays hitbox visualization, memory stats, and full power bars.

### 4.5 Custom Launch Options Modal
- **Purpose**: Flexible parameter composer for developers and modders.
- **Supported Flags**:
  - `-windowed` / `-fullscreen`: Force display mode.
  - `-width <w>` & `-height <h>`: Custom resolution.
  - `-nosound`: Mute all audio.
  - `-debug`: Enable hitbox and debug text overlay.
  - `-maxpowermode`: Unlimited power gauge.
  - `-speed <-9..9>`: Game speed modifier.
  - `-ailevel <1..8>`: AI difficulty override.
  - `-s <stage>`: Custom starting arena.
  - `-p1 <char>` / `-p2 <char>`: Custom fighter matchup.
  - Freeform custom arguments input with local history.
- **Expected Outcome**: Appends selected flags to the execution command and launches correctly.

### 4.6 Process Lifecycle & Active Session Management
- **Purpose**: Tracks game execution state, prevents zombie processes, and captures output.
- **How it works**:
  - Captures PID, stdout, and stderr into `save/logs/ikemen-latest.log` via circular buffer.
  - Toolbar reflects live state (`Idle` -> `Starting` -> `Running` -> `Stopping`).
  - Stop button allows graceful/forced process termination.
- **Expected Outcome**: Process state is accurate; game terminates cleanly when stopped.

---

## 5. Maintenance & Repair Hub

### 5.1 Baseline Engine Difference Scanner
- **Purpose**: Compares project structure against the active clean engine template.
- **How it works**:
  - Byte-by-byte file hashing and recursive directory comparisons.
  - Categorizes components into: Stock Fighters, Default Stages, Screenpack & Scripts, Stock Audio/BGM, Fonts, Runtime Lua VM, and `save/config.ini`.
  - Assigns real statuses: `Clean` (green), `Outdated / Modified` (amber), or `Missing` (red).
- **Expected Outcome**: Accurately reflects modified or outdated files.

### 5.2 1-Click Safe Repair
- **Purpose**: Fixes all identified discrepancies in one operation.
- **How it works**:
  - Copies and overwrites outdated baseline assets from the installed engine.
  - Updates `data/common1.cns.zss`, `data/training.zss`, `data/fight.def`, `sound/`, `font/`, `stages/`, and stock fighters.
  - Leaves all custom fighters, custom stages, and roster configuration (`select.def`) untouched.
  - Normalizes `save/config.ini` to modern OpenGL 3.3 standards.
- **Expected Outcome**: All issue badges switch to `Clean` and the project health badge turns green.

### 5.3 Individual Category "Fix" Buttons
- **Purpose**: Target specific components for repair without touching other areas.
- **How it works**:
  - Each issue card provides a dedicated wrench **Fix** button.
  - Opens a dynamic confirmation preview for that specific category.
- **Expected Outcome**: Only the selected component is updated.

### 5.4 Audit & Logs Viewer
- **Purpose**: In-app inspection of session logs and repair verification logs.
- **How it works**:
  - Reads `save/logs/ikemen-latest.log` and `save/logs/asset_sync_report.log`.
  - Includes an "Open Logs Folder" shortcut.
- **Expected Outcome**: Displays full log contents with live refresh.

---

## 6. Game Configuration & Config Doctor

### 6.1 In-App Game Settings Modal
- **Purpose**: GUI editor for `save/config.ini`.
- **How it works**:
  - Form controls for Display Resolution, Window Mode, RenderMode, Master Volume, BGM Volume, SE Volume, Difficulty, Game Speed, and Round Time.
  - Reads and writes directly to `save/config.ini`.
- **Expected Outcome**: Changes take effect upon game launch.

### 6.2 Config Doctor
- **Purpose**: Diagnoses and corrects invalid, legacy, or crash-inducing `config.ini` keys.
- **How it works**:
  - Validates `RenderMode` (e.g. migrates deprecated DirectX / OpenGL 2.0 to `OpenGL 3.3`).
  - Checks for missing required sections.
  - Provides **Auto-Repair Config** and **Reset to Engine Defaults**.
- **Expected Outcome**: Prevents startup crashes caused by corrupt configuration files.

---

## 7. System Settings & UI Themes

### 7.1 Dark / Light Theme Mode
- **Purpose**: Visual theme customization.
- **How it works**:
  - Toggles CSS variables using CSS Color 4 standards (`--bg-app-rgb: R G B`).
  - Ensures clean dark borders without white outline artifacts.
- **Expected Outcome**: Smooth, instant theme switching across all views.

### 7.2 Custom Engines Cache Directory
- **Purpose**: Configurable engine download storage.
- **How it works**:
  - Allows selecting custom storage drives for engine builds.
- **Expected Outcome**: Engine downloads and projects resolve to the custom path.

---

## 8. Error Handling & Crash Diagnostics

### 8.1 Crash Diagnostics Modal
- **Purpose**: Catches engine panics and provides immediate context and recovery paths.
- **How it works**:
  - Monitors process exit codes and stderr streams.
  - Automatically parses panic stack traces (e.g. `training.zss:114: syntax error` or missing audio files).
  - Displays the exact error summary and opens the **Repair Hub** with one click.
- **Expected Outcome**: Users are never left with silent crashes; actionable fixes are provided immediately.
