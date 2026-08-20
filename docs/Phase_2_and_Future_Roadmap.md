# Ikemen Studio - Project State Analysis & Future Roadmap

**Document Version:** 1.0.0  
**Date:** August 20, 2026  
**Status:** Active Strategic Plan  

---

## 1. Executive Summary & Current State Analysis

Ikemen Studio has reached full maturity for its **Phase 1 (Core Manager & Workspace Hub)** milestones. The architecture has evolved into a robust, responsive, cross-platform command center that completely eliminates the historical friction of managing Ikemen GO and legacy MUGEN fighting games.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           IKEMEN STUDIO HUB                             │
├───────────────────┬─────────────────────────┬───────────────────────────┤
│  ENGINE MANAGER   │    PROJECT WORKSPACE    │        ASSET VAULT        │
│  • Stable/Nightly │    • Visual Scaffolder  │  • Multi-Vault Registries │
│  • Auto-Download  │    • Diagnostic Repair  │  • Archive Ingestion      │
│  • Version Backup │    • Legacy Matcher     │  • SFF v1/v2 Portraits    │
│  • Sandboxed Dir  │    • Crash Log Analyzer │  • Zero-Copy Symlinking   │
├───────────────────┴─────────────────────────┴───────────────────────────┤
│                         ROSTER MATRIX STUDIO                            │
│  • Two-Way select.def Sync   • Multi-Slot Selection   • Bulk Placer     │
│  • Arcade Grid Dimensioning  • Duplicate Indicators   • Auto-Populate   │
├─────────────────────────────────────────────────────────────────────────┤
│                         GLOBAL THEMING SYSTEM                           │
│  • NetherRealm MKX (Default) • Arcade Obsidian OLED   • Cyber Strike    │
│  • Capcom Classic Tournament • Clean Studio Light     • Brutalist Radii │
└─────────────────────────────────────────────────────────────────────────┘
```

### 1.1 Key Accomplishments Delivered

1. **Engine Manager & Sandboxed Runtime (`pkg/engine/`)**:
   - Automated querying, downloading, caching, and execution of official Ikemen GO releases (Stable and Nightly).
   - Process lifecycle management, crash log trapping, and custom CLI parameter injection.
   - Version rollback and engine snapshot backups.

2. **Project Scaffolding & Diagnostic Repair Hub (`pkg/project/`)**:
   - Clean scaffolding with zero loose-file clutter.
   - Legacy MUGEN game importer with 3 migration strategies: Complete Scaffolding Rebuild, Smart Diff Merge, and Legacy Exact Match.
   - Self-Repair Hub: automated detection and resolution of missing system files, broken font references, and invalid `config.ini` keys.

3. **Decentralized Multi-Vault & Ingestion Pipeline (`pkg/vault/`)**:
   - Universal ingestion of `.zip`, `.7z`, `.rar` archives and standalone game folders.
   - Subfolder normalizer, cross-contamination cleaner, deduplication engine, and recovery scanner.
   - Binary SFF extractor reading SFF v1 and SFF v2 headers to extract standard character portraits (`9000,0` and `9000,1`) into cached base64 thumbnails.
   - High-velocity zero-duplication linking (Symlink / Hardlink / Copy) directly into active game projects.
   - Bulk selection toolbar with 1-click batch linking and deletion.

4. **Visual Roster & Select Screen Matrix (`pkg/parser/`, `RosterEditorView.svelte`)**:
   - Live two-way synchronization with `data/select.def`.
   - Grid matrix with customizable rows, columns, densities (Arcade, Medium, Large), and motif matching from `system.def`.
   - Edge-to-edge arcade cards, boss tier badges, and duplicate character badges (`2x`, `3x`).
   - Quick workflow tools: Auto-Populate All, Shuffle Matrix, Sort Fighters (Name, Author, Boss Order), Fill Empty with `?` (Random Select), and Trim Trailing Empty Slots.
   - Dock library batch selection with sticky placement bar.

5. **Fighting Game Inspired Theming Engine (`frontend/`)**:
   - 5 themes tailored for fighting games: NetherRealm MKX (Default), Arcade Obsidian (OLED), Cyber Strike (Tekken Night), Capcom Classic (Street Fighter), and Clean Studio.
   - Corner Geometry System (Sharp MKX Brutalist `0–2px`, Subtle Minimal `4–6px`, Soft Rounded `8–12px`).
   - Interactive Settings Gallery with live color palette previews.

---

## 2. Strategic Objectives for Next Phases

With the core management infrastructure solid and battle-tested, the studio is positioned to expand into **Phase 2 (Deep Asset Analysis, Audio & Character Tooling)** and **Phase 3 (Visual State, Hitbox & Frame Studio)**.

```
       Phase 1 (Completed)                 Phase 2 (Next Target)               Phase 3 & 4 (Future)
