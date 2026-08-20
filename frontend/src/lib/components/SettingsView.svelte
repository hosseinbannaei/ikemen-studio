<script lang="ts">
  import { onMount } from 'svelte';
  import { settingsStore } from '../stores/settingsStore';
  import type { Settings } from '../types';
  import { OpenFolderInExplorer } from '../../../wailsjs/go/main/App';
  import {
    Settings as SettingsIcon,
    Folder,
    FolderOpen,
    Save,
    Info,
    Check,
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
</script>

<div class="p-8 max-w-4xl mx-auto space-y-8">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-black tracking-tight text-slate-100">Preferences</h1>
      <p class="text-xs text-slate-400 mt-0.5">Customize global storage paths and studio configuration</p>
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
          title="Choose a new engine directory"
        >
          <Folder class="w-4 h-4 text-indigo-400" />
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

    <!-- About Ikemen Studio -->
    <div class="p-6 rounded-2xl bg-dark-800/80 border border-dark-600/60 space-y-3 shadow-sm">
      <div class="flex items-center gap-2.5">
        <Info class="w-4 h-4 text-indigo-400" />
        <span class="text-xs font-bold uppercase tracking-wider text-slate-300">About Studio</span>
      </div>

      <div class="grid grid-cols-2 gap-3 text-xs">
        <div class="p-4 rounded-xl bg-dark-900/60 border border-dark-600/40">
          <div class="text-slate-400 text-[10px] uppercase font-bold">Studio Version</div>
          <div class="text-slate-200 font-mono font-semibold mt-0.5 text-sm">v0.1.0 (Phase 1)</div>
        </div>
        <div class="p-4 rounded-xl bg-dark-900/60 border border-dark-600/40">
          <div class="text-slate-400 text-[10px] uppercase font-bold">Framework</div>
          <div class="text-slate-200 font-mono font-semibold mt-0.5 text-sm">Wails v2 + Svelte 5</div>
        </div>
      </div>
    </div>
  </div>
</div>
