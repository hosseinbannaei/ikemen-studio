<script lang="ts">
  import { projectStore } from '../stores/projectStore';
  import { Play, Square, Settings, Layers, FolderOpen, Plus, Gamepad2, Loader2 } from 'lucide-svelte';

  export let onOpenEngines: () => void;
  export let onOpenSettings: () => void;
  export let onOpenNewProject: () => void;
  export let onOpenExistingProject: () => void;

  function handleLaunchToggle() {
    if ($projectStore.gameState === 'running' && $projectStore.canStop) {
      projectStore.stop();
    } else if ($projectStore.gameState === 'idle') {
      projectStore.launch();
    }
  }

  $: isBusy = $projectStore.gameState !== 'idle';
</script>

<header class="h-14 bg-dark-800 border-b border-dark-600/60 px-4 flex items-center justify-between select-none z-10 flex-shrink-0">
  <!-- Left: Branding -->
  <div class="flex items-center gap-3">
    <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center shadow-md">
      <Gamepad2 class="w-5 h-5 text-white" />
    </div>
    <div>
      <div class="text-sm font-bold tracking-wide text-slate-100 flex items-center gap-2">
        Ikemen GO Studio
        <span class="text-[10px] uppercase font-mono px-1.5 py-0.5 rounded bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">Phase 1</span>
      </div>
    </div>
  </div>

  <!-- Center: Active Project Status & Launch Button -->
  <div class="flex items-center gap-3">
    {#if $projectStore.current}
      <div class="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-dark-900/80 border border-dark-600/40 text-xs">
        <span class="font-medium text-slate-200">{$projectStore.current.name}</span>
        <span class="text-slate-500">•</span>
        <span class="font-mono text-[11px] px-1.5 py-0.5 rounded bg-purple-900/40 text-purple-300 border border-purple-700/30">
          {$projectStore.current.engine.version}
        </span>
        {#if $projectStore.gameState === 'running'}
          <span class="flex items-center gap-1.5 text-emerald-400 font-medium pl-1">
            <span class="w-2 h-2 rounded-full bg-emerald-400 animate-ping"></span>
            Running
          </span>
        {:else if $projectStore.gameState === 'starting'}
          <span class="flex items-center gap-1.5 text-amber-400 font-medium pl-1">
            <Loader2 class="w-3 h-3 animate-spin" />
            Starting...
          </span>
        {:else if $projectStore.gameState === 'stopping'}
          <span class="flex items-center gap-1.5 text-rose-400 font-medium pl-1">
            <Loader2 class="w-3 h-3 animate-spin" />
            Stopping...
          </span>
        {/if}
      </div>

      <!-- Quick Play/Stop Button with Cooldown & Spinner States -->
      <button
        type="button"
        disabled={$projectStore.gameState === 'starting' || $projectStore.gameState === 'stopping' || ($projectStore.gameState === 'running' && !$projectStore.canStop)}
        class="flex items-center gap-2 px-4 py-1.5 rounded-lg font-medium text-xs shadow-sm transition-all {
          $projectStore.gameState === 'starting' || $projectStore.gameState === 'stopping'
            ? 'bg-dark-700 text-slate-400 cursor-not-allowed border border-dark-600'
            : $projectStore.gameState === 'running'
              ? $projectStore.canStop
                ? 'bg-rose-600 hover:bg-rose-500 text-white shadow-rose-950/40'
                : 'bg-rose-800 text-rose-200 cursor-wait opacity-80'
              : 'bg-emerald-600 hover:bg-emerald-500 text-white shadow-emerald-950/40'
        }"
        on:click={handleLaunchToggle}
      >
        {#if $projectStore.gameState === 'starting'}
          <Loader2 class="w-3.5 h-3.5 animate-spin" />
          <span>Starting...</span>
        {:else if $projectStore.gameState === 'stopping'}
          <Loader2 class="w-3.5 h-3.5 animate-spin" />
          <span>Stopping...</span>
        {:else if $projectStore.gameState === 'running'}
          <Square class="w-3.5 h-3.5 fill-current" />
          <span>{$projectStore.canStop ? 'Stop Game' : 'Running...'}</span>
        {:else}
          <Play class="w-3.5 h-3.5 fill-current" />
          <span>Play</span>
        {/if}
      </button>
    {:else}
      <div class="flex items-center gap-2">
        <button
          type="button"
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-indigo-600/90 hover:bg-indigo-500 text-white text-xs font-medium transition"
          on:click={onOpenNewProject}
        >
          <Plus class="w-3.5 h-3.5" />
          New Project
        </button>
        <button
          type="button"
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-dark-700 hover:bg-dark-600 text-slate-200 text-xs font-medium border border-dark-600/60 transition"
          on:click={onOpenExistingProject}
        >
          <FolderOpen class="w-3.5 h-3.5" />
          Open Project
        </button>
      </div>
    {/if}
  </div>

  <!-- Right: Global Actions -->
  <div class="flex items-center gap-1.5">
    {#if $projectStore.current}
      <button
        type="button"
        title={isBusy ? 'Cannot switch projects while game is running' : 'Open Another Project'}
        disabled={isBusy}
        class="p-2 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-dark-700/80 disabled:opacity-40 disabled:hover:bg-transparent transition"
        on:click={onOpenExistingProject}
      >
        <FolderOpen class="w-4 h-4" />
      </button>
      <button
        type="button"
        title={isBusy ? 'Cannot create projects while game is running' : 'Create New Project'}
        disabled={isBusy}
        class="p-2 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-dark-700/80 disabled:opacity-40 disabled:hover:bg-transparent transition"
        on:click={onOpenNewProject}
      >
        <Plus class="w-4 h-4" />
      </button>
    {/if}

    <button
      type="button"
      title="Engine Manager"
      class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-slate-300 hover:text-white hover:bg-dark-700/80 text-xs font-medium transition"
      on:click={onOpenEngines}
    >
      <Layers class="w-4 h-4 text-indigo-400" />
      <span class="hidden sm:inline">Engines</span>
    </button>

    <button
      type="button"
      title="Settings"
      class="p-2 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-dark-700/80 transition"
      on:click={onOpenSettings}
    >
      <Settings class="w-4 h-4" />
    </button>
  </div>
</header>