┌───────────────────────────────┐   ┌─────────────────────────────────┐   ┌────────────────────────────────┐
│ • Engine Lifecycle & Launcher │   │ • Visual SFF Sprite Explorer    │   │ • Interactive Hitbox Editor    │
│ • Project Scaffolding & Hub   │──▶│ • Built-in .snd & Audio Studio  │──▶│ • .air Animation Keyframer     │
│ • Asset Vault & Ingestion     │   │ • Fighter & Stage Linter        │   │ • 1-Click Standalone Exporter  │
│ • select.def Roster Matrix    │   │ • Auto Movelist.dat Generator   │   │ • Automated Credits Generator  │
└───────────────────────────────┘   └─────────────────────────────────┘   └────────────────────────────────┘
```

---

## 3. Phase 2 Detailed Roadmap: Deep Asset & Character Tooling

### 🎯 Feature 2.1: SFF Visual Sprite & Palette Explorer
**Objective**: Enable creators to browse all character sprites, animation frames, and color palettes without needing external third-party software.

- **Capabilities**:
  - Full support for both **SFF v1** and **SFF v2** formats.
  - Group and Sprite index browser (e.g. Group `0` Stand, Group `20` Walk, Group `5000` GetHit).
  - Palette (`.act` / embedded palette) switcher: preview characters in alternative costume colors in real-time.
  - Export selected sprites or entire sprite sheets as PNG.
  - Zoom, pan, transparency toggle, and onion-skinning preview.

### 🎯 Feature 2.2: Built-in Audio & Voice Studio (`.snd` & BGM)
**Objective**: Provide an integrated audio inspection tool for character sound effects, voice lines, and stage music.

- **Capabilities**:
  - Binary `.snd` file parser to list and play all sound clips (Group / Index).
  - Visual waveform player with volume slider, looping toggle, and frequency display.
  - Stage BGM player for `.mp3`, `.ogg`, and `.wav` tracks defined in stage `.def` files.
  - Sound effect exporter and replacement tool.

### 🎯 Feature 2.3: Character Health & Integrity Linter
**Objective**: Automated diagnostic scanner that identifies bugs, missing dependencies, and performance bottlenecks in character files before launching the game.

- **Capabilities**:
  - **Missing Assets Checker**: Scans `.air`, `.cns`, and `.zss` for references to missing sprite groups or sound indices.
  - **Syntax & State Validator**: Flags deprecated state controllers, undefined variables, and broken state references.
  - **Palette Consistency Check**: Detects broken color tables that cause sprite discoloration.
  - **One-Click Fixer**: Offers automated patches for common syntax issues.

### 🎯 Feature 2.4: Automated Movelist & Command List Generator
**Objective**: Parse character `.cmd` files to automatically generate clean command lists.

- **Capabilities**:
  - Extracts special moves, super arts, command normals, and combo chains from `.cmd` files.
  - Renders arcade-style input icon strips (e.g. `↓ ↘ → + Punch`).
  - Exports to in-game `movelist.dat` or printable HTML/PDF cheat sheets.

---

## 4. Phase 3 Detailed Roadmap: Visual State & Frame Studio

### 🥋 Feature 3.1: Interactive Hitbox & Hurtbox Previewer (`.air` / `.cns`)
- Visual overlay of collision boxes directly on character sprite frames:
  - **Clsn1 (Red)**: Attack / Hitboxes.
  - **Clsn2 (Blue)**: Vulnerability / Hurtboxes.
- Frame-by-frame scrubbing with attack priority indicators.
- Box resizing, repositioning, and keyframe propagation.

### 🎬 Feature 3.2: Animation Timeline & Keyframe Editor
- Interactive timeline bar with tick rate control (standard 60 FPS).
- Edit frame duration, interpolation, flip flags (`H`, `V`), and alpha blending modes (`Add`, `Sub`, `AS50D50`).
- Visual loop point markers and transition previews.

### 💻 Feature 3.3: Modern Script & State Graph Editor (.zss / .cns)
- Code editor with syntax highlighting for ZSS (Z-Script) and CNS.
- Visual State Flow Graph: Node-based visualizer showing transitions between Idle -> Attack -> Hit -> Recovery states.

---

## 5. Phase 4 Detailed Roadmap: Game Packaging & Release

### 📦 Feature 5.1: 1-Click Standalone Game Exporter
- Packages a complete, standalone game distribution for Windows, Linux, or macOS.
- Bundles the active Ikemen GO binary and libraries.
- Strips unnecessary development files, temp logs, and unlinked assets to reduce file size.
- Includes custom game icon and executable naming.

### 📜 Feature 5.2: Automated Game Credits Generator
- Pulls author, website, source URL, and license metadata from all linked Vault assets.
- Generates a formatted `CREDITS.md` and in-game text file giving full attribution to all creators.

---

## 6. Immediate Next Steps (Priority Execution Plan)

### Sprint A (Immediate Focus: Audio & Sprite Inspection)
1. **Build `pkg/parser/snd`**: Go parser to read `.snd` audio packs and stream WAV/PCM audio buffers.
2. **Build Audio Inspector UI**: In-app sound board to listen to character voice clips and SFX.
3. **Build `pkg/parser/sff` Full Explorer**: Expand beyond portraits to decode and stream arbitrary sprite groups.

### Sprint B (Character Integrity & Movelist)
1. **Build Character Linter**: Automated scan of character folders for missing sprite/sound indices.
2. **Build Movelist Generator**: `.cmd` parser converting raw inputs to arcade joystick/button icon sequences.

### Sprint C (Packaging & Standalone Export)
1. **Build Game Exporter (`pkg/export/`)**: 1-click packager that bundles engine binaries and creates ready-to-play `.zip` releases.
2. **Build Credits Compiler**: Auto-generates `CREDITS.md` from project and vault manifests.
