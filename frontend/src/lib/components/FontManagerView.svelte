<script lang="ts">
  import { onMount } from 'svelte';
  import { projectStore } from '../stores/projectStore';
  import { vaultStore } from '../stores/vaultStore';
  import { toastStore } from '../stores/toastStore';
  import type { ProjectFontInfo } from '../types';
  import AddFromVaultModal from './AddFromVaultModal.svelte';
  import {
    Type,
    ArrowLeft,
    Sparkles,
    Plus,
    FolderOpen,
    Check,
    Loader2,
    Search,
    Sliders,
    Tag,
  } from 'lucide-svelte';

  export let onBackToWorkspace: () => void;

  let fonts: ProjectFontInfo[] = [];
  let loading = true;
  let searchQuery = '';
  let showAddFromVaultModal = false;

  onMount(async () => {
    await loadFonts();
  });

  async function loadFonts() {
    loading = true;
    fonts = await projectStore.loadFonts();
    loading = false;
  }

  $: filteredFonts = fonts.filter((f) => {
    const q = searchQuery.toLowerCase();
    return (
      f.file_name.toLowerCase().includes(q) ||
      f.font_type.toLowerCase().includes(q) ||
      f.system_slot_mappings.some((s) => s.toLowerCase().includes(q))
    );
  });

  async function handleMapFont(targetDef: 'system' | 'fight', slot: string, font: ProjectFontInfo) {
    const ok = await projectStore.setSystemFontMapping(targetDef, slot, font.relative_path);
    if (ok) {
      await loadFonts();
    }
  }

  async function handleBrowseFont() {
    try {
      const filePath = await projectStore.selectFontFileDialog();
      if (filePath) {
        await vaultStore.ingestAsset(filePath, 'fonts');
        await loadFonts();
        toastStore.success('Font Ingested', filePath);
      }
    } catch (err: any) {
      toastStore.error('Import Failed', err?.message || 'Could not import font');
    }
  }

  function formatBytes(b: number): string {
    if (b < 1024) return `${b} B`;
    if (b < 1024 * 1024) return `${(b / 1024).toFixed(1)} KB`;
    return `${(b / (1024 * 1024)).toFixed(1)} MB`;
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
            <Type class="w-5 h-5 text-purple-400" />
            <span>Fonts & Typography</span>
          </h1>
          <span class="text-xs px-2 py-0.5 rounded-full bg-purple-500/10 text-purple-300 border border-purple-500/20 font-mono">
            {fonts.length} Fonts
          </span>
        </div>
        <p class="text-xs text-slate-400 mt-0.5">Manage bitmap (.fnt), Ikemen font defs, TrueType fonts, and slot mappings in system.def & fight.def</p>
      </div>
    </div>

    <!-- Actions -->
    <div class="flex items-center gap-2.5 flex-wrap">
      <button
        type="button"
        class="px-3.5 py-2 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600/70 text-slate-200 text-xs font-semibold flex items-center gap-2 transition shadow-sm"
        on:click={() => projectStore.openFolder('font')}
        title="Open font/ folder"
      >
        <FolderOpen class="w-4 h-4 text-indigo-400" />
        <span>Open font/</span>
      </button>

      <button
        type="button"
        class="px-3.5 py-2 rounded-xl bg-brand-600/20 hover:bg-brand-600/30 border border-brand-500/40 text-brand-300 text-xs font-semibold flex items-center gap-2 transition shadow-sm"
        on:click={() => (showAddFromVaultModal = true)}
        title="Link fonts from Asset Vault"
      >
        <Sparkles class="w-4 h-4 text-brand-400" />
        <span>+ From Vault</span>
      </button>

      <button
        type="button"
        class="px-4 py-2 rounded-xl bg-purple-600 hover:bg-purple-500 text-white text-xs font-bold flex items-center gap-2 transition shadow-md shadow-purple-950/40"
        on:click={handleBrowseFont}
        title="Import font file (.fnt, .def, .ttf)"
      >
        <Plus class="w-4 h-4" />
        <span>Add Font</span>
      </button>
    </div>
  </header>

  <!-- Search Bar -->
  <div class="p-4 px-6 border-b border-dark-600/40 bg-dark-900 flex items-center justify-between gap-4 flex-shrink-0">
    <div class="relative flex-1 max-w-md">
      <Search class="w-4 h-4 text-slate-500 absolute left-3 top-1/2 -translate-y-1/2" />
      <input
        type="text"
        bind:value={searchQuery}
        placeholder="Search fonts by name, type, or slot mapping..."
        class="w-full bg-dark-800 border border-dark-600/70 rounded-xl pl-9 pr-3 py-1.5 text-xs text-slate-200 placeholder:text-slate-500 focus:outline-none focus:border-purple-500 transition"
      />
    </div>
  </div>

  <!-- Main Content Area -->
  <main class="flex-1 overflow-y-auto p-6">
    {#if loading}
      <div class="h-full flex flex-col items-center justify-center gap-3 text-slate-400">
        <Loader2 class="w-8 h-8 animate-spin text-purple-400" />
        <span class="text-xs">Scanning font definitions and typefaces...</span>
      </div>
    {:else if filteredFonts.length === 0}
      <div class="h-full flex flex-col items-center justify-center p-12 text-center text-slate-500 border-2 border-dashed border-dark-700/60 rounded-3xl">
        <Type class="w-16 h-16 stroke-1 opacity-40 mb-3 text-purple-400" />
        <h3 class="text-base font-bold text-slate-300">No Fonts Found</h3>
        <p class="text-xs text-slate-500 max-w-sm mt-1.5">
          Add font files (.fnt, .def, .ttf) to font/ or link them from Vault.
        </p>
      </div>
    {:else}
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {#each filteredFonts as font (font.relative_path)}
          <div
            class="p-5 rounded-2xl bg-dark-800 border border-dark-600/70 hover:border-purple-500/50 transition-all flex flex-col justify-between group shadow-sm"
          >
            <div class="space-y-4">
              <!-- Specimen Sample Box -->
              <div class="p-4 rounded-xl bg-dark-900 border border-dark-700/80 flex flex-col justify-between min-h-[90px] group-hover:border-purple-500/30 transition">
                <div class="text-lg font-black text-slate-100 tracking-wider">
                  FIGHT! K.O. 99 COMBO
                </div>
                <div class="text-xs font-mono text-slate-400 mt-2">
                  0123456789 ABCDEFGHIJKLMNOPQRSTUVWXYZ
                </div>
              </div>

              <!-- Meta Info -->
              <div>
                <div class="flex items-center justify-between gap-2">
                  <h3 class="text-sm font-bold text-slate-100 group-hover:text-purple-300 transition truncate">
                    {font.file_name}
                  </h3>
                  <span class="text-[10px] font-mono px-2 py-0.5 rounded-full bg-purple-500/10 text-purple-300 border border-purple-500/30 font-semibold">
                    {font.font_type}
                  </span>
                </div>
                <div class="flex items-center gap-2 text-xs text-slate-500 mt-1">
                  <span>{formatBytes(font.size_bytes)}</span>
                  <span>&bull;</span>
                  <span class="font-mono text-[11px] truncate">{font.relative_path}</span>
                </div>
              </div>

              <!-- Active Slot Mappings -->
              <div class="space-y-1.5 pt-1">
                <div class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Active Mappings:</div>
                <div class="flex flex-wrap gap-1.5 min-h-[24px]">
                  {#if font.system_slot_mappings.length > 0}
                    {#each font.system_slot_mappings as mapping}
                      <span class="text-[10px] font-bold px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 flex items-center gap-1">
                        <Tag class="w-2.5 h-2.5" />
                        {mapping}
                      </span>
                    {/each}
                  {:else}
                    <span class="text-[11px] text-slate-500 italic">Unassigned slot</span>
                  {/if}
                </div>
              </div>
            </div>

            <!-- Mapping Controls -->
            <div class="pt-3 mt-3 border-t border-dark-600/40">
              <div class="flex items-center justify-between gap-2">
                <span class="text-[11px] text-slate-400 font-semibold">Quick Map:</span>
                <div class="flex items-center gap-1.5 flex-wrap">
                  <button
                    type="button"
                    class="px-2 py-1 bg-dark-700 hover:bg-dark-600 text-slate-300 hover:text-white text-[10px] font-semibold rounded-lg transition"
                    on:click={() => handleMapFont('system', 'font1', font)}
                    title="Map as System Menu font (font1)"
                  >
                    Sys 1
                  </button>
                  <button
                    type="button"
                    class="px-2 py-1 bg-dark-700 hover:bg-dark-600 text-slate-300 hover:text-white text-[10px] font-semibold rounded-lg transition"
                    on:click={() => handleMapFont('system', 'font2', font)}
                    title="Map as System Select font (font2)"
                  >
                    Sys 2
                  </button>
                  <button
                    type="button"
                    class="px-2 py-1 bg-dark-700 hover:bg-dark-600 text-slate-300 hover:text-white text-[10px] font-semibold rounded-lg transition"
                    on:click={() => handleMapFont('fight', 'font1', font)}
                    title="Map as Fight HUD Combo font (font1)"
                  >
                    HUD 1
                  </button>
                  <button
                    type="button"
                    class="px-2 py-1 bg-dark-700 hover:bg-dark-600 text-slate-300 hover:text-white text-[10px] font-semibold rounded-lg transition"
                    on:click={() => handleMapFont('fight', 'font2', font)}
                    title="Map as Fight HUD Name font (font2)"
                  >
                    HUD 2
                  </button>
                </div>
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
    targetCategory="fonts"
    onClose={() => {
      showAddFromVaultModal = false;
      loadFonts();
    }}
  />
{/if}
