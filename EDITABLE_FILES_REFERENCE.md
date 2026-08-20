# Ikemen GO & MUGEN Editable Files & Asset Specification Reference

This document provides a comprehensive, exhaustive reference for all editable file formats, configurations, scripts, and asset types supported by **Ikemen Studio** and the **Ikemen GO** / **MUGEN** engine ecosystem.

---

## Quick Format Overview

| Extension / File | Category | Purpose | Syntax / Architecture |
| :--- | :--- | :--- | :--- |
| **`.def`** | Definition | Character, Stage, Motif, Lifebar, Storyboard, Font, Global FX | INI-style sections `[Section]` |
| **`.air`** | Animation | Sprite animation frames, timings, and collision boxes (`Clsn1`/`Clsn2`) | Custom Action blocks `[Begin Action <n>]` |
| **`.cmd`** | Commands | Input sequences, directional motions, buffer times, command triggers | INI sections `[Command]`, `[State -1]` |
| **`.cns`** | State Machine | Character physical constants, stats, victory quotes, states & controllers | INI `[Data]`, `[Size]`, `[StateDef <n>]` |
| **`.zss`** | ZSS Script | Modern Ikemen GO state scripting language (functions, loops, maps) | Modern ZSS syntax |
| **`.sff`** | Sprite Archive | Graphic storage (v1 PCX indexed, v2 PNG/RLE5/LZ5/RGBA, Group 9000) | Binary Sprite Archive |
| **`.snd`** | Sound Archive | Sound effects, voice lines, and hit sounds mapped by group and index | Binary WAV Audio Archive |
| **`.act`** | Color Palette | 256-color palette tables for 1P–12P character outfit variations | Raw 768-byte RGB Table |
| **`.fnt` / `.ttf`** | Typography | Bitmap font glyph maps and vector TrueType/OpenType typography | Bitmap FNT or TrueType DEF |
| **`.glb` / `.gltf`** | 3D Stage Model| 3D stage background meshes and environments for modern arenas | Binary GLTF 2.0 3D Mesh |
| **`.frag` / `.vert`**| Shaders | Screen filters (CRT, Scanlines, HQ2x, HQ4x) & stage distortion shaders | GLSL 330 / SPIR-V bytecode |
| **`select.def`** | Roster Config | Character select screen order, stage pairings, arcade rules | INI-style `[Characters]`, `[ExtraStages]` |
| **`config.ini`** | Engine Config | Resolution, OpenGL/Vulkan renderers, MSAA, VSync, audio volumes | INI `[Options]` |
| **`gamecontrollerdb.txt`**| Input DB | SDL2 Gamepad GUID mappings and button assignments | SDL2 CSV format |
| **`.lua`** | Engine Script | Main game loop, UI menus, story arcs, training overlays | Lua 5.1/LuaJIT |

---

## 1. Character Definition Files (`.def`, `.cns`, `.cmd`, `.air`, `.zss`, `.act`)

### 1.1 Character Master Definition (`<character>.def`)
The central manifest that binds all character assets together.

* **`[Info]`**:
  * `name`: Internal character identifier (e.g. `"Cammy"`).
  * `displayname`: Screen name shown in menus and lifebars (e.g. `"Cammy White"`).
  * `versiondate`: Release date (e.g. `05,20,2026`).
  * `mugenversion`: Compatible MUGEN version (`1.0`, `1.1`).
  * `ikemenversion`: Target Ikemen GO version (`0.99`, `1.0`).
  * `author`: Creator / editor name.
  * `pal.defaults`: Default palette order (e.g. `1,2,3,4,5,6`).
  * `localcoord`: Coordinate resolution space (e.g. `320,240`, `1280,720`).
  * `portraitscale`: Scaling factor for portrait rendering.
* **`[Files]`**:
  * `cmd`: Command definition file (`<char>.cmd`).
  * `cns`: Physical constants and base parameters (`<char>.cns`).
  * `anim`: Animation file (`<char>.air`).
  * `sprite`: Sprite archive (`<char>.sff`).
  * `sound`: Sound effects and voice clips (`<char>.snd`).
  * `movelist`: In-game pause command list (`movelist.dat`).
  * `st`: Main state file (`<char>.cns` or `<char>.zss`).
  * `st1` to `st9`: Modular auxiliary state files (e.g. specials, supers, helpers).
  * `stcommon`: Common universal state overrides (e.g. `common1.cns` or `data/action.zss`).
  * `pal1` to `pal12`: Individual palette files (`<char>_01.act`).
