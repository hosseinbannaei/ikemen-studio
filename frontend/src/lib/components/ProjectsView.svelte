<script lang="ts">
  import { onMount } from 'svelte';
  import { projectStore } from '../stores/projectStore';
  import {
    Plus,
    FolderOpen,
    Search,
    Play,
    Square,
    Folder,
    HardDrive,
    Calendar,
    ChevronRight,
    ExternalLink,
    Gamepad2,
    Loader2,
  } from 'lucide-svelte';

  export let onNewProject: () => void;
  export let onOpenProjectWorkspace: () => void;

  let searchQuery = '';

  onMount(() => {
    projectStore.loadRecent();
  });

  async function handleOpenExisting() {
    const success = await projectStore.selectAndOpen();
    if (success) {
      onOpenProjectWorkspace();
    }
  }

  async function handleSelectProject(path: string) {
    const success = await projectStore.open(path);
    if (success) {
      onOpenProjectWorkspace();
    }
  }

  $: filteredProjects = $projectStore.recent.filter((p) => {
    if (!searchQuery.trim()) return true;
    const name = p.split(/[/\\]/).pop() || '';
    return name.toLowerCase().includes(searchQuery.toLowerCase()) || p.toLowerCase().includes(searchQuery.toLowerCase());
  });

  function getProjectName(p: string): string {
    return p.split(/[/\\]/).pop() || p;
  }
</script>

