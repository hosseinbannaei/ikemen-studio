<script lang="ts">
  import { onMount } from 'svelte';
  import { projectStore } from '../stores/projectStore';
  import { vaultStore } from '../stores/vaultStore';
  import { toastStore } from '../stores/toastStore';
  import type { ProjectStoryboardInfo } from '../types';
  import AddFromVaultModal from './AddFromVaultModal.svelte';
  import {
    Film,
    ArrowLeft,
    Sparkles,
    Plus,
    FolderOpen,
    Play,
    Check,
    Loader2,
    Music,
    Layers,
    Tag,
  } from 'lucide-svelte';

  export let onBackToWorkspace: () => void;

  let storyboards: ProjectStoryboardInfo[] = [];
  let loading = true;
  let showAddFromVaultModal = false;

  onMount(async () => {
    await loadStoryboards();
  });

  async function loadStoryboards() {
    loading = true;
    storyboards = await projectStore.loadStoryboards();
    loading = false;
  }

  async function handleAssignStoryboard(storyType: 'intro' | 'ending' | 'credits', sb: ProjectStoryboardInfo) {
    const ok = await projectStore.setSystemStoryboard(storyType, sb.relative_path);
    if (ok) {
      await loadStoryboards();
    }
  }

  async function handleBrowseArchive() {
    try {
      const archives = await (window as any).go.main.App.SelectMultipleArchivesDialog();
      if (archives && archives.length > 0) {
        loading = true;
        for (const arch of archives) {
          await vaultStore.ingestAsset(arch, 'storyboards');
        }
        await loadStoryboards();
        toastStore.success('Storyboards Ingested', `Processed ${archives.length} archive(s)`);
      }
    } catch (err: any) {
      toastStore.error('Import Failed', err?.message || 'Could not import storyboard');
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
            <Film class="w-5 h-5 text-amber-400" />
            <span>Storyboards & Cutscenes</span>
          </h1>
          <span class="text-xs px-2 py-0.5 rounded-full bg-amber-500/10 text-amber-300 border border-amber-500/20 font-mono">
            {storyboards.length} Sequences
          </span>
        </div>
        <p class="text-xs text-slate-400 mt-0.5">Manage game intros, story endings, and credits sequences mapped in system.def</p>
      </div>
    </div>

    <!-- Actions -->
    <div class="flex items-center gap-2.5 flex-wrap">
      <button
        type="button"
        class="px-3.5 py-2 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600/70 text-slate-200 text-xs font-semibold flex items-center gap-2 transition shadow-sm"
        on:click={() => projectStore.openFolder('data')}
        title="Open data/ folder"
      >
        <FolderOpen class="w-4 h-4 text-indigo-400" />
        <span>Open data/</span>
      </button>

      <button
        type="button"
        class="px-3.5 py-2 rounded-xl bg-brand-600/20 hover:bg-brand-600/30 border border-brand-500/40 text-brand-300 text-xs font-semibold flex items-center gap-2 transition shadow-sm"
        on:click={() => (showAddFromVaultModal = true)}
        title="Link storyboards from Asset Vault"
      >
        <Sparkles class="w-4 h-4 text-brand-400" />
        <span>+ From Vault</span>
      </button>

      <button
        type="button"
        class="px-4 py-2 rounded-xl bg-amber-600 hover:bg-amber-500 text-white text-xs font-bold flex items-center gap-2 transition shadow-md shadow-amber-950/40"
        on:click={handleBrowseArchive}
        title="Import storyboard archive (.zip, .rar, .7z)"
      >
        <Plus class="w-4 h-4" />
        <span>Ingest Storyboard</span>
      </button>
    </div>
  </header>

  <!-- Main Content Area -->
  <main class="flex-1 overflow-y-auto p-6">
    {#if loading}
      <div class="h-full flex flex-col items-center justify-center gap-3 text-slate-400">
        <Loader2 class="w-8 h-8 animate-spin text-amber-400" />
        <span class="text-xs">Scanning cinematic storyboards and scene definitions...</span>
      </div>
    {:else if storyboards.length === 0}
      <div class="h-full flex flex-col items-center justify-center p-12 text-center text-slate-500 border-2 border-dashed border-dark-700/60 rounded-3xl">
        <Film class="w-16 h-16 stroke-1 opacity-40 mb-3 text-amber-400" />
        <h3 class="text-base font-bold text-slate-300">No Storyboards Found</h3>
        <p class="text-xs text-slate-500 max-w-sm mt-1.5">
          Add intro, ending, or credits cutscenes (.def files with [SceneDef]) to data/ or link from Vault.
        </p>
      </div>
    {:else}
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {#each storyboards as sb (sb.relative_path)}
          <div
            class="p-5 rounded-2xl bg-dark-800 border border-dark-600/70 hover:border-amber-500/50 transition-all flex flex-col justify-between group shadow-sm"
          >
            <div class="space-y-4">
              <!-- Banner Preview -->
              <div class="w-full h-32 rounded-xl bg-dark-900 border border-dark-700/80 overflow-hidden flex items-center justify-center relative">
                {#if sb.preview_base64}
                  <img src={sb.preview_base64} alt={sb.display_name} class="w-full h-full object-cover" />
                {:else}
                  <div class="flex flex-col items-center gap-2 text-slate-600">
                    <Film class="w-8 h-8 text-amber-400/50" />
                    <span class="text-xs font-mono">{sb.scene_count} Scenes</span>
                  </div>
                {/if}
              </div>

              <!-- Title & Meta -->
              <div>
                <h3 class="text-sm font-bold text-slate-100 group-hover:text-amber-300 transition truncate">
                  {sb.display_name}
                </h3>
                <div class="flex items-center gap-2 text-xs text-slate-500 mt-1">
                  <span class="font-mono text-[11px] truncate">{sb.relative_path}</span>
                </div>
              </div>

              <!-- Specs & Assignments -->
              <div class="space-y-2 pt-1 text-xs">
                {#if sb.bgm_path}
                  <div class="flex items-center gap-1.5 text-slate-400 truncate">
                    <Music class="w-3.5 h-3.5 text-rose-400 flex-shrink-0" />
                    <span class="truncate">{sb.bgm_path}</span>
                  </div>
                {/if}

                <div class="flex flex-wrap gap-1.5">
                  {#if sb.assigned_slots.length > 0}
                    {#each sb.assigned_slots as slot}
                      <span class="text-[10px] font-bold px-2 py-0.5 rounded-full bg-amber-500/20 text-amber-300 border border-amber-500/30 flex items-center gap-1">
                        <Tag class="w-2.5 h-2.5" />
                        {slot}
                      </span>
                    {/each}
                  {:else}
                    <span class="text-[11px] text-slate-500 italic">Unassigned sequence</span>
                  {/if}
                </div>
              </div>
            </div>

            <!-- Actions Row -->
            <div class="pt-3 mt-3 border-t border-dark-600/40 flex items-center justify-between gap-2">
              <span class="text-[11px] text-slate-400 font-semibold">Assign:</span>
              <div class="flex items-center gap-1.5">
                <button
                  type="button"
                  class="px-2.5 py-1 bg-dark-700 hover:bg-dark-600 text-slate-300 hover:text-white text-[11px] font-semibold rounded-lg transition"
                  on:click={() => handleAssignStoryboard('intro', sb)}
                  title="Assign as Game Opening Intro"
                >
                  Intro
                </button>
                <button
                  type="button"
                  class="px-2.5 py-1 bg-dark-700 hover:bg-dark-600 text-slate-300 hover:text-white text-[11px] font-semibold rounded-lg transition"
                  on:click={() => handleAssignStoryboard('ending', sb)}
                  title="Assign as Story Ending"
                >
                  Ending
                </button>
                <button
                  type="button"
                  class="px-2.5 py-1 bg-dark-700 hover:bg-dark-600 text-slate-300 hover:text-white text-[11px] font-semibold rounded-lg transition"
                  on:click={() => handleAssignStoryboard('credits', sb)}
                  title="Assign as Credits Roll"
                >
                  Credits
                </button>
              </div>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </main>
</div>

<!-- Add From Vault Modal -->
{#if showAddFromVaultModal}
  <AddFromVaultModal
    isOpen={showAddFromVaultModal}
    projectDir={$projectStore.current?.path || ''}
    targetCategory="storyboards"
    onClose={() => {
      showAddFromVaultModal = false;
      loadStoryboards();
    }}
  />
{/if}