* **`[Palette Keymap]`**:
  * `a`, `b`, `c`, `x`, `y`, `z`: Maps specific controller buttons to palette numbers 1–12.
  * `a2`, `b2`, `c2`, `x2`, `y2`, `z2`: Hold-start button palette mappings.
* **`[Arcade]`**:
  * `intro.storyboard`: Opening cinematic storyboard (`intro.def`).
  * `ending.storyboard`: Arcade victory cinematic storyboard (`ending.def`).
  * `victoryscreen`: Character victory portrait or custom dialogue storyboard.

---

### 1.2 Animation & Collision Box Definition (`<character>.air`)
Defines frame-by-frame sprite playback, durations, looping, and hit/hurt collision boxes.

* **`[Begin Action <n>]`**: Declares an animation sequence where `<n>` is the Action ID:
  * **Stand**: `0`, **Walk Forward**: `20`, **Walk Back**: `21`, **Jump**: `40`, **Crouch**: `10`.
  * **Basic Attacks**: `200`–`499`, **Special Moves**: `1000`–`1999`, **Super Moves**: `3000`–`3999`.
  * **Hit / Hurt States**: `5000`–`5999`.
* **Collision Boxes**:
  * `Clsn2: <count>`: **Hurtboxes** (vulnerable areas, blue boxes in debug).
    * `Clsn2[0] = x1, y1, x2, y2`
  * `Clsn1: <count>`: **Hitboxes** (active attack strike boxes, red boxes in debug).
    * `Clsn1[0] = x1, y1, x2, y2`
* **Frame Line Syntax**:
  * `group, image, x_offset, y_offset, duration_in_ticks, [flip_flags], [blend_flags]`
  * `Loopstart`: Marks the frame loop point.
  * `flip_flags`: `H` (Horizontal flip), `V` (Vertical flip), `HV` (Both).
  * `blend_flags`: `A` (Additive alpha 50/50), `A1` (Full additive), `S` (Subtractive), `ASxxxDyyy` (Custom alpha source/dest).

---

### 1.3 Command Definition (`<character>.cmd`)
Translates player gamepad/keyboard inputs into named triggers and handles AI inputs.

* **`[Remap]`**: Button aliases (e.g. `x = x`, `y = y`, `z = z`).
* **`[Defaults]`**:
  * `command.time = 15`: Window in game ticks to complete the input sequence.
  * `command.buffer.time = 1`: Window in game ticks to hold the triggered command buffer.
* **`[Command]`**:
  * `name = "QCF_x"`: Name referenced in state triggers.
  * `command = ~D, DF, F, x`: Input sequence (`~` release trigger, `/$` hold directions, `>` sequential without delay).
  * `time = 15`, `buffer.time = 3`.
* **`[State -1]` (Command Controller Block)**:
  * Evaluates commands and activates states via `ChangeState`:
    ```ini
    [State -1, Hadouken]
    type = ChangeState
    value = 1000
    triggerall = command = "QCF_x"
    trigger1 = statetype != A && ctrl
    ```

---

### 1.4 Character Physics, Constants & State Machine (`<character>.cns`)
Defines character stats, physics attributes, and state logic.

* **`[Data]`**:
  * `life`: Starting maximum health (standard: `1000`).
  * `power`: Starting super meter (standard: `3000`).
  * `attack`: Base damage scaling percentage (standard: `100`).
  * `defence`: Defense scaling percentage (standard: `100`).
  * `fall.defence_up`: Defense boost when grounded after falling.
  * `liedown.time`: Ticks grounded before recovering.
  * `airjuggle`: Maximum juggle points per aerial combo.
  * `sparkno`, `guard.sparkno`: Default hitspark and guardspark animation IDs.
* **`[Size]`**:
  * `xscale`, `yscale`: Sprite rendering multiplier (`1.0`).
  * `ground.back`, `ground.front`: Ground boundary width for pushing other fighters.
  * `air.back`, `air.front`: Aerial collision width.
  * `height`: Character height for jump collision.
  * `head.pos`, `mid.pos`: Coordinates for spark placement and projectile targeting.
  * `shadowoffset`: Ground shadow vertical offset.
