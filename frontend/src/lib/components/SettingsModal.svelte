<script lang="ts">
  import { onMount } from 'svelte';
  import { settingsStore } from '../stores/settingsStore';
  import type { Settings } from '../types';
  import { X, Settings as SettingsIcon, Folder, Save, Sliders, Moon, Sun } from 'lucide-svelte';

  export let onClose: () => void;

  let enginesDir = '';
  let theme = 'dark';
  let defaultChannel = 'stable';

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
    onClose();
  }
</script>

<div class="fixed inset-0 z-40 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
  <div class="bg-dark-800 border border-dark-600/80 rounded-2xl w-full max-w-lg shadow-2xl overflow-hidden">
    <!-- Header -->
    <div class="p-5 border-b border-dark-600/60 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <div class="w-9 h-9 rounded-xl bg-indigo-500/10 border border-indigo-500/30 flex items-center justify-center text-indigo-400">
          <SettingsIcon class="w-5 h-5" />
        </div>
        <div>
          <h2 class="text-base font-bold text-slate-100">Preferences</h2>
          <p class="text-xs text-slate-400">Configure global directories & engine options</p>
        </div>
      </div>
      <button
        type="button"
        class="p-2 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-dark-700 transition"
        on:click={onClose}
      >
        <X class="w-5 h-5" />
      </button>
    </div>

    <!-- Content -->
    <div class="p-5 space-y-4">
      <!-- Engines Directory -->
      <div>
        <label for="enginesDir" class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
          Engine Storage Directory
        </label>
        <div class="flex gap-2">
          <input
            id="enginesDir"
            type="text"
            bind:value={enginesDir}
            placeholder="Engines path..."
            class="flex-1 px-3.5 py-2.5 rounded-xl bg-dark-900 border border-dark-600 text-xs font-mono text-slate-100 placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition"
          />
          <button
            type="button"
            class="px-3.5 py-2.5 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600 text-slate-200 text-xs font-medium flex items-center gap-1.5 transition"
            on:click={handleBrowseEnginesDir}
          >
            <Folder class="w-4 h-4 text-indigo-400" />
            Browse
          </button>
        </div>
        <p class="text-[11px] text-slate-500 mt-1">Where downloaded Ikemen GO binaries and assets are saved.</p>
      </div>

      <!-- Default Release Channel -->
      <div>
        <span class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
          Default Engine Channel
        </span>
        <div class="grid grid-cols-2 gap-3">
          <button
            type="button"
            class="p-3 rounded-xl border text-left flex items-center gap-3 transition {
              defaultChannel === 'stable'
                ? 'border-indigo-500 bg-indigo-950/20 text-indigo-300'
                : 'border-dark-600 bg-dark-900/60 text-slate-400 hover:text-slate-200'
            }"
            on:click={() => (defaultChannel = 'stable')}
          >
            <div class="w-2.5 h-2.5 rounded-full {defaultChannel === 'stable' ? 'bg-indigo-400' : 'bg-dark-600'}"></div>
            <div>
              <div class="text-xs font-semibold">Stable</div>
              <div class="text-[10px] opacity-70">Official releases</div>
            </div>
          </button>

          <button
            type="button"
            class="p-3 rounded-xl border text-left flex items-center gap-3 transition {
              defaultChannel === 'nightly'
                ? 'border-indigo-500 bg-indigo-950/20 text-indigo-300'
                : 'border-dark-600 bg-dark-900/60 text-slate-400 hover:text-slate-200'
            }"
            on:click={() => (defaultChannel = 'nightly')}
          >
            <div class="w-2.5 h-2.5 rounded-full {defaultChannel === 'nightly' ? 'bg-indigo-400' : 'bg-dark-600'}"></div>
            <div>
              <div class="text-xs font-semibold">Nightly</div>
              <div class="text-[10px] opacity-70">Pre-release builds</div>
            </div>
          </button>
        </div>
      </div>

      <!-- Theme -->
      <div>
        <span class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
          Theme
        </span>
        <div class="grid grid-cols-2 gap-3">
          <button
            type="button"
            class="p-3 rounded-xl border text-left flex items-center gap-3 transition {
              theme === 'dark'
                ? 'border-indigo-500 bg-indigo-950/20 text-indigo-300'
                : 'border-dark-600 bg-dark-900/60 text-slate-400 hover:text-slate-200'
            }"
            on:click={() => (theme = 'dark')}
          >
            <Moon class="w-4 h-4" />
            <span class="text-xs font-semibold">Dark Theme</span>
          </button>
          <button
            type="button"
            class="p-3 rounded-xl border text-left flex items-center gap-3 transition opacity-50 cursor-not-allowed border-dark-600 bg-dark-900/60 text-slate-400"
            title="Light theme coming soon"
          >
            <Sun class="w-4 h-4" />
            <span class="text-xs font-semibold">Light (Phase 2)</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="p-5 border-t border-dark-600/60 flex items-center justify-end gap-3">
      <button
        type="button"
        class="px-4 py-2 rounded-xl text-slate-400 hover:text-slate-200 text-xs font-medium transition"
        on:click={onClose}
      >
        Cancel
      </button>
      <button
        type="button"
        class="px-5 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-md flex items-center gap-2 transition"
        on:click={handleSave}
      >
        <Save class="w-3.5 h-3.5" />
        Save Settings
      </button>
    </div>
  </div>
</div>
