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

> [!WARNING]
> ### Project Status: Inactive / On Hold
> Development on Ikemen GO Studio has been stopped for the foreseeable future.

### Note from the Creator

> *"Sorry... I guess, I don't know, hope I can do a better job on such things in the future."*

I worked on this project over several days—reaching up to the roster editor and core management features. While it wasn't bad, it didn't become what I originally envisioned (Mostly in technical aspect or... development wise, and somewhat the whole thing itself in general).

It wasn't that I came to dislike the project itself, but that I had no idea how it exactly worked, I knew what systems it had and how most of things worked, but only at a basic level, I decided how most of the systems work and designed them to some extend but... the AI sometimes went ahead and did things that i wasn't sure about how they worked and didn't understand their design or system or whatever as much. I don't really know what I was chasing after, but this wasn't exactly it and it was growing more and more into something i thought i knew what it is but wasn't sure about it. The vibe coding process or how i actually approached it for this project just didn't feel inline with whatever feeling or output I was expecting from myself.

So for now, I've decided to put it to rest, keep searching, and try other things. maybe i should look at things from a different angle, think about other stuff, who knows. The repository is made public for anyone interested in referencing the code or ideas, hope it has some use, that's it for now... I guess.

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
