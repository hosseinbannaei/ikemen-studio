<script lang="ts">
  import { projectStore } from '../stores/projectStore';
  import { engineStore } from '../stores/engineStore';
  import {
    FolderKanban,
    Layers,
    Settings,
    Gamepad2,
    Play,
    Square,
    Loader2,
    ChevronRight,
    Sparkles,
  } from 'lucide-svelte';

  export let activeTab: 'projects' | 'engines' | 'settings';
  export let onSelectTab: (tab: 'projects' | 'engines' | 'settings') => void;
  export let onOpenProjectWorkspace: () => void;

  $: isRunning = $projectStore.gameState === 'running';
  $: isStarting = $projectStore.gameState === 'starting';
  $: isStopping = $projectStore.gameState === 'stopping';
</script>

<aside class="w-60 bg-dark-850 border-r border-dark-600/60 flex flex-col justify-between select-none flex-shrink-0 z-20">
  <!-- Brand / Hub Header -->
  <div class="p-4 border-b border-dark-600/40">
    <div class="flex items-center gap-3">
      <div class="w-9 h-9 rounded-xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center shadow-lg shadow-indigo-500/20 text-white flex-shrink-0">
        <Gamepad2 class="w-5 h-5" />
      </div>
      <div class="min-w-0">
        <div class="text-sm font-bold text-slate-100 flex items-center gap-1.5">
          <span class="truncate">Ikemen Studio</span>
        </div>
        <div class="text-[10px] font-mono text-slate-400">Hub &bull; Phase 1</div>
      </div>
    </div>
  </div>

  <!-- Navigation Links -->
  <div class="p-3 space-y-1 flex-1 overflow-y-auto">
    <!-- Projects Tab -->
    <button
      type="button"
      class="w-full flex items-center justify-between px-3.5 py-2.5 rounded-xl text-xs font-semibold transition-all {
        activeTab === 'projects'
          ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/20'
          : 'text-slate-400 hover:text-slate-200 hover:bg-dark-700/60'
      }"
      on:click={() => onSelectTab('projects')}
    >
      <div class="flex items-center gap-2.5">
        <FolderKanban class="w-4 h-4 {activeTab === 'projects' ? 'text-white' : 'text-indigo-400'}" />
        <span>Projects</span>
      </div>
      {#if isRunning}
        <span class="w-2 h-2 rounded-full bg-emerald-400 animate-ping" title="Game Running"></span>
      {:else if $projectStore.current}
        <span class="text-[10px] font-mono opacity-80 px-1.5 py-0.2 rounded bg-black/20">Active</span>
      {/if}
    </button>

    <!-- Engines Tab -->
    <button
      type="button"
      class="w-full flex items-center justify-between px-3.5 py-2.5 rounded-xl text-xs font-semibold transition-all {
        activeTab === 'engines'
          ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/20'
          : 'text-slate-400 hover:text-slate-200 hover:bg-dark-700/60'
      }"
      on:click={() => onSelectTab('engines')}
    >
      <div class="flex items-center gap-2.5">
        <Layers class="w-4 h-4 {activeTab === 'engines' ? 'text-white' : 'text-purple-400'}" />
        <span>Engines</span>
      </div>
      <span class="text-[10px] font-mono px-1.5 py-0.5 rounded-full {
        activeTab === 'engines' ? 'bg-indigo-700 text-white' : 'bg-dark-700 text-slate-400'
      }">
        {$engineStore.installed.length}
      </span>
    </button>

    <!-- Settings Tab -->
    <button
      type="button"
      class="w-full flex items-center justify-between px-3.5 py-2.5 rounded-xl text-xs font-semibold transition-all {
        activeTab === 'settings'
          ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/20'
          : 'text-slate-400 hover:text-slate-200 hover:bg-dark-700/60'
      }"
      on:click={() => onSelectTab('settings')}
    >
      <div class="flex items-center gap-2.5">
        <Settings class="w-4 h-4 {activeTab === 'settings' ? 'text-white' : 'text-cyan-400'}" />
        <span>Settings</span>
      </div>
    </button>
  </div>

  <!-- Bottom: Active Project Widget (if loaded) & App Info -->
  <div class="p-3 border-t border-dark-600/40 space-y-2 bg-dark-900/40">
    {#if $projectStore.current}
      <button
        type="button"
        class="w-full p-2.5 rounded-xl bg-dark-800 border border-dark-600/60 hover:border-indigo-500/50 flex flex-col gap-1.5 text-left transition group shadow-sm"
        on:click={onOpenProjectWorkspace}
      >
        <div class="flex items-center justify-between w-full">
          <span class="text-[10px] uppercase font-bold tracking-wider text-slate-400">Current Project</span>
          {#if isRunning}
            <span class="flex items-center gap-1 text-[10px] text-emerald-400 font-semibold">
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-ping"></span>
              Live
            </span>
          {:else if isStarting}
            <span class="flex items-center gap-1 text-[10px] text-amber-400 font-semibold">
              <Loader2 class="w-2.5 h-2.5 animate-spin" />
              Starting
            </span>
          {/if}
        </div>
        <div class="flex items-center justify-between w-full">
          <span class="text-xs font-bold text-slate-100 truncate group-hover:text-indigo-300 transition">
            {$projectStore.current.name}
          </span>
          <ChevronRight class="w-3.5 h-3.5 text-slate-500 group-hover:text-slate-300 transition flex-shrink-0" />
        </div>
        <span class="text-[10px] font-mono text-purple-300/80 truncate">
          {$projectStore.current.engine.version}
        </span>
      </button>
    {/if}

    <div class="px-2 pt-1 flex items-center justify-between text-[11px] text-slate-500 font-mono">
      <span>Ikemen GO Studio</span>
      <span>v0.1.0</span>
    </div>
  </div>
</aside>