* **`[Velocity]`**:
  * `walk.fwd`, `walk.back`: Horizontal walking speed.
  * `run.fwd`, `run.back`: Dash speeds.
  * `jump.neu`, `jump.fwd`, `jump.back`: Jump velocities (`x, y`).
* **`[Movement]`**:
  * `airjump.num`: Number of aerial double-jumps.
  * `yaccel`: Gravity acceleration (`0.44`).
  * `stand.friction`, `crouch.friction`: Ground deceleration rate.
* **`[StateDef <n>]`**:
  * `type = S/C/A/L` (Stand, Crouch, Air, Liedown).
  * `movetype = A/I/H` (Attack, Idle, Hit).
  * `physics = S/C/A/N` (Stand, Crouch, Air, None).
  * `anim = <action_no>`: Initial animation to play.
  * `velset = <x>, <y>`: Initial velocity.
  * `ctrl = 0/1`: Player control enabled.
  * `poweradd = <n>`: Super meter gain.
* **State Controllers (`[State <id>]`)**:
  * `HitDef`: Complete attack strike definition (damage, hitflag, guardflag, pausetime, hitshake, spark, knockback, fall).
  * `ChangeState`: Transitions to another state.
  * `ChangeAnim`: Swaps animation without changing state.
  * `PlaySnd`: Plays audio from character or system `.snd`.
  * `Projectile`: Spawns a dedicated projectile object.
  * `Helper`: Spawns an autonomous sub-entity (assist characters, persistent projectiles, clones).
  * `Explod`: Spawns custom visual VFX, sparks, and after-images.
  * `VarSet` / `VarAdd`: Modifies internal integer/float variables.
  * `AssertSpecial`: Disables engine features (e.g. `noko`, `nobardisplay`, `invisible`, `nomusic`).
  * `EnvShake`: Triggers camera / screen rumble.
  * `PalFX` / `AfterImage`: Palette flashing and motion blur trails.

---

### 1.5 Modern Ikemen GO Scripting (`.zss`)
Ikemen GO native high-performance scripting language with modern syntax:

* Native variable typing (`let`, `var`, `fvar`, `map`, `const`).
* Native looping constructs (`for`, `while`) and nested condition blocks (`if / else if / else`).
* Direct function calls (`call FuncName()`).
* Global event controllers:
  * `[StateDef -4]`: Global engine-level state (not frozen by Pause/SuperPause).
  * `[StateDef -3]`: Character state executed every tick.
  * `[StateDef -2]`: Character state executed every tick during gameplay.
  * `[StateDef -1]`: Command interpreter block.

---

## 2. Stage Definition Files (`<stage>.def`)

Stages represent battle environments and can be 2D parallax arenas, 3D glTF/GLB environments, or interactive stages.

* **`[Info]`**:
  * `name`, `displayname`, `versiondate`, `author`, `mugenversion`, `ikemenversion`.
  * `attachedChar`: Sub-character helper that controls interactive elements or stage hazards.
* **`[Camera]`**:
  * `startx`, `starty`: Camera starting position.
  * `boundleft`, `boundright`: Horizontal panning boundaries.
  * `boundhigh`, `boundlow`: Vertical scrolling boundaries.
  * `verticalfollow`: Vertical camera tracking speed (`0.0` to `1.0`).
  * `tension`: Distance player reaches toward screen edge before camera scrolls.
  * `startzoom`, `zoomin`, `zoomout`: Camera dynamic zoom levels.
  * `fov`: Field of view for 3D stages (e.g. `30`–`45`).
* **`[PlayerInfo]`**:
  * `p1startx`, `p1starty`, `p1facing`: Player 1 initial position.
  * `p2startx`, `p2starty`, `p2facing`: Player 2 initial position.
  * `leftbound`, `rightbound`: Physical walls constraining player movement.
* **`[StageInfo]`**:
  * `zoffset`: Ground baseline where character feet touch the floor.
  * `localcoord`: Coordinate resolution space (e.g. `1280, 720`).
  * `xscale`, `yscale`: Global stage rendering scale.
  * `portraitscale`: Stage select preview icon scaling.
* **`[Shadow]` & `[Reflection]`**:
  * `intensity`: Shadow/reflection alpha (`0`–`256`).
  * `color`: RGB color tint (e.g. `0, 0, 0`).
  * `yscale`: Shadow height / projection angle.
  * `fade.range`: Elevation range where shadow fades out.
