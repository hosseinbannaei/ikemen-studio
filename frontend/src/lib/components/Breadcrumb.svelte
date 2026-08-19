<script lang="ts">
  import { projectStore } from '../stores/projectStore';
  import { ArrowLeft, ChevronRight, Play, Square, Loader2 } from 'lucide-svelte';

  export let items: { label: string; onClick?: () => void }[] = [];
  export let onBack: (() => void) | null = null;
  export let backLabel = 'Back';
  export let showPlayButton = false;

  function handleLaunchToggle() {
    if ($projectStore.gameState === 'running' && $projectStore.canStop) {
      projectStore.stop();
    } else if ($projectStore.gameState === 'idle') {
      projectStore.launch();
    }
  }
</script>

<div class="h-14 bg-dark-850/80 backdrop-blur border-b border-dark-600/60 px-6 flex items-center justify-between select-none z-10 flex-shrink-0">
  <!-- Left: Back button & Breadcrumb Trail -->
  <div class="flex items-center gap-3">
    {#if onBack}
      <button
        type="button"
        class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg bg-dark-700/80 hover:bg-dark-700 text-slate-300 hover:text-white text-xs font-semibold border border-dark-600/60 transition shadow-sm"
        on:click={onBack}
      >
        <ArrowLeft class="w-3.5 h-3.5" />
        <span>{backLabel}</span>
      </button>
      <div class="h-4 w-px bg-dark-600/60"></div>
    {/if}

    <nav aria-label="Breadcrumb" class="flex items-center gap-1.5 text-xs font-medium">
      {#each items as item, index}
        {#if index > 0}
          <ChevronRight class="w-3.5 h-3.5 text-slate-600 flex-shrink-0" />
        {/if}

        {#if item.onClick && index < items.length - 1}
          <button
            type="button"
            class="text-slate-400 hover:text-indigo-300 font-semibold transition"
            on:click={item.onClick}
          >
            {item.label}
          </button>
        {:else}
          <span class="text-slate-100 font-bold">{item.label}</span>
        {/if}
      {/each}
    </nav>
  </div>

  <!-- Right: Actions / Play Button (if active in workspace) -->
  <div class="flex items-center gap-3">
    {#if showPlayButton && $projectStore.current}
      <button
        type="button"
        disabled={$projectStore.gameState === 'starting' || $projectStore.gameState === 'stopping' || ($projectStore.gameState === 'running' && !$projectStore.canStop)}
        class="flex items-center gap-2 px-4 py-1.5 rounded-xl font-bold text-xs shadow-md transition-all {
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
    {/if}

    <slot name="actions" />
  </div>
</div>