<div class="p-8 max-w-6xl mx-auto space-y-6">
  <!-- Top Hub Actions & Search Header -->
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
    <div>
      <h1 class="text-2xl font-black tracking-tight text-slate-100">Projects</h1>
      <p class="text-xs text-slate-400 mt-0.5">Manage, launch, and configure your Ikemen GO fighting game projects</p>
    </div>

    <div class="flex items-center gap-2.5">
      <button
        type="button"
        class="flex items-center gap-2 px-3.5 py-2 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600/80 text-slate-200 text-xs font-semibold shadow-sm transition"
        on:click={handleOpenExisting}
      >
        <FolderOpen class="w-4 h-4 text-indigo-400" />
        <span>Open</span>
      </button>

      <button
        type="button"
        class="flex items-center gap-2 px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold shadow-md shadow-indigo-600/30 transition"
        on:click={onNewProject}
      >
        <Plus class="w-4 h-4" />
        <span>New Project</span>
      </button>
    </div>
  </div>

  <!-- Search / Filter Bar -->
  {#if $projectStore.recent.length > 0}
    <div class="relative">
      <Search class="w-4 h-4 text-slate-500 absolute left-3.5 top-1/2 -translate-y-1/2" />
      <input
        type="text"
        bind:value={searchQuery}
        placeholder="Search projects by name or path..."
        class="w-full pl-10 pr-4 py-2.5 rounded-xl bg-dark-800 border border-dark-600/70 text-xs text-slate-100 placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition shadow-inner"
      />
    </div>
  {/if}

  <!-- Active Project Banner (if open) -->
  {#if $projectStore.current}
    <div class="p-5 rounded-2xl bg-gradient-to-r from-indigo-950/40 via-purple-950/30 to-dark-800 border border-indigo-500/40 flex flex-col sm:flex-row sm:items-center justify-between gap-4 shadow-lg">
      <div class="flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-indigo-500/20 text-indigo-400 border border-indigo-500/30 flex items-center justify-center flex-shrink-0">
          <Gamepad2 class="w-6 h-6" />
        </div>
        <div class="space-y-1 min-w-0">
          <div class="flex items-center gap-2 flex-wrap">
            <span class="text-base font-bold text-slate-100">{$projectStore.current.name}</span>
            <span class="text-[10px] font-mono px-2 py-0.5 rounded-full bg-purple-500/10 text-purple-300 border border-purple-500/20">
              {$projectStore.current.engine.version}
            </span>
            {#if $projectStore.gameState === 'running'}
              <span class="flex items-center gap-1.5 text-xs text-emerald-400 font-semibold">
                <span class="w-2 h-2 rounded-full bg-emerald-400 animate-ping"></span>
                Active Game
              </span>
            {/if}
          </div>
          <div class="text-xs text-slate-400 font-mono truncate max-w-lg">
            {$projectStore.current.path}
          </div>
        </div>
      </div>

      <div class="flex items-center gap-2">
        <button
          type="button"
          class="flex items-center gap-2 px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold shadow transition"
          on:click={onOpenProjectWorkspace}
        >
          <span>Open Workspace</span>
          <ChevronRight class="w-4 h-4" />
        </button>
      </div>
    </div>
  {/if}

  <!-- Projects List / Table -->
  {#if $projectStore.recent.length === 0}
    <!-- Empty State -->
    <div class="py-20 text-center flex flex-col items-center justify-center gap-4 bg-dark-800/40 rounded-2xl border border-dark-600/40 p-8">
      <div class="w-16 h-16 rounded-2xl bg-dark-700/80 text-slate-400 flex items-center justify-center border border-dark-600/60 shadow-inner">
        <HardDrive class="w-8 h-8 text-slate-500" />
      </div>
      <div class="space-y-1 max-w-sm">
        <h3 class="text-base font-bold text-slate-200">No Projects Found</h3>
        <p class="text-xs text-slate-400">
          Create a new Ikemen GO project or open an existing directory to start building.
        </p>
      </div>
      <div class="flex items-center gap-3 mt-2">
        <button
          type="button"
          class="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold shadow-md transition"
          on:click={onNewProject}
        >
          <Plus class="w-4 h-4" />
          <span>New Project</span>
        </button>
        <button
          type="button"
          class="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-dark-700 hover:bg-dark-600 text-slate-200 text-xs font-semibold border border-dark-600 transition"
          on:click={handleOpenExisting}
        >
          <FolderOpen class="w-4 h-4" />
          <span>Open Existing</span>
        </button>
      </div>
    </div>
  {:else}
    <div class="space-y-2.5">
      <div class="text-xs font-bold uppercase tracking-wider text-slate-400 px-1">
        Recent Workspaces ({filteredProjects.length})
      </div>

      <div class="grid grid-cols-1 gap-2.5">
        {#each filteredProjects as path (path)}
          {@const isActive = $projectStore.current?.path.toLowerCase() === path.toLowerCase()}
          <div class="p-4 rounded-xl bg-dark-800/80 hover:bg-dark-750 border {
            isActive ? 'border-indigo-500/60 shadow-md shadow-indigo-950/20' : 'border-dark-600/60'
          } flex flex-col md:flex-row md:items-center justify-between gap-4 transition group">
            <div class="flex items-center gap-3.5 min-w-0 flex-1">
              <div class="w-10 h-10 rounded-xl {
                isActive ? 'bg-indigo-500/20 text-indigo-400 border border-indigo-500/30' : 'bg-dark-700 text-slate-400 group-hover:text-slate-200'
              } flex items-center justify-center flex-shrink-0 transition">
                <HardDrive class="w-5 h-5" />
              </div>

              <div class="space-y-0.5 min-w-0 flex-1">
                <div class="flex items-center gap-2 flex-wrap">
                  <span class="text-sm font-bold text-slate-100 group-hover:text-white transition">
                    {getProjectName(path)}
                  </span>
                  {#if isActive}
                    <span class="text-[10px] font-mono px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 font-semibold">
                      Current
                    </span>
                  {/if}
                </div>
                <div class="text-xs font-mono text-slate-400 truncate max-w-xl">
                  {path}
                </div>
              </div>
            </div>

            <!-- Actions on Project Item -->
            <div class="flex items-center gap-2 self-end md:self-center flex-shrink-0">
              <button
                type="button"
                class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-dark-700/80 hover:bg-dark-700 text-slate-300 hover:text-white text-xs font-medium border border-dark-600/60 transition"
                on:click={() => projectStore.openFolder(path)}
                title="Open Folder in Explorer"
              >
                <Folder class="w-3.5 h-3.5 text-indigo-400" />
                <span>Explorer</span>
              </button>

              <button
                type="button"
                class="flex items-center gap-1.5 px-4 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold shadow-sm transition"
                on:click={() => handleSelectProject(path)}
              >
                <span>{isActive ? 'Workspace' : 'Open'}</span>
                <ChevronRight class="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>