* **`[Music]`**:
  * `bgmusic`: Path to audio track (`sound/theme.mp3`, `sound/theme.ogg`).
  * `bgmvolume`: Playback volume (`0`–`100`).
  * `bgmloopstart`, `bgmloopend`: Audio sample points for seamless loop.
* **`[BGdef]`**:
  * `spr`: Stage sprite archive (`<stage>.sff`).
  * `model`: 3D model file (`<stage>.glb` or `<stage>.gltf`).
  * `debugbg`: 1 to enable magenta debug background.
* **`[Model]` (3D Stages)**:
  * `offset`: X, Y, Z coordinates of the 3D model.
  * `scale`: X, Y, Z scaling factors.
* **`[BG <name>]` (2D Layers)**:
  * `type`: `normal` (static/parallax image), `parallax` (perspective floor/sky), `anim` (animated sequence).
  * `spriteno`: Group and image index (`0, 0`).
  * `actionno`: Animation action ID in stage `.def`.
  * `start`: Initial coordinate offset (`x, y`).
  * `delta`: Parallax scroll speed relative to camera movement (`x_delta, y_delta`).
  * `trans`: Blending mode (`none`, `add`, `addalpha`, `sub`).
  * `alpha`: Source and destination alpha levels (`src, dst`).
  * `tile`: Tiling repetition (`x_tile, y_tile`).
  * `tilespacing`: Distance between repetitions (`x, y`).
  * `velocity`: Continuous scrolling velocity (`x, y`).
  * `layerno`: `0` for background (behind players), `1` for foreground (in front of players), `-1` for deep backdrop.
* **`[BGCtrlDef <name>]` & `[BGCtrl <name>]`**:
  * Controls dynamic background movement, color pulsing, and triggered events via `Enable`, `PosSet`, `PosAdd`, `VelSet`, `Anim`, `Visible`, `SinX`, `SinY`.

---

## 3. Screenpack & Motif Definitions (`system.def`)

Controls the visual theme, menus, layouts, and character select grid of the entire game.

* **`[Info]`**:
  * `name`: Screenpack name (e.g. `"Mugen 1.1 HD"`, `"Ikemen Default"`).
  * `author`, `versiondate`, `localcoord` (`1280, 720`).
* **`[Files]`**:
  * `spr`: System sprite archive (`system.sff`).
  * `snd`: System UI audio sound archive (`system.snd`).
  * `select`: Character selection roster definition (`select.def`).
  * `fight`: In-game fight HUD and lifebar definition (`fight.def`).
  * `font1` to `fontN`: Registered UI fonts.
  * `logo.storyboard`: Opening studio logo cinematic (`logo.def`).
  * `intro.storyboard`: Opening game cinematic (`intro.def`).
  * `module`: Optional external Lua script extension.
* **`[Music]`**:
  * `title.bgm`, `select.bgm`, `vs.bgm`, `victory.bgm`, `option.bgm`, `continue.bgm`, `results.bgm`, `hiscore.bgm`.
  * `.volume`, `.loop`, `.loopstart`, `.loopend` settings for each screen.
* **`[Title Info]`**:
  * `menu.pos`: Screen coordinates of main menu items.
  * `menu.item.font`, `menu.item.active.font`: Font slot and color codes.
  * `menu.item.spacing`: Vertical distance between menu rows.
  * `menu.itemname.<mode>`: Label text for all submenus (Arcade, Versus, Team Versus, Story Mode, Network, Training, Watch, Options, Exit).
  * `cursor.move.snd`, `cursor.done.snd`, `cancel.snd`: UI sound effect mappings.
* **`[Select Info]` (Roster Grid Architecture)**:
  * `rows`, `columns`: Character select screen grid dimensions (e.g. `6, 12` = 72 slots).
  * `wrapping`: 1 allows cursor to wrap around grid borders.
  * `showemptyboxes`: 1 displays placeholder frames for empty slots.
  * `pos`: X, Y screen coordinates of the roster grid.
  * `cell.size`: Width and height of each character icon cell.
  * `cell.spacing`: Pixel gap between cells.
  * `portrait.offset`, `portrait.scale`: Small icon rendering parameters.
  * `p1.face.offset`, `p1.face.scale`: Large character preview artwork positioning.
  * `stage.pos`, `stage.active.font`: Stage selector widget positioning.
