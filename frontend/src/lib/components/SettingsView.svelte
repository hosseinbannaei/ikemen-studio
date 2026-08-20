<script lang="ts">
  import { onMount } from 'svelte';
  import { settingsStore, THEME_PRESETS } from '../stores/settingsStore';
  import type { Settings, ThemeId, RadiusStyle } from '../types';
  import { OpenFolderInExplorer } from '../../../wailsjs/go/main/App';
  import {
    Settings as SettingsIcon,
    Folder,
    FolderOpen,
    Save,
    Info,
    Check,
    Palette,
    Sparkles,
    Sliders,
    Square,
    Circle,
    CheckCircle2,
    Flame,
    Moon,
    Sun,
    Zap,
    Trophy
  } from 'lucide-svelte';

  let enginesDir = '';
  let saved = false;

  onMount(async () => {
    await settingsStore.load();
    enginesDir = $settingsStore.enginesDir;
  });

  async function handleBrowseEnginesDir() {
    const chosen = await settingsStore.chooseEnginesDir();
    if (chosen) {
      enginesDir = chosen;
    }
  }

  async function handleOpenEnginesDir() {
    if (enginesDir) {
      await OpenFolderInExplorer(enginesDir);
    }
  }

  async function handleSave() {
    const updated: Settings = {
      ...$settingsStore,
      enginesDir: enginesDir.trim(),
    };

    await settingsStore.save(updated);
    saved = true;
    setTimeout(() => {
      saved = false;
    }, 2500);
  }

  function getThemeIcon(id: ThemeId) {
    switch (id) {
      case 'mkx':
        return Flame;
      case 'obsidian':
        return Moon;
      case 'cyber':
        return Zap;
      case 'capcom':
        return Trophy;
      case 'light':
        return Sun;
      default:
        return Palette;
    }
  }
</script>

