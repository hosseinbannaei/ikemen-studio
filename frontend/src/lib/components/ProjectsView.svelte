<script lang="ts">
  import { projectStore } from '../stores/projectStore';
  import type { ExistingGameInspection } from '../types';
  import ConfirmModal from './ConfirmModal.svelte';
  import ImportGameModal from './ImportGameModal.svelte';
  import {
    Plus,
    FolderOpen,
    Search,
    HardDrive,
    Play,
    Clock,
    Sparkles,
    Trash2,
    FolderDown,
    ExternalLink,
  } from 'lucide-svelte';

  export let onNewProject: () => void;
  export let onOpenProjectWorkspace: () => void;

  let searchQuery = '';
  let projectToRemove: string | null = null;
  let rawGameToImport: ExistingGameInspection | null = null;

  $: filteredRecent = $projectStore.recent.filter((p) =>
    p.toLowerCase().includes(searchQuery.toLowerCase())
  );

  function getProjectName(path: string): string {
    const parts = path.replace(/\\/g, '/').split('/').filter(Boolean);
    return parts[parts.length - 1] || path;
  }

  async function handleOpenExisting() {
    const res = await projectStore.selectAndOpen();
    if (res.opened) {
      onOpenProjectWorkspace();
    } else if (res.inspection) {
      rawGameToImport = res.inspection;
    }
  }

  async function handleSelectRecent(path: string) {
    const ok = await projectStore.open(path);
    if (ok) {
      onOpenProjectWorkspace();
    }
  }

  async function confirmRemoveRecent() {
    if (projectToRemove) {
      await projectStore.removeRecent(projectToRemove);
      projectToRemove = null;
    }
  }
</script>

