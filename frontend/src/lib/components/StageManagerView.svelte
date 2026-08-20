<script lang="ts">
  import { onMount } from 'svelte';
  import { projectStore } from '../stores/projectStore';
  import { vaultStore } from '../stores/vaultStore';
  import { toastStore } from '../stores/toastStore';
  import type { ProjectStageInfo } from '../types';
  import AddFromVaultModal from './AddFromVaultModal.svelte';
  import ConfirmModal from './ConfirmModal.svelte';
  import {
    Mountain,
    Search,
    Plus,
    Sparkles,
    Play,
    Check,
    FolderOpen,
    Trash2,
    Music,
    ArrowLeft,
    Sliders,
    Loader2,
    Grid,
    List,
    Layers,
    User,
    CheckCircle2,
    ExternalLink,
    Maximize2,
  } from 'lucide-svelte';

  export let onBackToWorkspace: () => void;

  let stages: ProjectStageInfo[] = [];
  let loading = true;
  let searchQuery = '';
  let filterMode: 'all' | 'extra' | 'assigned' = 'all';
  let viewMode: 'grid' | 'table' = 'grid';

  let showAddFromVaultModal = false;
  let stageToDelete: ProjectStageInfo | null = null;
  let selectedStageForDetail: ProjectStageInfo | null = null;

  onMount(async () => {
    await loadStages();
  });

  async function loadStages() {
    loading = true;
    stages = await projectStore.loadStages();
    loading = false;
  }

  $: filteredStages = stages.filter((st) => {
    const q = searchQuery.toLowerCase();
    const matchesSearch =
      st.display_name.toLowerCase().includes(q) ||
      st.author.toLowerCase().includes(q) ||
      st.relative_path.toLowerCase().includes(q);

    if (!matchesSearch) return false;

    if (filterMode === 'extra') return st.is_extra_stage;
    if (filterMode === 'assigned') return st.assigned_characters.length > 0;
    return true;
  });

  async function handleToggleExtra(stage: ProjectStageInfo) {
    const nextState = !stage.is_extra_stage;
    const ok = await projectStore.toggleStageExtra(stage.relative_path, nextState);
    if (ok) {
      stage.is_extra_stage = nextState;
      stages = [...stages];
    }
  }

  async function handleTestStage(stage: ProjectStageInfo) {
    const data = await projectStore.getFightersAndStages();
    const char = data.characters.length > 0 ? data.characters[0] : 'kfm';
    await projectStore.launchWithOptions(['-s', stage.relative_path, '-p1', char, '-p2', char, '-p2.ai', '0', '-time', '-1']);
  }

  async function confirmDelete() {
    if (!stageToDelete) return;
    const target = stageToDelete;
    stageToDelete = null;
    const ok = await projectStore.deleteStage(target.relative_path);
    if (ok) {
      await loadStages();
    }
  }

  async function handleBrowseArchive() {
    try {
      const archives = await (window as any).go.main.App.SelectMultipleArchivesDialog();
      if (archives && archives.length > 0) {
        if ($projectStore.current?.path) {
          loading = true;
          // Ingest into vault and link, or copy into stages
          for (const arch of archives) {
            await vaultStore.ingestAsset(arch, 'stages');
          }
          await loadStages();
          toastStore.success('Stages Ingested', `Processed ${archives.length} archive(s)`);
        }
      }
    } catch (err: any) {
      toastStore.error('Import Failed', err?.message || 'Could not import stage archive');
    } finally {
      loading = false;
    }
  }
</script>