<div class="p-8 max-w-5xl mx-auto space-y-8">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-black tracking-tight text-slate-100 flex items-center gap-2.5">
        <Sliders class="w-6 h-6 text-brand-400" />
        <span>Preferences & Appearance</span>
      </h1>
      <p class="text-xs text-slate-400 mt-0.5">Customize global themes, corner geometry, and storage configuration</p>
    </div>

    <button
      type="button"
      class="flex items-center gap-2 px-5 py-2.5 rounded-xl bg-brand-600 hover:bg-brand-500 text-white text-xs font-bold shadow-md shadow-brand-600/30 transition"
      on:click={handleSave}
    >
      {#if saved}
        <Check class="w-4 h-4 text-emerald-300" />
        <span>Saved!</span>
      {:else}
        <Save class="w-4 h-4" />
        <span>Save Changes</span>
      {/if}
    </button>
  </div>

  <div class="space-y-6">
    <!-- 1. Global Themes Gallery -->
    <div class="p-6 rounded-2xl bg-dark-800/80 border border-dark-600/60 space-y-4 shadow-sm">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2.5">
          <Palette class="w-4 h-4 text-brand-400" />
          <div>
            <h2 class="text-xs font-bold uppercase tracking-wider text-slate-300">Studio Theme Presets</h2>
            <p class="text-xs text-slate-400 mt-0.5">Choose a dark fighting-game inspired palette for the interface</p>
          </div>
        </div>
      </div>

      <!-- Theme Cards Grid -->
      <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3.5 pt-1">
        {#each THEME_PRESETS as preset}
          {@const isSelected = $settingsStore.theme === preset.id || (!$settingsStore.theme && preset.id === 'mkx')}
          {@const IconComponent = getThemeIcon(preset.id)}
          <button
            type="button"
            class="group text-left p-4 rounded-xl border transition-all relative overflow-hidden flex flex-col justify-between {
              isSelected
                ? 'bg-dark-750 border-brand-500 ring-2 ring-brand-500/50 shadow-lg shadow-brand-950/50'
                : 'bg-dark-900/80 border-dark-600/70 hover:border-dark-500 hover:bg-dark-800'
            }"
            on:click={() => settingsStore.setTheme(preset.id)}
          >
            <!-- Card Top: Icon, Tag & Selected Badge -->
            <div class="flex items-center justify-between w-full mb-3">
              <div class="flex items-center gap-2">
                <div
                  class="w-7 h-7 rounded-lg flex items-center justify-center text-white shadow-sm"
                  style="background-color: {preset.accentHex};"
                >
                  <svelte:component this={IconComponent} class="w-4 h-4" />
                </div>
                <span class="text-[9px] font-mono font-bold px-1.5 py-0.2 rounded bg-black/40 text-slate-300 uppercase">
                  {preset.tag}
                </span>
              </div>

              {#if isSelected}
                <div class="flex items-center gap-1 text-[10px] font-bold text-brand-400">
                  <CheckCircle2 class="w-4 h-4" />
                  <span>Active</span>
                </div>
              {/if}
            </div>

            <!-- Card Middle: Color Palette Swatch -->
            <div class="flex items-center gap-1.5 p-1.5 rounded-lg bg-black/30 border border-white/5 mb-3 w-full">
              <div class="w-5 h-5 rounded border border-white/10 flex-shrink-0" style="background-color: {preset.bgHex};" title="Background"></div>
              <div class="w-5 h-5 rounded border border-white/10 flex-shrink-0" style="background-color: {preset.cardHex};" title="Surface / Card"></div>
              <div class="w-5 h-5 rounded border border-white/10 flex-shrink-0" style="background-color: {preset.borderHex};" title="Border"></div>
              <div class="flex-1 h-5 rounded border border-white/10 flex items-center justify-center text-[10px] font-bold text-white shadow-inner" style="background-color: {preset.accentHex};" title="Accent Accent">
                Accent
              </div>
            </div>

            <!-- Card Bottom: Name & Description -->
            <div>
              <div class="text-xs font-bold text-slate-100 group-hover:text-brand-300 transition">
                {preset.name}
              </div>
              <div class="text-[11px] font-medium text-slate-400">
                {preset.subtitle}
              </div>
              <p class="text-[10px] text-slate-500 mt-1 leading-tight">
                {preset.description}
              </p>
            </div>
          </button>
        {/each}
      </div>
    </div>

    <!-- 2. Corner Geometry / Brutalist Radii -->
    <div class="p-6 rounded-2xl bg-dark-800/80 border border-dark-600/60 space-y-3 shadow-sm">
      <div>
        <h2 class="text-xs font-bold uppercase tracking-wider text-slate-300">Corner Geometry</h2>
        <p class="text-xs text-slate-400 mt-0.5">
          Select corner radius style across buttons, cards, and dialogue panels.
        </p>
      </div>

      <div class="grid grid-cols-3 gap-3 pt-1">
        <!-- Sharp / Brutalist (MKX) -->
        <button
          type="button"
          class="p-3.5 rounded-sm border text-left transition {
            ($settingsStore.radiusStyle || 'sharp') === 'sharp'
              ? 'bg-dark-750 border-brand-500 ring-2 ring-brand-500/50'
              : 'bg-dark-900 border-dark-600/70 hover:border-dark-500'
          }"
          on:click={() => settingsStore.setRadius('sharp')}
        >
          <div class="flex items-center justify-between mb-1">
            <span class="text-xs font-bold text-slate-200">Sharp (MKX Brutalist)</span>
            {#if ($settingsStore.radiusStyle || 'sharp') === 'sharp'}
              <Check class="w-3.5 h-3.5 text-brand-400" />
            {/if}
          </div>
          <p class="text-[10px] text-slate-400">0px – 2px radius. Crisp technical fighting game arcade feel.</p>
        </button>

        <!-- Subtle Minimal -->
        <button
          type="button"
          class="p-3.5 rounded-md border text-left transition {
            $settingsStore.radiusStyle === 'subtle'
              ? 'bg-dark-750 border-brand-500 ring-2 ring-brand-500/50'
              : 'bg-dark-900 border-dark-600/70 hover:border-dark-500'
          }"
          on:click={() => settingsStore.setRadius('subtle')}
        >
          <div class="flex items-center justify-between mb-1">
            <span class="text-xs font-bold text-slate-200">Subtle Minimal</span>
            {#if $settingsStore.radiusStyle === 'subtle'}
              <Check class="w-3.5 h-3.5 text-brand-400" />
            {/if}
          </div>
          <p class="text-[10px] text-slate-400">4px – 6px radius. Balanced sleek modern geometry.</p>
        </button>

        <!-- Soft Rounded -->
        <button
          type="button"
          class="p-3.5 rounded-xl border text-left transition {
            $settingsStore.radiusStyle === 'rounded'
              ? 'bg-dark-750 border-brand-500 ring-2 ring-brand-500/50'
              : 'bg-dark-900 border-dark-600/70 hover:border-dark-500'
          }"
          on:click={() => settingsStore.setRadius('rounded')}
        >
          <div class="flex items-center justify-between mb-1">
            <span class="text-xs font-bold text-slate-200">Soft Rounded</span>
            {#if $settingsStore.radiusStyle === 'rounded'}
              <Check class="w-3.5 h-3.5 text-brand-400" />
            {/if}
          </div>
          <p class="text-[10px] text-slate-400">8px – 12px radius. Standard smooth curved UI.</p>
        </button>
      </div>
    </div>

    <!-- 3. Engine Storage Directory -->
    <div class="p-6 rounded-2xl bg-dark-800/80 border border-dark-600/60 space-y-3 shadow-sm">
      <div>
        <label for="settingsEnginesDir" class="block text-xs font-bold uppercase tracking-wider text-slate-300">
          Engine Storage Directory
        </label>
        <p class="text-xs text-slate-400 mt-0.5">
          Local directory where downloaded Ikemen GO binaries and runtime files are stored.
        </p>
      </div>

      <div class="flex gap-2.5">
        <input
          id="settingsEnginesDir"
          type="text"
          bind:value={enginesDir}
          placeholder="Storage path..."
          class="flex-1 px-4 py-2.5 rounded-xl bg-dark-900 border border-dark-600/80 text-xs font-mono text-slate-100 placeholder-slate-500 focus:outline-none focus:border-brand-500 transition"
        />
        <button
          type="button"
          class="px-4 py-2.5 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600 text-slate-200 text-xs font-semibold flex items-center gap-2 transition"
          on:click={handleBrowseEnginesDir}
          title="Choose a new engine directory"
        >
          <Folder class="w-4 h-4 text-brand-400" />
          <span>Browse</span>
        </button>

        <button
          type="button"
          class="px-4 py-2.5 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600 text-slate-200 text-xs font-semibold flex items-center gap-2 transition"
          on:click={handleOpenEnginesDir}
          title="Open folder in File Explorer"
        >
          <FolderOpen class="w-4 h-4 text-amber-400" />
          <span>Open</span>
        </button>
      </div>
    </div>

    <!-- 4. About Ikemen Studio -->
    <div class="p-6 rounded-2xl bg-dark-800/80 border border-dark-600/60 space-y-3 shadow-sm">
      <div class="flex items-center gap-2.5">
        <Info class="w-4 h-4 text-brand-400" />
        <span class="text-xs font-bold uppercase tracking-wider text-slate-300">About Studio</span>
      </div>

      <div class="grid grid-cols-2 gap-3 text-xs">
        <div class="p-4 rounded-xl bg-dark-900/60 border border-dark-600/40">
          <div class="text-slate-400 text-[10px] uppercase font-bold">Studio Version</div>
          <div class="text-slate-200 font-mono font-semibold mt-0.5 text-sm">v0.1.0 (Phase 1)</div>
        </div>
        <div class="p-4 rounded-xl bg-dark-900/60 border border-dark-600/40">
          <div class="text-slate-400 text-[10px] uppercase font-bold">Framework</div>
          <div class="text-slate-200 font-mono font-semibold mt-0.5 text-sm">Wails v2 + Svelte 5 + Tailwind</div>
        </div>
      </div>
    </div>
  </div>
</div>
