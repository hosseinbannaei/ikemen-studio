Ikemen GO Studio - Project Overview & Vision

1. The Core Problem

Ikemen GO is one of the most powerful, feature-rich 2D fighting game engines available. Built in modern Go with open-source licensing (MIT), native cross-platform support, and built-in rollback netcode, it provides the exact mechanical foundation fighting game creators need.

However, creator adoption is heavily throttled by legacy MUGEN configuration debt. Editing nested .def, .cns, and .zss plain text files manually, wrestling with relative file paths, handling broken font references, and managing rosters via raw text lines creates immense friction. Existing tools (like V-Select and Fighter Factory) are either abandoned, Windows-only, or clunky for modern developer workflows.

2. The Vision

Ikemen GO Studio is a modern, fast, lightweight desktop manager and development environment designed to make creating, managing, and packaging 2D fighting games effortless across Linux, Windows, and macOS.

Instead of replacing Ikemen GO, the Studio wraps around it as a high-velocity command center. It bridges the gap between raw configuration files and a modern visual workflow.

3. Core Architectural Principles

Separation of Engine and Project Data:
To guarantee true cross-platform capability and shareable project files, engine binaries are stored in a central global cache (~/.local/share/ikemen-studio/engines/). Project folders contain only user assets (chars/, stages/, data/, manifest).

Incremental Evolution:
The Studio starts as a clean, reliable Version & Roster Manager before expanding into complex visual editors (hitboxes, sprite animation timelines, state graphs).

Human Architecture, AI Execution Velocity:
The project leverages AI agents for rapid feature iteration and boilerplate generation. Architecture, data structures, and state mutations are human-designed and strictly owned to prevent code rot.

Open Source Transparency:
The project will be clearly documented on GitHub with explicit disclosure of AI-assisted generation, welcoming community contributions and custom plugins.

4. Long-Term Evolution Strategy

Phase 1: Project & Engine Manager (Current Goal)
Engine downloader, cross-platform launcher, project scaffolding, and visual select.def roster grid management.

Phase 2: Asset & Linter Inspector
Visual .sff sprite previewer, .snd player, missing file/font validator, and automated movelist.dat generator.

Phase 3: Visual State & Frame Studio
Interactive hitbox/ Hurtbox editor, .air animation timeline keyframing, and .zss script syntax highlighting.