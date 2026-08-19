Phase 1 Technical Guide & Roadmap

This document outlines the architecture, directory structure, data schemas, and implementation steps required to complete Phase 1 of Ikemen GO Studio.

1. Phase 1 Scope & Objectives

The goal of Phase 1 is to ship a fully functional desktop app capable of:

Downloading and caching official Ikemen GO releases (stable/nightly) from GitHub.

Scaffolding new Ikemen GO projects with a clean directory layout.

Managing local projects via a root manifest (ikemen-project.json).

Scanning character folders and providing a visual grid editor for data/select.def.

Launching the game using cached OS-specific binaries with correct working directories.

2. Recommended Directory & Project Architecture

The project uses Wails v2 (Go backend + Web frontend using Svelte/React & TypeScript).

```
ikemen-studio/
├── app.go                  # Main Wails application binding methods
├── main.go                 # Application entry point & setup
├── go.mod / go.sum
├── pkg/                    # Core Go backend packages
│   ├── engine/             # Engine download, extraction, & execution
│   │   ├── downloader.go
│   │   └── runner.go
│   ├── project/            # Project manifest, scaffolding, & filesystem operations
│   │   ├── manifest.go
│   │   └── scaffolder.go
│   └── parser/             # MUGEN/Ikemen file parsers (.def, select.def)
│       └── select_def.go
├── frontend/               # UI codebase (Svelte or React + Tailwind CSS)
│   ├── src/
│   │   ├── lib/            # Reusable UI components (Grid, Modals, Topbar)
│   │   ├── stores/         # UI state management
│   │   └── App.svelte
│   └── wailsjs/            # Auto-generated TypeScript bindings for Go
└── docs/                   # Developer documentation
```

3. Project Manifest Schema (ikemen-project.json)

Every project managed by the Studio contains a root JSON manifest:
```
{
  "name": "My Custom Fighter",
  "version": "0.1.0",
  "engine": {
    "version": "v0.99.0",
    "channel": "stable"
  },
  "created_at": "2026-08-19T14:00:00Z",
  "author": "Developer"
}
```

4. Phase 1 Implementation Roadmap

Step 1: Engine Download & Caching Manager

Build pkg/engine/downloader.go:

Query GitHub API for releases (https://api.github.com/repos/ikemen-engine/Ikemen-GO/releases).

Download OS-specific tarball/zip to ~/.local/share/ikemen-studio/engines/<version>/.

Decompress files and set execution permissions on Linux/macOS (chmod +x).

Step 2: Project Scaffolder & Launcher

Build pkg/project/scaffolder.go:

Create project directory structure (chars/, stages/, data/, font/, sound/).

Copy base system scripts from cached engine into project data/.

Write initial ikemen-project.json.

Build pkg/engine/runner.go:

Launch cached engine process setting cmd.Dir = projectPath.

Step 3: Character Scanner & select.def Visual Grid

Build pkg/parser/select_def.go:

Read and parse existing data/select.def files.

Scan chars/ subdirectories to detect valid character .def entries.

Build Frontend Drag-and-Drop Grid:

Display character cards and empty slots.

Allow drag-and-drop reordering and slot assignment.

Serialize user layout back to data/select.def.

5. README.md Guidelines for the Repository

When setting up the project's main README.md, include the following structure:

Project Title & Tagline: Highlighting the Studio's purpose.

Features Showcase: Key phase capabilities (Engine management, visual roster grid, one-click launcher).

AI Development Transparency Statement:

"Notice on AI Usage: This project's architecture, data contracts, and UX decisions are human-led and designed. AI code agents are utilized during development for rapid iteration, boilerplate reduction, and prototyping speed. Every contribution is strictly reviewed and tested."

Getting Started & Building:

Prerequisites (go 1.20+, wails v2, npm).

Commands for wails dev and wails build.

License: MIT License.