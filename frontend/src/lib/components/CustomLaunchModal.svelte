<script lang="ts">
  import { onMount } from 'svelte';
  import { projectStore } from '../stores/projectStore';
  import {
    Play,
    Terminal,
    History,
    X,
    Sparkles,
    ShieldAlert,
    Trash2,
    Sliders,
  } from 'lucide-svelte';

  export let onClose: () => void;

  let customArgs = '';
  let flagTraining = false;
  let flagDebug = false;
  let flagNoMusic = false;
  let flagFullscreen = false;

  let history: string[] = [];

  const HISTORY_KEY = 'ikemen_studio_launch_history';

  onMount(() => {
    try {
      const saved = localStorage.getItem(HISTORY_KEY);
      if (saved) {
        history = JSON.parse(saved);
      }
    } catch {}
  });

  function saveToHistory(argStr: string) {
    if (!argStr || !argStr.trim()) return;
    const trimmed = argStr.trim();
    const filtered = history.filter((h) => h !== trimmed);
    history = [trimmed, ...filtered].slice(0, 10);
    try {
      localStorage.setItem(HISTORY_KEY, JSON.stringify(history));
    } catch {}
  }

  function clearHistory() {
    history = [];
    try {
      localStorage.removeItem(HISTORY_KEY);
    } catch {}
  }

  function compileArgs(): string[] {
    const args: string[] = [];

    if (flagTraining) args.push('-training');
    if (flagDebug) {
      args.push('-log');
      args.push('-debug');
    }
    if (flagNoMusic) args.push('-nomusic');
    if (flagFullscreen) args.push('-fullscreen');

    if (customArgs.trim()) {
      // Split whitespace respecting quotes if possible or standard split
      const parts = customArgs.trim().split(/\s+/);
      args.push(...parts);
    }

    return args;
  }

  async function handleLaunch() {
    const args = compileArgs();
    if (customArgs.trim()) {
      saveToHistory(customArgs.trim());
    }
    onClose();
    await projectStore.launchWithOptions(args);
  }

  async function handleLaunchHistoryItem(item: string) {
    saveToHistory(item);
    onClose();
    const parts = item.trim().split(/\s+/);
    await projectStore.launchWithOptions(parts);
  }
</script>

