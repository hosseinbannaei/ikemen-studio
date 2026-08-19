<script lang="ts">
  import { onMount } from 'svelte';
  import { settingsStore } from '../stores/settingsStore';
  import { projectStore } from '../stores/projectStore';
  import type { Settings } from '../types';
  import {
    Settings as SettingsIcon,
    Folder,
    Save,
    Sliders,
    Moon,
    Sun,
    Info,
    Trash2,
    HardDrive,
    Check,
  } from 'lucide-svelte';

  let enginesDir = '';
  let theme = 'dark';
  let defaultChannel = 'stable';
  let saved = false;

  onMount(async () => {
    await settingsStore.load();
    enginesDir = $settingsStore.enginesDir;
    theme = $settingsStore.theme;
    defaultChannel = $settingsStore.defaultChannel;
  });

  async function handleBrowseEnginesDir() {
    const chosen = await settingsStore.chooseEnginesDir();
    if (chosen) {
      enginesDir = chosen;
    }
  }

  async function handleSave() {
    const updated: Settings = {
      ...$settingsStore,
      enginesDir: enginesDir.trim(),
      theme,
      defaultChannel,
    };

    await settingsStore.save(updated);
    saved = true;
    setTimeout(() => {
      saved = false;
    }, 2500);
  }
</script>

<div class="p-8 max-w-4xl mx-auto space-y-8">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-black tracking-tight text-slate-100">Preferences</h1>
      <p class="text-xs text-slate-400 mt-0.5">Customize global paths, default engine channels, and studio settings</p>
    </div>

    <button
      type="button"
      class="flex items-center gap-2 px-5 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold shadow-md shadow-indigo-600/30 transition"
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
    <!-- Engine Storage Directory -->
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
          class="flex-1 px-4 py-2.5 rounded-xl bg-dark-900 border border-dark-600/80 text-xs font-mono text-slate-100 placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition"
        />
        <button
          type="button"
          class="px-4 py-2.5 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600 text-slate-200 text-xs font-semibold flex items-center gap-2 transition"
          on:click={handleBrowseEnginesDir}
        >
          <Folder class="w-4 h-4 text-indigo-400" />
          <span>Browse</span>
        </button>
      </div>
    </div>

    <!-- Default Release Channel -->
    <div class="p-6 rounded-2xl bg-dark-800/80 border border-dark-600/60 space-y-3 shadow-sm">
      <div>
        <span class="block text-xs font-bold uppercase tracking-wider text-slate-300">
          Preferred Release Channel
        </span>
        <p class="text-xs text-slate-400 mt-0.5">
          Default version type selected when scaffolding new fighting game projects.
        </p>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <button
          type="button"
          class="p-4 rounded-xl border text-left flex items-center gap-3.5 transition {
            defaultChannel === 'stable'
              ? 'border-indigo-500 bg-indigo-950/20 text-indigo-300 shadow-sm'
              : 'border-dark-600/80 bg-dark-900/60 text-slate-400 hover:text-slate-200'
          }"
          on:click={() => (defaultChannel = 'stable')}
        >
          <div class="w-3 h-3 rounded-full {defaultChannel === 'stable' ? 'bg-indigo-400 ring-4 ring-indigo-500/20' : 'bg-dark-600'}"></div>
          <div>
            <div class="text-xs font-bold text-slate-200">Stable Channel</div>
            <div class="text-[11px] opacity-70">Recommended for finalized games and tournaments</div>
          </div>
        </button>

        <button
          type="button"
          class="p-4 rounded-xl border text-left flex items-center gap-3.5 transition {
            defaultChannel === 'nightly'
              ? 'border-indigo-500 bg-indigo-950/20 text-indigo-300 shadow-sm'
              : 'border-dark-600/80 bg-dark-900/60 text-slate-400 hover:text-slate-200'
          }"
          on:click={() => (defaultChannel = 'nightly')}
        >
          <div class="w-3 h-3 rounded-full {defaultChannel === 'nightly' ? 'bg-indigo-400 ring-4 ring-indigo-500/20' : 'bg-dark-600'}"></div>
          <div>
            <div class="text-xs font-bold text-slate-200">Nightly / Bleeding-Edge</div>
            <div class="text-[11px] opacity-70">Latest features, Lua additions, and active experimental updates</div>
          </div>
        </button>
      </div>
    </div>

    <!-- Theme Selection -->
    <div class="p-6 rounded-2xl bg-dark-800/80 border border-dark-600/60 space-y-3 shadow-sm">
      <div>
        <span class="block text-xs font-bold uppercase tracking-wider text-slate-300">
          Studio Appearance
        </span>
        <p class="text-xs text-slate-400 mt-0.5">
          Select interface theme.
        </p>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <button
          type="button"
          class="p-4 rounded-xl border text-left flex items-center gap-3.5 transition {
            theme === 'dark'
              ? 'border-indigo-500 bg-indigo-950/20 text-indigo-300 shadow-sm'
              : 'border-dark-600/80 bg-dark-900/60 text-slate-400 hover:text-slate-200'
          }"
          on:click={() => (theme = 'dark')}
        >
          <Moon class="w-4 h-4 text-indigo-400" />
          <div>
            <div class="text-xs font-bold text-slate-200">Dark Theme</div>
            <div class="text-[11px] opacity-70">Optimized for low-light game development</div>
          </div>
        </button>

        <button
          type="button"
          class="p-4 rounded-xl border text-left flex items-center gap-3.5 opacity-50 cursor-not-allowed border-dark-600/60 bg-dark-900/40 text-slate-400"
          title="Light theme planned for Phase 2"
        >
          <Sun class="w-4 h-4" />
          <div>
            <div class="text-xs font-bold text-slate-300">Light Theme</div>
            <div class="text-[11px] opacity-70">Planned for Phase 2</div>
          </div>
        </button>
      </div>
    </div>

    <!-- About Ikemen Studio -->
    <div class="p-6 rounded-2xl bg-dark-800/80 border border-dark-600/60 space-y-3 shadow-sm">
      <div class="flex items-center gap-2.5">
        <Info class="w-4 h-4 text-indigo-400" />
        <span class="text-xs font-bold uppercase tracking-wider text-slate-300">About Studio</span>
      </div>

      <div class="grid grid-cols-2 sm:grid-cols-3 gap-3 text-xs">
        <div class="p-3 rounded-xl bg-dark-900/60 border border-dark-600/40">
          <div class="text-slate-500 text-[10px] uppercase font-bold">Studio Version</div>
          <div class="text-slate-200 font-mono font-semibold mt-0.5">v0.1.0 (Phase 1)</div>
        </div>
        <div class="p-3 rounded-xl bg-dark-900/60 border border-dark-600/40">
          <div class="text-slate-500 text-[10px] uppercase font-bold">Framework</div>
          <div class="text-slate-200 font-mono font-semibold mt-0.5">Wails v2 + Svelte 5</div>
        </div>
        <div class="p-3 rounded-xl bg-dark-900/60 border border-dark-600/40">
          <div class="text-slate-500 text-[10px] uppercase font-bold">Target Engine</div>
          <div class="text-slate-200 font-mono font-semibold mt-0.5">Ikemen GO v0.99+</div>
        </div>
      </div>
    </div>
  </div>
</div>