<div class="max-w-6xl mx-auto p-8 space-y-8">
  <!-- Top Action Bar -->
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
    <div>
      <h1 class="text-2xl font-black text-slate-100 tracking-tight">Projects Hub</h1>
      <p class="text-xs text-slate-400 mt-1">Manage, test, and develop your Ikemen GO fighting games</p>
    </div>

    <div class="flex items-center gap-3">
      <button
        type="button"
        class="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-dark-800 hover:bg-dark-700 border border-dark-600/80 text-slate-200 text-xs font-semibold shadow-sm transition"
        on:click={handleOpenExisting}
      >
        <FolderOpen class="w-4 h-4 text-indigo-400" />
        <span>Open Project</span>
      </button>

      <button
        type="button"
        class="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold shadow-lg shadow-indigo-950/50 transition"
        on:click={onNewProject}
      >
        <Plus class="w-4 h-4" />
        <span>New Project</span>
      </button>
    </div>
  </div>

  <!-- Search & Filters -->
  <div class="flex items-center gap-3">
    <div class="relative flex-1">
      <Search class="w-4 h-4 text-slate-500 absolute left-3.5 top-1/2 -translate-y-1/2" />
      <input
        type="text"
        bind:value={searchQuery}
        placeholder="Search recent projects by name or directory..."
        class="w-full bg-dark-800 border border-dark-600/70 rounded-xl pl-10 pr-4 py-2.5 text-xs text-slate-200 placeholder:text-slate-500 focus:outline-none focus:border-indigo-500 transition shadow-inner"
      />
    </div>
  </div>

  <!-- Recent Projects List -->
  <div class="space-y-3">
    <div class="flex items-center justify-between px-1">
      <h2 class="text-xs font-bold uppercase tracking-wider text-slate-400">
        Recent Workspaces ({filteredRecent.length})
      </h2>
    </div>

    {#if filteredRecent.length === 0}
      <div class="p-12 text-center rounded-2xl bg-dark-800/40 border border-dashed border-dark-600/60 flex flex-col items-center justify-center gap-3">
        <div class="w-12 h-12 rounded-2xl bg-dark-700/60 flex items-center justify-center text-slate-500">
          <HardDrive class="w-6 h-6" />
        </div>
        <div class="space-y-1">
          <div class="text-sm font-semibold text-slate-300">No Projects Found</div>
          <p class="text-xs text-slate-500 max-w-sm">
            {#if searchQuery}
              No workspaces match "{searchQuery}". Try a different search query.
            {:else}
              Get started by creating a new Ikemen GO project or opening an existing MUGEN game.
            {/if}
          </p>
        </div>
        {#if !searchQuery}
          <div class="flex items-center gap-3 mt-2">
            <button
              type="button"
              class="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold rounded-xl transition shadow"
              on:click={onNewProject}
            >
              Create New Project
            </button>
            <button
              type="button"
              class="px-4 py-2 bg-dark-700 hover:bg-dark-600 text-slate-200 text-xs font-semibold rounded-xl transition border border-dark-600"
              on:click={handleOpenExisting}
            >
              Open Folder
            </button>
          </div>
        {/if}
      </div>
    {:else}
      <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
        {#each filteredRecent as path}
          <div
            class="group p-4 rounded-2xl bg-dark-800/70 hover:bg-dark-800 border border-dark-600/60 hover:border-indigo-500/40 transition shadow-sm flex flex-col justify-between gap-4"
          >
            <div class="flex items-start justify-between gap-3">
              <div
                class="flex items-start gap-3 min-w-0 cursor-pointer flex-1"
                on:click={() => handleSelectRecent(path)}
              >
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-indigo-500/20 to-purple-500/20 text-indigo-400 border border-indigo-500/30 flex items-center justify-center flex-shrink-0 group-hover:scale-105 transition">
                  <HardDrive class="w-5 h-5" />
                </div>
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <span class="text-sm font-bold text-slate-100 group-hover:text-indigo-300 transition truncate">
                      {getProjectName(path)}
                    </span>
                    {#if $projectStore.current?.path === path}
                      <span class="text-[10px] font-bold px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
                        Active
                      </span>
                    {/if}
                  </div>
                  <p class="text-[11px] font-mono text-slate-500 truncate mt-0.5">{path}</p>
                </div>
              </div>

              <!-- Delete / Remove from Recent Button -->
              <button
                type="button"
                class="p-2 rounded-lg text-slate-500 hover:text-rose-400 hover:bg-rose-500/10 transition flex-shrink-0"
                title="Remove from recent list"
                on:click|stopPropagation={() => (projectToRemove = path)}
              >
                <Trash2 class="w-4 h-4" />
              </button>
            </div>

            <!-- Card Bottom Bar -->
            <div class="flex items-center justify-between pt-2 border-t border-dark-600/30 text-xs">
              <span class="text-[11px] text-slate-500 flex items-center gap-1">
                <Clock class="w-3.5 h-3.5" />
                <span>Workspace</span>
              </span>

              <button
                type="button"
                class="flex items-center gap-1 text-xs font-semibold text-indigo-400 hover:text-indigo-300 transition"
                on:click={() => handleSelectRecent(path)}
              >
                <span>Open Studio</span>
                <ExternalLink class="w-3 h-3" />
              </button>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<!-- Confirm Remove Recent Modal -->
{#if projectToRemove}
  <ConfirmModal
    title="Remove Project from Recent?"
    message="Are you sure you want to remove '{getProjectName(projectToRemove)}' from your recent list? Your files and directories on disk will NOT be deleted."
    confirmLabel="Remove"
    confirmVariant="danger"
    onConfirm={confirmRemoveRecent}
    onCancel={() => (projectToRemove = null)}
  />
{/if}

<!-- Import Raw Game Modal -->
{#if rawGameToImport}
  <ImportGameModal
    inspection={rawGameToImport}
    onClose={() => (rawGameToImport = null)}
    onSuccess={() => {
      rawGameToImport = null;
      onOpenProjectWorkspace();
    }}
  />
{/if}
