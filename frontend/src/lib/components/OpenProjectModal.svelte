<script lang="ts">
  import { onMount } from 'svelte';
  import { projectStore } from '../stores/projectStore';
  import { X, FolderOpen, Clock, ChevronRight, HardDrive } from 'lucide-svelte';

  export let onClose: () => void;

  onMount(() => {
    projectStore.loadRecent();
  });

  async function handleBrowse() {
    const success = await projectStore.selectAndOpen();
    if (success) {
      onClose();
    }
  }

  async function handleOpenRecent(path: string) {
    const success = await projectStore.open(path);
    if (success) {
      onClose();
    }
  }
</script>

<div class="fixed inset-0 z-40 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
  <div class="bg-dark-800 border border-dark-600/80 rounded-2xl w-full max-w-lg shadow-2xl overflow-hidden">
    <!-- Header -->
    <div class="p-5 border-b border-dark-600/60 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <div class="w-9 h-9 rounded-xl bg-indigo-500/10 border border-indigo-500/30 flex items-center justify-center text-indigo-400">
          <FolderOpen class="w-5 h-5" />
        </div>
        <div>
          <h2 class="text-base font-bold text-slate-100">Open Project</h2>
          <p class="text-xs text-slate-400">Select an existing Ikemen GO project directory</p>
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
    <div class="p-5 space-y-5">
      <!-- Browse Button Card -->
      <button
        type="button"
        class="w-full p-4 rounded-xl border border-dashed border-dark-600 hover:border-indigo-500 bg-dark-900/60 hover:bg-dark-900 text-center flex flex-col items-center justify-center gap-2 group transition"
        on:click={handleBrowse}
      >
        <div class="w-10 h-10 rounded-xl bg-indigo-500/10 text-indigo-400 group-hover:bg-indigo-600 group-hover:text-white flex items-center justify-center transition">
          <FolderOpen class="w-5 h-5" />
        </div>
        <div>
          <div class="text-xs font-semibold text-slate-200 group-hover:text-white transition">Browse Filesystem</div>
          <div class="text-[11px] text-slate-500">Pick folder containing ikemen-project.json</div>
        </div>
      </button>

      <!-- Recent Projects -->
      <div>
        <div class="flex items-center gap-2 text-xs font-bold text-slate-400 uppercase tracking-wider mb-2.5">
          <Clock class="w-3.5 h-3.5" />
          <span>Recent Projects</span>
        </div>

        {#if $projectStore.recent.length === 0}
          <div class="p-6 rounded-xl bg-dark-900/40 border border-dark-600/30 text-center text-xs text-slate-500">
            No recently opened projects found.
          </div>
        {:else}
          <div class="space-y-1.5 max-h-56 overflow-y-auto">
            {#each $projectStore.recent as path}
              <button
                type="button"
                class="w-full p-3 rounded-xl bg-dark-900/60 hover:bg-dark-900 border border-dark-600/40 hover:border-indigo-500/50 flex items-center justify-between text-left transition group"
                on:click={() => handleOpenRecent(path)}
              >
                <div class="flex items-center gap-3 min-w-0">
                  <HardDrive class="w-4 h-4 text-slate-500 group-hover:text-indigo-400 flex-shrink-0 transition" />
                  <div class="min-w-0">
                    <div class="text-xs font-semibold text-slate-200 group-hover:text-white truncate">
                      {path.split(/[/\\]/).pop()}
                    </div>
                    <div class="text-[11px] font-mono text-slate-500 truncate">{path}</div>
                  </div>
                </div>
                <ChevronRight class="w-4 h-4 text-slate-600 group-hover:text-slate-300 transition" />
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>