* **`[VS Screen]`**:
  * Match loading screen layouts, portrait positions, versus artwork, and stage banner font.
* **`[Victory Screen]`**:
  * Post-match winner screen, win quotes, character artwork, and timeout duration.
* **`[Option Info]`**:
  * In-game settings menu layout, slider fonts, and navigation parameters.

---

## 4. Fight HUD & Lifebar Definitions (`fight.def`)

Controls the in-match graphical user interface, health meters, super gauges, announcers, and timers.

* **`[Info]` & `[Files]`**:
  * `sff`: Fight sprite archive (`fight.sff`).
  * `snd`: Announcer and hit sound archive (`fight.snd`).
  * `fightfx.sff`, `fightfx.air`: Universal common hitsparks and dust effects.
  * `common.snd`: Common block and fall sounds.
  * `font1` to `fontN`: Lifebar timer, combo, and name fonts.
* **`[Lifebar]`**:
  * `p1.pos`, `p2.pos`: Lifebar anchor positions.
  * `p1.bg0.spr`: Lifebar container frame background.
  * `p1.mid.spr`: Red damage indicator bar (slowly empties).
  * `p1.front.spr`: Active health bar fill texture.
  * `p1.range.x`: Pixel start and end range for health bar draining.
  * `p1.xshear`: Slanted bar angle (e.g. `3` or `-3`).
  * `p1.scalefill`: 1 scales the sprite; 0 crops the sprite.
* **`[Powerbar]`**:
  * Super meter gauge styling, level indicators, level 1/2/3 counter fonts, and flashing effects.
* **`[Guardbar]` & `[Stunbar]`**:
  * Ikemen GO guard break and stun / dizzy meters.
* **`[Face]` & `[Name]`**:
  * In-match character icon portraits (Sprite `9000, 0`) and name font placement.
* **`[Time]`**:
  * Match countdown timer position, tick speed (`framespercount = 60`), and background frames.
* **`[Combo]`**:
  * Hit counter text, positioning, digit animations, and shake duration.
* **`[Round]`**:
  * Round announcements ("Round 1", "Fight!", "K.O.", "Double K.O.", "Time Over", "Draw Game", "Perfect!").
  * Sound effects, display times, and fade animations for each announcer banner.
* **`[WinIcon]`**:
  * Round victory badges (Normal, Special, Hyper, Throw, Time Over, Perfect icons).

---

## 5. Storyboard & Cinematic Definitions (`*.def`)

Controls opening cinematics, character story intros, arcade endings, and credit rolls.

* **`[Info]`**:
  * `localcoord`: Coordinate resolution (`1280, 720`).
* **`[SceneDef]`**:
  * `spr`: Storyboard sprite archive (`intro.sff`).
  * `snd`: Audio archive for voice lines and sound effects (`intro.snd`).
  * `font1` to `fontN`: Subtitle fonts.
* **`[Scene <n>]`**:
  * `fadein.time`, `fadeout.time`: Fade transitions in game ticks.
  * `end.time`: Total duration of the scene in ticks.
  * `bgm`: Background music track for this scene.
  * `layerall.pos`: Base coordinate offset.
  * `layer0.anim`: Action ID in storyboard `.def` or sprite `layer0.spr`.
  * `text`: Subtitle dialogue or narrative text (`\n` for newlines).
  * `text.font`: Font index, bank, alignment, and RGB color.
  * `text.offset`: Screen position of subtitle.
  * `text.delay`: Delay between character typing animation.
* **`[BG <name>]` (Video Scenes)**:
  * `type = video`: Enables native video playback (`webm`, `mp4`).
  * `path = video/intro.webm`: Relative path to video file.

---

## 6. Typography & Font Specifications (`.def`, `.fnt`, `.ttf`)

* **Bitmap Fonts (`.fnt` / `.def`)**:
  * Pre-rendered pixel font glyphs.
  * `Type = bitmap`: Fixed or variable glyph width table (`[Font <ascii>] = x, y, w, h`).
  * Bank color shifts for different team/state colors.
* **TrueType Fonts (`.def` with TrueType Mode)**:
  * Vector scalable fonts using `.ttf` or `.otf` files.
  * `Type = truetype`
  * `Size = <width>, <height>`: Primary height scale.
  * `Spacing = <x>, <y>`: Letter spacing and line height.
  * `File = Open_Sans/OpenSans-Bold.ttf`: Path to font file.
  * `Blend = 1`: Anti-aliasing alpha blending.

