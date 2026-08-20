<script lang="ts">
  import { onMount } from 'svelte';
  import { projectStore } from '../stores/projectStore';
  import { vaultStore } from '../stores/vaultStore';
  import { toastStore } from '../stores/toastStore';
  import type { ProjectMotifInfo } from '../types';
  import AddFromVaultModal from './AddFromVaultModal.svelte';
  import {
    Sliders,
    ArrowLeft,
    Sparkles,
    Plus,
    FolderOpen,
    Play,
    CheckCircle2,
    Check,
    Loader2,
    Grid,
    Monitor,
    Maximize2,
    Palette,
  } from 'lucide-svelte';

  export let onBackToWorkspace: () => void;

  let motifs: ProjectMotifInfo[] = [];
  let loading = true;
  let showAddFromVaultModal = false;

  onMount(async () => {
    await loadMotifs();
  });

  async function loadMotifs() {
    loading = true;
    motifs = await projectStore.loadMotifs();
    loading = false;
  }

  async function handleActivateMotif(motif: ProjectMotifInfo) {
    const ok = await projectStore.activateMotif(motif.key);
    if (ok) {
      for (const m of motifs) {
        m.is_active = m.key === motif.key;
      }
      motifs = [...motifs];
    }
  }

  async function handleBrowseArchive() {
    try {
      const archives = await (window as any).go.main.App.SelectMultipleArchivesDialog();
      if (archives && archives.length > 0) {
        loading = true;
        for (const arch of archives) {
          await vaultStore.ingestAsset(arch, 'motifs');
        }
        await loadMotifs();
        toastStore.success('Motifs Ingested', `Processed ${archives.length} archive(s)`);
      }
    } catch (err: any) {
      toastStore.error('Import Failed', err?.message || 'Could not import screenpack');
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
            <Palette class="w-5 h-5 text-indigo-400" />
            <span>Screenpacks & Motifs</span>
          </h1>
          <span class="text-xs px-2 py-0.5 rounded-full bg-indigo-500/10 text-indigo-300 border border-indigo-500/20 font-mono">
            {motifs.length} Themes
          </span>
        </div>
        <p class="text-xs text-slate-400 mt-0.5">Configure title screen, select grid dimensions, and switch the active game screenpack (config.ini)</p>
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
        title="Link screenpack from Asset Vault"
      >
        <Sparkles class="w-4 h-4 text-brand-400" />
        <span>+ From Vault</span>
      </button>

      <button
        type="button"
        class="px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold flex items-center gap-2 transition shadow-md shadow-indigo-950/40"
        on:click={handleBrowseArchive}
        title="Import screenpack archive (.zip, .rar, .7z)"
      >
        <Plus class="w-4 h-4" />
        <span>Ingest Screenpack</span>
      </button>
    </div>
  </header>

  <!-- Main Content Area -->
  <main class="flex-1 overflow-y-auto p-6">
    {#if loading}
      <div class="h-full flex flex-col items-center justify-center gap-3 text-slate-400">
        <Loader2 class="w-8 h-8 animate-spin text-indigo-400" />
        <span class="text-xs">Analyzing system motifs and UI themes...</span>
      </div>
    {:else if motifs.length === 0}
      <div class="h-full flex flex-col items-center justify-center p-12 text-center text-slate-500 border-2 border-dashed border-dark-700/60 rounded-3xl">
        <Palette class="w-16 h-16 stroke-1 opacity-40 mb-3 text-indigo-400" />
        <h3 class="text-base font-bold text-slate-300">No Motifs Found</h3>
        <p class="text-xs text-slate-500 max-w-sm mt-1.5">
          Install a custom screenpack or ensure data/system.def exists in your project.
        </p>
      </div>
    {:else}
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {#each motifs as motif (motif.key)}
          <div
            class="p-5 rounded-2xl bg-dark-800 border transition-all flex flex-col justify-between group shadow-md {motif.is_active ? 'border-indigo-500/80 shadow-indigo-950/30 ring-1 ring-indigo-500/40' : 'border-dark-600/70 hover:border-dark-600'}"
          >
            <div class="space-y-4">
              <!-- Banner Preview -->
              <div class="w-full h-36 rounded-xl bg-dark-900 border border-dark-700/80 overflow-hidden flex items-center justify-center relative">
                {#if motif.preview_base64}
                  <img src={motif.preview_base64} alt={motif.display_name} class="w-full h-full object-cover" />
                {:else}
                  <div class="flex flex-col items-center gap-2 text-slate-600">
                    <Palette class="w-10 h-10 text-indigo-400/50" />
                    <span class="text-xs font-mono">{motif.resolution}</span>
                  </div>
                {/if}

                {#if motif.is_active}
                  <span class="absolute top-3 right-3 px-3 py-1 rounded-full bg-emerald-500 text-slate-950 text-xs font-black uppercase tracking-wider shadow-lg flex items-center gap-1">
                    <CheckCircle2 class="w-3.5 h-3.5 fill-current" />
                    Active Theme
                  </span>
                {/if}
              </div>

              <!-- Title & Meta -->
              <div>
                <h3 class="text-base font-bold text-slate-100 group-hover:text-indigo-300 transition truncate">
                  {motif.display_name}
                </h3>
                <div class="flex items-center gap-2 text-xs text-slate-400 mt-1">
                  <span>By {motif.author}</span>
                  <span>&bull;</span>
                  <span class="font-mono text-[11px] text-slate-500 truncate">{motif.key}</span>
                </div>
              </div>

              <!-- Layout Specs Grid -->
              <div class="grid grid-cols-2 gap-2 pt-2 text-xs">
                <div class="p-2.5 rounded-xl bg-dark-900/70 border border-dark-700/60 flex items-center gap-2">
                  <Monitor class="w-4 h-4 text-cyan-400" />
                  <div>
                    <div class="text-[10px] text-slate-500 font-semibold uppercase">Resolution</div>
                    <div class="font-mono text-slate-200 font-bold">{motif.resolution}</div>
                  </div>
                </div>

                <div class="p-2.5 rounded-xl bg-dark-900/70 border border-dark-700/60 flex items-center gap-2">
                  <Grid class="w-4 h-4 text-purple-400" />
                  <div>
                    <div class="text-[10px] text-slate-500 font-semibold uppercase">Select Grid</div>
                    <div class="font-mono text-slate-200 font-bold">{motif.grid_columns} &times; {motif.grid_rows} ({motif.total_slots})</div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Actions Row -->
            <div class="pt-4 mt-4 border-t border-dark-600/40 flex items-center justify-between gap-3">
              <button
                type="button"
                class="px-3 py-2 rounded-xl bg-dark-700 hover:bg-dark-600 text-slate-300 text-xs font-semibold flex items-center gap-1.5 transition"
                on:click={() => projectStore.launch()}
                title="Launch game to test screenpack"
              >
                <Play class="w-3.5 h-3.5 fill-current" />
                <span>Test Game</span>
              </button>

              {#if motif.is_active}
                <div class="flex items-center gap-1.5 text-xs font-bold text-emerald-400 px-3 py-2 bg-emerald-500/10 rounded-xl border border-emerald-500/30">
                  <Check class="w-4 h-4" />
                  <span>Selected</span>
                </div>
              {:else}
                <button
                  type="button"
                  class="px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold transition shadow-md shadow-indigo-950/40"
                  on:click={() => handleActivateMotif(motif)}
                >
                  Activate Screenpack
                </button>
              {/if}
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
    targetCategory="motifs"
    onClose={() => {
      showAddFromVaultModal = false;
      loadMotifs();
    }}
  />
{/if}