<div class="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
  <div class="bg-dark-800 border border-dark-600/80 rounded-2xl w-full max-w-xl shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-150 flex flex-col max-h-[90vh]">
    <!-- Header -->
    <div class="p-5 border-b border-dark-600/60 flex items-center justify-between bg-dark-850">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-purple-500/10 border border-purple-500/20 text-purple-400 flex items-center justify-center flex-shrink-0">
          <Terminal class="w-5 h-5" />
        </div>
        <div>
          <h2 class="text-base font-bold text-slate-100">Custom Launch Options</h2>
          <p class="text-xs text-slate-400">Configure command-line arguments and test presets</p>
        </div>
      </div>
      <button
        type="button"
        class="p-2 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-dark-700/60 transition"
        on:click={onClose}
      >
        <X class="w-5 h-5" />
      </button>
    </div>

    <!-- Content -->
    <div class="p-6 space-y-5 overflow-y-auto flex-1">
      <!-- Quick Toggles -->
      <div class="space-y-2">
        <label class="text-[11px] font-bold uppercase tracking-wider text-slate-400">Quick Flag Presets</label>
        <div class="grid grid-cols-2 gap-2.5">
          <label class="flex items-center gap-3 p-3 rounded-xl bg-dark-900 border border-dark-600/60 hover:border-dark-500 cursor-pointer transition">
            <input type="checkbox" bind:checked={flagTraining} class="rounded border-dark-600 bg-dark-800 text-indigo-600 focus:ring-0 w-4 h-4" />
            <div class="text-xs">
              <span class="font-semibold text-slate-200 block">Training Mode</span>
              <span class="text-[11px] text-slate-500 font-mono">-training</span>
            </div>
          </label>

          <label class="flex items-center gap-3 p-3 rounded-xl bg-dark-900 border border-dark-600/60 hover:border-dark-500 cursor-pointer transition">
            <input type="checkbox" bind:checked={flagDebug} class="rounded border-dark-600 bg-dark-800 text-indigo-600 focus:ring-0 w-4 h-4" />
            <div class="text-xs">
              <span class="font-semibold text-slate-200 block">Developer / Debug Mode</span>
              <span class="text-[11px] text-slate-500 font-mono">-log -debug</span>
            </div>
          </label>

          <label class="flex items-center gap-3 p-3 rounded-xl bg-dark-900 border border-dark-600/60 hover:border-dark-500 cursor-pointer transition">
            <input type="checkbox" bind:checked={flagFullscreen} class="rounded border-dark-600 bg-dark-800 text-indigo-600 focus:ring-0 w-4 h-4" />
            <div class="text-xs">
              <span class="font-semibold text-slate-200 block">Force Fullscreen</span>
              <span class="text-[11px] text-slate-500 font-mono">-fullscreen</span>
            </div>
          </label>

          <label class="flex items-center gap-3 p-3 rounded-xl bg-dark-900 border border-dark-600/60 hover:border-dark-500 cursor-pointer transition">
            <input type="checkbox" bind:checked={flagNoMusic} class="rounded border-dark-600 bg-dark-800 text-indigo-600 focus:ring-0 w-4 h-4" />
            <div class="text-xs">
              <span class="font-semibold text-slate-200 block">Mute BGM</span>
              <span class="text-[11px] text-slate-500 font-mono">-nomusic</span>
            </div>
          </label>
        </div>
      </div>

      <!-- Freeform CLI Input -->
      <div class="space-y-2">
        <label class="text-[11px] font-bold uppercase tracking-wider text-slate-400">Custom CLI Arguments</label>
        <div class="relative">
          <input
            type="text"
            bind:value={customArgs}
            placeholder="-p1 kfm -p2 kfm720 -s stages/stage0.def -p1.ai 8"
            class="w-full bg-dark-900 border border-dark-600 rounded-xl px-4 py-2.5 text-xs text-slate-100 font-mono placeholder:text-slate-600 focus:outline-none focus:border-indigo-500 shadow-inner"
          />
        </div>
        <p class="text-[11px] text-slate-500">
          Examples: <code class="text-purple-300 font-mono">-p1 char1 -p2 char2</code> or <code class="text-purple-300 font-mono">-r 1920x1080</code>
        </p>
      </div>

      <!-- Launch History -->
      {#if history.length > 0}
        <div class="space-y-2 pt-2 border-t border-dark-600/40">
          <div class="flex items-center justify-between">
            <label class="text-[11px] font-bold uppercase tracking-wider text-slate-400 flex items-center gap-1.5">
              <History class="w-3.5 h-3.5 text-indigo-400" />
              <span>Recent Custom Launches</span>
            </label>
            <button
              type="button"
              class="text-[10px] text-slate-500 hover:text-rose-400 transition"
              on:click={clearHistory}
            >
              Clear History
            </button>
          </div>

          <div class="space-y-1.5 max-h-36 overflow-y-auto">
            {#each history as item}
              <div class="flex items-center justify-between p-2.5 rounded-xl bg-dark-900 border border-dark-600/40 hover:border-dark-600 transition group">
                <span class="text-xs font-mono text-slate-300 truncate max-w-sm">{item}</span>
                <button
                  type="button"
                  class="px-2.5 py-1 rounded-lg bg-dark-700 group-hover:bg-indigo-600 text-slate-300 group-hover:text-white text-[11px] font-semibold transition flex items-center gap-1"
                  on:click={() => handleLaunchHistoryItem(item)}
                >
                  <Play class="w-3 h-3 fill-current" />
                  <span>Run</span>
                </button>
              </div>
            {/each}
          </div>
        </div>
      {/if}
    </div>

    <!-- Footer -->
    <div class="p-4 border-t border-dark-600/60 bg-dark-850 flex items-center justify-between">
      <button
        type="button"
        class="px-4 py-2 rounded-xl text-slate-400 hover:text-slate-200 text-xs font-semibold hover:bg-dark-700/60 transition"
        on:click={onClose}
      >
        Cancel
      </button>

      <button
        type="button"
        class="px-6 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold shadow-md shadow-indigo-950/50 flex items-center gap-2 transition"
        on:click={handleLaunch}
      >
        <Play class="w-4 h-4 fill-current" />
        <span>Launch with Arguments</span>
      </button>
    </div>
  </div>
</div>