---

## 7. Roster & Arcade Configuration (`data/select.def`)

The master configuration defining which fighters appear on the select screen and in arcade ladders.

* **`[Characters]`**:
  * Format: `<char_name_or_folder>, [stage_path], [order=<n>], [music=<bgm_path>], [includestage=<0/1>], [vsscreen=<0/1>], [victoryscreen=<0/1>], [ai=<level>]`
  * Special Keywords:
    * `randomselect`: Random character selector cell.
    * `empty`: Empty grid space (creates spacing or custom grid patterns).
* **`[ExtraStages]`**:
  * List of standalone stage `.def` files available in training, versus, and survival modes.
* **`[Options]`**:
  * `arcade.maxmatches = 6, 1, 1, 0, 0, 0, 0, 0, 0, 0`: Number of fighters fought per `order=<n>` tier in single-player Arcade mode.
  * `team.maxmatches = 4, 1, 1, 0, 0, 0, 0, 0, 0, 0`: Ladder matches for Team Arcade.
  * `survival.maxmatches = -1`: Unlimited survival matches.

---

## 8. Game Configuration (`save/config.ini`)

Controls native engine execution, audio, display, rendering, and gameplay settings.

* **`[Options]`**:
  * `Width = 1280`, `Height = 720`: Native render resolution.
  * `Fullscreen = 0`: 1 for fullscreen, 0 for windowed.
  * `Vsync = 1`: Vertical synchronization.
  * `RenderMode = "OpenGL"`: Engine backend (`OpenGL` or `Vulkan`).
  * `MSAA = 4`: Multi-sample anti-aliasing level (`0`, `2`, `4`, `8`).
  * `Motif = "data/ikemen1/system.def"`: Active screenpack definition.
  * `VolumeMaster = 80`, `VolumeBgm = 70`, `VolumeSfx = 80`: Audio channels (0–100).
  * `GameSpeed = 60`: Target engine frames per second (FPS).
  * `Difficulty = 4`: AI difficulty rating (1–8).
  * `Life = 100`: Base health percentage modifier.
  * `Time = 99`: Match round timer duration in seconds (-1 for infinite).
  * `AutoGuard = 0`: 1 to enable automatic guarding for beginners.
  * `Team1VS2Life = 120`: Handicap modifier for team modes.
  * `Shader = "external/shaders/HQ2x.frag"`: Active post-processing shader.

---

## 9. Controller Mapping (`external/gamecontrollerdb.txt`)

Maps physical hardware gamepads, fightsticks, and arcade controllers via SDL2 GameController database syntax.

* **Entry Format**: `<GUID>,<Controller Name>,<Button & Axis Mappings>,platform:<OS>`
* **Buttons**: `a`, `b`, `x`, `y`, `back`, `start`, `guide`, `leftshoulder`, `rightshoulder`, `leftstick`, `rightstick`, `dpup`, `dpdown`, `dpleft`, `dpright`, `lefttrigger`, `righttrigger`.
* **Axes**: `leftx:a0`, `lefty:a1`, `rightx:a2`, `righty:a3`.

---

## 10. Lua Engine Scripting (`external/script/*.lua`)

Ikemen GO allows deep modification of game flow and UI through its embedded Lua runtime.

* **`main.lua`**: Engine bootstrap, game loop, mode dispatcher, and global helper functions.
* **`options.lua`**: In-game settings interface, key rebinding logic, resolution presets.
* **`storymode.lua`**: Custom story mode arcs, dialogue trees, and branching paths.
* **`training.lua`**: Training mode overlays, input history display, dummy AI behavior.
* **`menu.lua`**: Custom main menu entries and modal dialogs.
* **`debug.lua`**: Developer tools, hitbox visualizations, frame advantage calculators.

---

## Summary of Studio Tooling Support

All of the above file types can be:
1. **Inspected Structurally**: View sections, key-values, frame counts, command buffers, and state counts with `InspectProjectFile`.
2. **Safely Read & Edited**: Live edit any `.def`, `.cns`, `.zss`, `.cmd`, `.air`, `.ini`, `.lua`, `.txt` file using the built-in **Universal File Inspector & Code Editor** (`FileEditorModal.svelte`).
3. **Automated & Repaired**: Validated and repaired across character ingestion, asset synchronization, motif switching, and runtime execution.