<div class="h-full flex flex-col min-w-0 bg-dark-900 select-none overflow-hidden">
  <!-- Top Bar -->
  <header class="p-6 border-b border-dark-600/60 bg-dark-850/60 flex flex-col md:flex-row md:items-center justify-between gap-4 flex-shrink-0">
    <div class="flex items-center gap-3">
      <button
        type="button"
        class="p-2 rounded-xl bg-dark-700 hover:bg-dark-600 text-slate-300 hover:text-white transition shadow-sm"
        on:click={onBackToWorkspace}
        title="Back to Workspace"
      >
        <ArrowLeft class="w-4 h-4" />
      </button>

      <div>
        <div class="flex items-center gap-2">
          <h1 class="text-xl font-black text-slate-100 flex items-center gap-2">
            <Mountain class="w-5 h-5 text-emerald-400" />
            <span>Stages Manager</span>
          </h1>
          <span class="text-xs px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-300 border border-emerald-500/20 font-mono">
            {stages.length} Installed
          </span>
        </div>
        <p class="text-xs text-slate-400 mt-0.5">Manage battle arenas, ExtraStages select screen registration, BGM, and camera bounds</p>
      </div>
    </div>

    <!-- Actions -->
    <div class="flex items-center gap-2.5 flex-wrap">
      <button
        type="button"
        class="px-3.5 py-2 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600/70 text-slate-200 text-xs font-semibold flex items-center gap-2 transition shadow-sm"
        on:click={() => projectStore.openFolder('stages')}
        title="Open stages/ folder"
      >
        <FolderOpen class="w-4 h-4 text-indigo-400" />
        <span>Open Folder</span>
      </button>

      <button
        type="button"
        class="px-3.5 py-2 rounded-xl bg-brand-600/20 hover:bg-brand-600/30 border border-brand-500/40 text-brand-300 text-xs font-semibold flex items-center gap-2 transition shadow-sm"
        on:click={() => (showAddFromVaultModal = true)}
        title="Link stages from Asset Vault"
      >
        <Sparkles class="w-4 h-4 text-brand-400" />
        <span>+ From Vault</span>
      </button>

      <button
        type="button"
        class="px-4 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-bold flex items-center gap-2 transition shadow-md shadow-emerald-950/40"
        on:click={handleBrowseArchive}
        title="Import stage archives (.zip, .rar, .7z)"
      >
        <Plus class="w-4 h-4" />
        <span>Ingest Archives</span>
      </button>
    </div>
  </header>

  <!-- Filter & Search Controls -->
  <div class="p-4 px-6 border-b border-dark-600/40 bg-dark-900 flex flex-col sm:flex-row sm:items-center justify-between gap-3 flex-shrink-0">
    <div class="flex items-center gap-2 flex-1 max-w-md">
      <div class="relative flex-1">
        <Search class="w-4 h-4 text-slate-500 absolute left-3 top-1/2 -translate-y-1/2" />
        <input
          type="text"
          bind:value={searchQuery}
          placeholder="Search stages by name, author, or file..."
          class="w-full bg-dark-800 border border-dark-600/70 rounded-xl pl-9 pr-3 py-1.5 text-xs text-slate-200 placeholder:text-slate-500 focus:outline-none focus:border-emerald-500 transition"
        />
      </div>
    </div>

    <div class="flex items-center gap-3">
      <!-- Filter Tabs -->
      <div class="flex items-center bg-dark-800 p-1 rounded-xl border border-dark-600/60 text-xs font-medium">
        <button
          type="button"
          class="px-3 py-1 rounded-lg transition {filterMode === 'all' ? 'bg-emerald-600 text-white shadow-sm' : 'text-slate-400 hover:text-slate-200'}"
          on:click={() => (filterMode = 'all')}
        >
          All ({stages.length})
        </button>
        <button
          type="button"
          class="px-3 py-1 rounded-lg transition {filterMode === 'extra' ? 'bg-emerald-600 text-white shadow-sm' : 'text-slate-400 hover:text-slate-200'}"
          on:click={() => (filterMode = 'extra')}
        >
          Extra Stages ({stages.filter((s) => s.is_extra_stage).length})
        </button>
        <button
          type="button"
          class="px-3 py-1 rounded-lg transition {filterMode === 'assigned' ? 'bg-emerald-600 text-white shadow-sm' : 'text-slate-400 hover:text-slate-200'}"
          on:click={() => (filterMode = 'assigned')}
        >
          Home Stages ({stages.filter((s) => s.assigned_characters.length > 0).length})
        </button>
      </div>

      <!-- View Switcher -->
      <div class="flex items-center bg-dark-800 p-1 rounded-xl border border-dark-600/60 text-xs">
        <button
          type="button"
          class="p-1.5 rounded-lg transition {viewMode === 'grid' ? 'bg-dark-700 text-emerald-400' : 'text-slate-400 hover:text-slate-200'}"
          on:click={() => (viewMode = 'grid')}
          title="Grid View"
        >
          <Grid class="w-4 h-4" />
        </button>
        <button
          type="button"
          class="p-1.5 rounded-lg transition {viewMode === 'table' ? 'bg-dark-700 text-emerald-400' : 'text-slate-400 hover:text-slate-200'}"
          on:click={() => (viewMode = 'table')}
          title="List View"
        >
          <List class="w-4 h-4" />
        </button>
      </div>
    </div>
  </div>

  <!-- Main Content Area -->
  <main class="flex-1 overflow-y-auto p-6">
    {#if loading}
      <div class="h-full flex flex-col items-center justify-center gap-3 text-slate-400">
        <Loader2 class="w-8 h-8 animate-spin text-emerald-400" />
        <span class="text-xs">Scanning stage definitions and assets...</span>
      </div>
    {:else if filteredStages.length === 0}
      <div class="h-full flex flex-col items-center justify-center p-12 text-center text-slate-500 border-2 border-dashed border-dark-700/60 rounded-3xl">
        <Mountain class="w-16 h-16 stroke-1 opacity-40 mb-3 text-emerald-400" />
        <h3 class="text-base font-bold text-slate-300">No Stages Found</h3>
        <p class="text-xs text-slate-500 max-w-sm mt-1.5">
          {#if searchQuery}
            No stages match "{searchQuery}".
          {:else}
            Drop stage archives (.zip, .rar, .7z) or link stages from your Asset Vault library to get started.
          {/if}
        </p>
        <div class="flex items-center gap-3 mt-4">
          <button
            type="button"
            class="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-bold rounded-xl transition shadow"
            on:click={handleBrowseArchive}
          >
            Import Stage Archive
          </button>
          <button
            type="button"
            class="px-4 py-2 bg-dark-700 hover:bg-dark-600 text-slate-200 text-xs font-semibold rounded-xl border border-dark-600 transition"
            on:click={() => (showAddFromVaultModal = true)}
          >
            Browse Vault
          </button>
        </div>
      </div>
    {:else if viewMode === 'grid'}
      <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        {#each filteredStages as stage (stage.relative_path)}
          <div
            class="p-4 rounded-2xl bg-dark-800 border border-dark-600/70 hover:border-emerald-500/50 transition flex flex-col justify-between group shadow-sm"
          >
            <div class="space-y-3">
              <!-- Thumbnail Banner -->
              <div class="w-full h-32 rounded-xl bg-dark-900 border border-dark-700/70 overflow-hidden flex items-center justify-center relative group-hover:border-emerald-500/30 transition">
                {#if stage.preview_base64}
                  <img src={stage.preview_base64} alt={stage.display_name} class="w-full h-full object-cover" />
                {:else}
                  <div class="flex flex-col items-center gap-1.5 text-slate-600">
                    <Mountain class="w-8 h-8" />
                    <span class="text-[10px] font-mono">No Preview</span>
                  </div>
                {/if}

                <!-- Extra Stage Badge -->
                {#if stage.is_extra_stage}
                  <span class="absolute top-2 right-2 px-2 py-0.5 rounded-full bg-emerald-500/90 text-slate-900 text-[10px] font-black uppercase tracking-wider shadow">
                    Extra Stage
                  </span>
                {/if}

                {#if stage.is_linked_from_vault}
                  <span class="absolute top-2 left-2 px-1.5 py-0.5 rounded-md bg-purple-500/80 text-white text-[9px] font-bold tracking-wider">
                    Vault Link
                  </span>
                {/if}
              </div>

              <!-- Meta -->
              <div>
                <div class="flex items-start justify-between gap-2">
                  <h3 class="text-sm font-bold text-slate-100 group-hover:text-emerald-300 transition truncate">
                    {stage.display_name}
                  </h3>
                </div>
                <div class="flex items-center gap-2 text-[11px] text-slate-400 mt-1">
                  <span class="truncate">By {stage.author}</span>
                  <span class="text-slate-600">&bull;</span>
                  <span class="font-mono text-slate-500 truncate">{stage.relative_path}</span>
                </div>
              </div>

              <!-- Sound & Zoom Info -->
              <div class="space-y-1 pt-1 text-[11px]">
                {#if stage.bgm_path}
                  <div class="flex items-center gap-1.5 text-slate-400 truncate">
                    <Music class="w-3 h-3 text-rose-400 flex-shrink-0" />
                    <span class="truncate">{stage.bgm_path}</span>
                  </div>
                {/if}
                {#if stage.assigned_characters && stage.assigned_characters.length > 0}
                  <div class="flex items-center gap-1.5 text-indigo-300 truncate">
                    <User class="w-3 h-3 text-indigo-400 flex-shrink-0" />
                    <span class="truncate">Home: {stage.assigned_characters.join(', ')}</span>
                  </div>
                {/if}
              </div>
            </div>

            <!-- Card Bottom Action Row -->
            <div class="pt-3 mt-3 border-t border-dark-600/40 flex items-center justify-between gap-2">
              <button
                type="button"
                class="px-2.5 py-1.5 rounded-lg bg-emerald-600/20 hover:bg-emerald-600/30 text-emerald-300 border border-emerald-500/30 text-xs font-semibold flex items-center gap-1.5 transition"
                on:click={() => handleTestStage(stage)}
                title="Test stage in sparring mode"
              >
                <Play class="w-3.5 h-3.5 fill-current" />
                <span>Test</span>
              </button>

              <div class="flex items-center gap-1.5">
                <button
                  type="button"
                  class="px-2.5 py-1.5 rounded-lg border text-xs font-semibold transition {stage.is_extra_stage ? 'bg-emerald-500/20 text-emerald-300 border-emerald-500/40' : 'bg-dark-700 text-slate-400 border-dark-600 hover:text-slate-200'}"
                  on:click={() => handleToggleExtra(stage)}
                  title="Toggle inclusion in select.def [ExtraStages]"
                >
                  {stage.is_extra_stage ? 'Extra: ON' : 'Extra: OFF'}
                </button>

                <button
                  type="button"
                  class="p-1.5 rounded-lg text-slate-500 hover:text-rose-400 hover:bg-rose-500/10 transition"
                  title="Delete Stage"
                  on:click={() => (stageToDelete = stage)}
                >
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          </div>
        {/each}
      </div>
    {:else}
      <!-- Table View -->
      <div class="bg-dark-850 border border-dark-700/80 rounded-2xl overflow-hidden shadow-sm">
        <table class="w-full text-left text-xs border-collapse">
          <thead>
            <tr class="bg-dark-900/80 border-b border-dark-700/80 text-[11px] font-bold uppercase tracking-wider text-slate-400">
              <th class="py-3 px-4 w-16">Preview</th>
              <th class="py-3 px-4">Stage Name</th>
              <th class="py-3 px-4">File Path</th>
              <th class="py-3 px-4">Author</th>
              <th class="py-3 px-4">Music</th>
              <th class="py-3 px-4">Extra Stage</th>
              <th class="py-3 px-4 text-right">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-dark-700/50">
            {#each filteredStages as stage (stage.relative_path)}
              <tr class="hover:bg-dark-800/80 transition group">
                <td class="py-2.5 px-4">
                  <div class="w-12 h-8 rounded-lg bg-dark-900 border border-dark-700 overflow-hidden flex items-center justify-center">
                    {#if stage.preview_base64}
                      <img src={stage.preview_base64} alt={stage.display_name} class="w-full h-full object-cover" />
                    {:else}
                      <Mountain class="w-4 h-4 text-slate-600" />
                    {/if}
                  </div>
                </td>
                <td class="py-2.5 px-4 font-bold text-slate-100 group-hover:text-emerald-300 transition">
                  {stage.display_name}
                </td>
                <td class="py-2.5 px-4 font-mono text-slate-400">
                  {stage.relative_path}
                </td>
                <td class="py-2.5 px-4 text-slate-300">
                  {stage.author}
                </td>
                <td class="py-2.5 px-4 text-slate-400 truncate max-w-xs">
                  {stage.bgm_path || '—'}
                </td>
                <td class="py-2.5 px-4">
                  <button
                    type="button"
                    class="px-2 py-0.5 rounded-full text-[10px] font-bold border transition {stage.is_extra_stage ? 'bg-emerald-500/20 text-emerald-300 border-emerald-500/40' : 'bg-dark-900 text-slate-500 border-dark-700'}"
                    on:click={() => handleToggleExtra(stage)}
                  >
                    {stage.is_extra_stage ? 'Registered' : 'Disabled'}
                  </button>
                </td>
                <td class="py-2.5 px-4 text-right">
                  <div class="flex items-center justify-end gap-1.5">
                    <button
                      type="button"
                      class="px-2.5 py-1 text-[11px] font-semibold text-emerald-300 hover:bg-emerald-600/20 rounded-lg transition"
                      on:click={() => handleTestStage(stage)}
                    >
                      Play
                    </button>
                    <button
                      type="button"
                      class="p-1 text-slate-500 hover:text-rose-400 hover:bg-rose-500/10 rounded-lg transition"
                      on:click={() => (stageToDelete = stage)}
                    >
                      <Trash2 class="w-3.5 h-3.5" />
                    </button>
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </main>
</div>

<!-- Confirm Delete Modal -->
{#if stageToDelete}
  <ConfirmModal
    title="Delete Stage?"
    message="Are you sure you want to remove '{stageToDelete.display_name}'? The stage file and graphics will be deleted and unlinked from select.def."
    confirmLabel="Delete Stage"
    confirmVariant="danger"
    onConfirm={confirmDelete}
    onCancel={() => (stageToDelete = null)}
  />
{/if}

<!-- Add From Vault Modal -->
{#if showAddFromVaultModal}
  <AddFromVaultModal
    isOpen={showAddFromVaultModal}
    projectDir={$projectStore.current?.path || ''}
    targetCategory="stages"
    onClose={() => {
      showAddFromVaultModal = false;
      loadStages();
    }}
  />
{/if}
