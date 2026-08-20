# Ikemen Studio - Asset Vault Architecture & Specification

## 1. Executive Summary & Vision

The **Asset Vault** is Ikemen Studio's decentralized local package repository and asset management system. 

Instead of treating character downloads and game assets as loose, duplicated files scattered across individual project folders, the Vault acts as a central indexed library. Projects connect to vaults through non-destructive file linking (Symlinks/Hardlinks/Copies), enabling zero-duplication asset sharing, automatic metadata extraction, smart tagging, and seamless project roster construction.

---

## 2. Core Concepts & Taxonomy

To prevent data ambiguity, the system strictly separates **Categories**, **Engine Compatibilities**, **Source Packages**, and **User Tags**.

```
Asset
├── Category (Fixed structural taxonomy: Fighter, Stage, Motif, Sound/BGM)
├── Metadata (Auto-extracted: Name, Author, Version, mugenversion, Source URL)
├── Source Package (Origin archive/project trace: e.g. "Hyper_DBZ_v5.zip")
└── Tags (Freeform multi-value labels: e.g. "DragonBall", "6-Button", "Custom AI")
```

### 2.1 Fixed Categories
1. **`fighters`** (Characters): Contains `.def`, `.sff`, `.air`, `.cns`/`.zss`, `.cmd`, `.snd`.
2. **`stages`** (Arenas): Contains `.def`, `.sff`, optional `.mp3`/`.ogg`.
3. **`motifs`** (Screenpacks / UI Themes): Contains `system.def`, fonts, motif sprites, and animations.
4. **`sounds`** (BGM & Audio Packs): Standalone background music and sound effects libraries.

### 2.2 Tags vs Categories
* **Category (Fixed)**: What the engine *mechanically* uses the file as (e.g. `Fighter`).
* **Compatibility (Auto-detected)**: `MUGEN 1.0`, `MUGEN 1.1`, `Ikemen GO`, `WinMugen`.
* **User & Smart Tags (Dynamic)**: `Capcom`, `Marvel`, `Boss`, `Original IP`, `HD`, `WIP`, `Chibi`, `Commercial Safe`.

---

## 3. Multi-Vault System Architecture

Users can register multiple independent vaults on any drive or network mount.

```
~/.local/share/ikemen-studio/ (or Custom Drive)
└── vaults/
    ├── default/                       # Default system vault
    │   ├── vault.json                 # Vault metadata & index
    │   ├── chars/
    │   │   └── ryu/
    │   ├── stages/
    │   └── sound/
    ├── custom_vault_anime/            # Additional custom vault
    │   └── vault.json
    └── custom_vault_commercial_ip/    # Project-specific isolated vault
        └── vault.json
```

### 3.1 Vault Registration & Scalability
- **Registration**: All registered vault paths are stored in the user's global `settings.json`.
- **Scaling to 100+ Vaults**:
  - In Go, reading 100 small `vault.json` files takes < 10 ms.
  - The frontend caches indexed asset headers in memory and supports global search across all vaults as well as single-vault isolation.
- **Portability**: A vault folder is 100% self-contained. Moving a vault to a USB drive or sharing it via Git/Nextcloud preserves all metadata, tags, and attribution.

### 3.2 Vault Manifest Schema (`vault.json`)
```json
{
  "version": "1.0",
  "id": "vault-98a7b6c5",
  "name": "Capcom Collection",
  "description": "Curated Capcom fighting game characters and stages",
  "created_at": "2026-08-20T17:00:00Z",
  "updated_at": "2026-08-20T17:15:00Z",
  "assets": {
    "chars/ryu_warusaki": {
      "category": "fighters",
      "display_name": "Ryu",
      "author": "Warusaki3",
      "version_date": "2008-04-14",
      "mugen_version": "1.0",
      "source_url": "https://mugenarchive.com/forums/downloads.php?id=1234",
      "source_package": "Warusaki_Capcom_Pack.zip",
      "license": "Fan-made / Non-commercial",
      "tags": ["Capcom", "CvS2", "6-Button"],
      "preview_image": ".previews/chars_ryu_warusaki.png",
      "notes": "Custom AI patched",
      "added_at": "2026-08-20T17:10:00Z"
    }
  }
}
```

---

## 4. Archive Ingestion & Project Dissection Pipeline

When users drop archives or folders into Ikemen Studio, the backend executes an automated zero-friction dissection pipeline:

```
[ Ingest: .zip / .rar / .7z / Folder ]
                  │
                  ▼
         [ Temp Unpack & Deep Scan ]
                  │
  ┌───────────────┴──────────────────┐
  ▼                                  ▼
[ Single Asset Archive ]     [ Multi-Asset Pack / Full Game ]
  │                                  │
  │ Auto-normalize subfolder         ├── Option A: Import into active vault (tracked by source_package)
  │ (e.g. kfm/kfm.def)               └── Option B: Convert directly into a dedicated new Vault
  │                                  │
  └───────────────┬──────────────────┘
                  ▼
      [ Metadata & Portrait Extraction ]
      • Read .def [Info] (name, author, version)
      • Scan readme/credits files for URLs & links
      • Extract SFF Sprite (9000,0 & 9000,1) to PNG
                  │
                  ▼
         [ Register in vault.json ]
```

### 4.1 Ingestion Scenarios

#### Scenario 1: Bulk Single Characters (e.g. 50 `.zip` files)
- Drag and drop 50 archives into the Vault view.
- Non-blocking background worker extracts and normalizes all 50 fighters.
- Populates `name`, `author`, `version`, and extracted portraits automatically.
- No blocking dialogs or mandatory fields.

#### Scenario 2: Multi-Asset Content Pack (e.g. 10 chars + 5 stages in one zip)
- Deep scanner parses the entire tree, separates characters into individual folders in `vault/chars/`, and stages into `vault/stages/`.
- Every extracted asset records `source_package: "pack_name.zip"`. Users can filter by source package at any time.

#### Scenario 3: Full Custom Game / Legacy MUGEN Build
- When dropping an entire standalone game project folder or archive:
  - Studio prompts:
    - *"Convert into a standalone Vault (Recommended for full games/mods)"* -> creates a new dedicated Vault named after the package.
    - *"Extract assets into current Vault"* -> dissects and tags with the package name.

---

## 5. Metadata & Portrait Extraction Engine

### 5.1 `.def` Parser
Scans character and stage definition files to extract:
* `name`, `displayname`
* `author`
* `versiondate`
* `mugenversion` / `ikemenversion`
* Top-of-file comments (extracting embedded author websites, changelogs, and notes).

### 5.2 Text & Link Scraping
* Scans `readme.txt`, `credits.txt`, `instructions.txt` for URLs matching `http://`, `https://`, and common repository/forum domains (Mugen Free For All, Mugen Archive, GitHub, itch.io).

### 5.3 SFF Portrait Extractor (`pkg/parser/sff`)
* Reads SFF v1 and SFF v2 binary headers.
* Extracts standard portrait sprites:
  * Group `9000`, Image `0` (Small 25x25 or custom Roster Icon).
  * Group `9000`, Image `1` (Large 120x140 or custom VS Portrait).
* Encodes and caches optimized `.png` thumbnails in `.previews/` for lightning-fast UI rendering.

---

## 6. Project Integration & Linking Engine

The Vault is accessed directly from inside the active project workspace.

```
Project Workspace ("Super Street Fighter IV Mugen")
    │
    ├── 1. User clicks "+ Add Fighter" (or "+ Add Stage")
    │
    ├── 2. Modal opens: Asset Vault Browser
    │      • Filter by Vault (All / Specific Vault)
    │      • Filter by Category / Tags
    │      • Search by Name / Author
    │      • Instant visual portrait cards
    │
    ├── 3. User selects 1 or multiple fighters -> clicks "Add to Roster"
    │
    └── 4. Linker connects the asset:
           • Creates link in project: `chars/kfm/` -> `vault/chars/kfm/`
           • Appends fighter to `data/select.def`
```

### 6.1 Linking Strategies
Configured globally in **Settings** (with smart fallback):

| Mode | Disk Space | Cross-Drive Support | Windows Admin Privileges Needed | Behavior |
| :--- | :--- | :--- | :--- | :--- |
| **Symlink (Default)** | 0 MB extra | Yes | No (Developer Mode enabled) / Fallback | Fast, edits in vault reflect across linked projects |
| **Hardlink** | 0 MB extra | Same drive only | No | Direct inode link, cannot break on folder rename |
| **Direct Copy** | Full size | Yes | No | Independent copy, zero external link dependencies |

---

## 7. Future Capabilities: Automatic Game Credits Generator

Because each Vault asset tracks author, source URL, license, and custom notes, and projects track which Vault assets are linked:
* Studio can compile a 1-click `CREDITS.md` or formatted in-game text file summarizing:
  * Character names & creators
  * Stages & creators
  * Custom sound & music licensing
  * Original download links and attributions
