# Ikemen GO Studio

> Modern, lightweight, cross-platform desktop manager and development environment for **Ikemen GO** fighting games.

Ikemen GO Studio streamlines the creation, management, and launching of Ikemen GO fighting game projects across Windows, Linux, and macOS.

---

## Features (Phase 1)

- **Engine Version Manager**: Discover, download, cache, and delete official Ikemen GO releases (Stable and Nightly) directly from GitHub with real-time download progress.
- **Project Scaffolder**: Scaffold clean Ikemen GO projects (`chars/`, `stages/`, `data/`, `font/`, `sound/`), automatically migrate base engine assets, and track project state via `ikemen-project.json`.
- **Game Launcher**: Run projects using cached engine binaries with proper working directory configuration and process lifecycle tracking.
- **Configurable Settings**: Customize global engine cache directories with persistent JSON preferences across app restarts.
- **Modern Desktop UI**: Built with Go, Wails v2, Svelte, TypeScript, and Tailwind CSS.

---

## AI Development Transparency Notice

> **Notice on AI Usage**: This project's architecture, data contracts, and UX decisions are human-led and designed. AI code agents are utilized during development for rapid iteration, boilerplate reduction, and prototyping speed. Every contribution is strictly reviewed and tested.

---

## Prerequisites

- **Go**: 1.20+ (tested on Go 1.25)
- **Node.js**: 18+ (tested on Node 22 & npm 10)
- **Wails CLI**: v2.15+ (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- **WebView2** (Windows) / **WebKitGTK** (Linux)

---

## Getting Started

### Development Mode

Run the application with live hot-reload for both the frontend and Go backend:

```bash
wails dev
```

### Production Build

Package the application into an optimized standalone binary:

```bash
wails build
```

The output executable will be generated under `build/bin/`.

---

## License

This project is licensed under the [MIT License](LICENSE).
