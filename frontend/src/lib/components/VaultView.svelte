<script lang="ts">
  import { onMount } from 'svelte';
  import type { AssetCategory, VaultAsset } from '../types';
  import {
    Package,
    FolderPlus,
    FolderOpen,
    UploadCloud,
    Search,
    Filter,
    Layers,
    Tag,
    Trash2,
    HardDrive,
    Sparkles,
    RefreshCw,
    X,
    Plus,
    LayoutGrid,
    List
  } from 'lucide-svelte';
  import { vaultStore, filteredAssets } from '../stores/vaultStore';
  import VaultCard from './VaultCard.svelte';
  import VaultInspector from './VaultInspector.svelte';
  import CreateVaultModal from './CreateVaultModal.svelte';
  import IngestPackModal from './IngestPackModal.svelte';

  let showCreateModal = false;
  let showPackModal = false;
  let pendingPackPath = '';
  let isDraggingOver = false;
  let viewMode: 'grid' | 'list' = 'grid';

  onMount(async () => {
    await vaultStore.loadVaults();
    await vaultStore.loadAssets();
  });

  function formatBytes(bytes: number): string {
    if (!bytes) return '0 B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  }

  function getCategoryBadge(cat: string): string {
    switch (cat) {
      case 'fighters':
        return 'bg-rose-500/20 text-rose-300 border-rose-500/30';
      case 'stages':
        return 'bg-amber-500/20 text-amber-300 border-amber-500/30';
      case 'motifs':
        return 'bg-purple-500/20 text-purple-300 border-purple-500/30';
      case 'sounds':
        return 'bg-cyan-500/20 text-cyan-300 border-cyan-500/30';
      default:
        return 'bg-slate-500/20 text-slate-300 border-slate-500/30';
    }
  }

  // Extract all unique tags across currently loaded assets
  $: allTags = Array.from(
    new Set(
      $vaultStore.assets
        .flatMap((a) => a.tags || [])
        .filter((t) => Boolean(t))
    )
  ).slice(0, 15);

  $: activeVaultObj = $vaultStore.vaults.find((v) => v.id === $vaultStore.activeVaultId);

  async function handleBrowseArchive() {
    const files = await vaultStore.browseArchives();
    if (files && files.length > 1) {
      // Direct bulk import for multiple selected files
      await vaultStore.ingestMultiple(files, $vaultStore.activeVaultId, 'auto');
    } else if (files && files.length === 1) {
      pendingPackPath = files[0];
      showPackModal = true;
    }
  }

  async function handleBrowseFolder() {
    const folder = await vaultStore.browseFolder();
    if (folder) {
      pendingPackPath = folder;
      showPackModal = true;
    }
  }

  async function handlePackConfirm(mode: 'new_vault' | 'current_vault') {
    showPackModal = false;
    if (pendingPackPath) {
      await vaultStore.ingest(pendingPackPath, $vaultStore.activeVaultId, mode);
      pendingPackPath = '';
    }
  }

  async function handleDrop(e: DragEvent) {
    e.preventDefault();
    isDraggingOver = false;
    if (e.dataTransfer && e.dataTransfer.files.length > 0) {
      const files = Array.from(e.dataTransfer.files);
      const paths: string[] = [];
      for (const file of files) {
        const path = (file as any).path || file.name;
        if (path) paths.push(path);
      }
      if (paths.length > 0) {
        await vaultStore.ingestMultiple(paths, $vaultStore.activeVaultId, 'auto');
      }
    }
  }

  function getCategoryCount(cat: AssetCategory | 'all'): number {
    if (cat === 'all') return $vaultStore.assets.length;
    return $vaultStore.assets.filter((a) => a.category === cat).length;
  }

  const categories: { id: AssetCategory | 'all'; label: string }[] = [
    { id: 'all', label: 'All Items' },
    { id: 'fighters', label: 'Fighters' },
    { id: 'stages', label: 'Stages' },
    { id: 'motifs', label: 'Screenpacks & Lifebars' },
    { id: 'sounds', label: 'Sounds & Audio' },
  ];
</script>

<div
  class="relative flex-1 flex h-full overflow-hidden bg-dark-900"
  style="--wails-drop-target: drop;"
  on:dragover|preventDefault={() => (isDraggingOver = true)}
  on:dragleave|preventDefault={() => (isDraggingOver = false)}
  on:drop={handleDrop}
>
  <!-- Drag-and-drop full-view overlay -->
  {#if isDraggingOver}
    <div class="absolute inset-0 z-40 bg-brand-950/80 backdrop-blur-md border-4 border-dashed border-brand-500 rounded-2xl m-4 flex flex-col items-center justify-center text-white pointer-events-none animate-in fade-in duration-150">
      <UploadCloud class="w-16 h-16 text-brand-400 animate-bounce mb-3" />
      <h3 class="text-xl font-bold">Drop Archives or Folders to Ingest</h3>
      <p class="text-xs text-slate-300 mt-1">Characters, stages, and audio will be automatically parsed into the Vault</p>
    </div>
  {/if}

  <!-- Main Content Area -->
  <div class="flex-1 flex flex-col h-full overflow-hidden">
    <!-- Top Action Bar -->
    <header class="p-6 border-b border-dark-700/80 bg-dark-850 flex flex-col gap-4">
      <div class="flex items-center justify-between">
        <!-- Title & Vault Switcher -->
        <div class="flex items-center gap-3">
          <div class="p-2.5 rounded-2xl bg-brand-500/10 border border-brand-500/20 text-brand-400 shadow-sm">
            <Package class="w-6 h-6" />
          </div>
          <div>
            <div class="flex items-center gap-2.5">
              <h2 class="text-xl font-black text-slate-100 tracking-tight">Asset Vault</h2>
              <!-- Vault Dropdown Selector -->
              <div class="relative">
                <select
                  value={$vaultStore.activeVaultId}
                  on:change={(e) => vaultStore.selectVault(e.currentTarget.value)}
                  class="bg-dark-900 border border-dark-600/80 hover:border-dark-500 text-slate-200 text-xs font-bold rounded-xl px-3 py-1.5 focus:outline-none focus:border-brand-500 transition cursor-pointer"
                >
                  <option value="all">All Registered Vaults ({$vaultStore.vaults.length})</option>
                  {#each $vaultStore.vaults as v}
                    <option value={v.id}>{v.name} {v.is_default ? '★ (Default)' : ''} ({v.asset_count} items)</option>
                  {/each}
                </select>
              </div>
            </div>
            <p class="text-xs text-slate-400 mt-0.5">
              {activeVaultObj ? activeVaultObj.path : 'Decentralized asset repository and multi-project library'}
            </p>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="flex items-center gap-2.5">
          <button
            on:click={() => (showCreateModal = true)}
            class="px-3 py-2 bg-dark-800 hover:bg-dark-700 border border-dark-600/80 text-slate-300 text-xs font-semibold rounded-xl transition flex items-center gap-1.5"
          >
            <FolderPlus class="w-4 h-4 text-brand-400" />
            <span>New / Mount Vault</span>
          </button>

          <button
            on:click={handleBrowseFolder}
            class="px-3 py-2 bg-dark-800 hover:bg-dark-700 border border-dark-600/80 text-slate-300 text-xs font-semibold rounded-xl transition flex items-center gap-1.5"
          >
            <FolderOpen class="w-4 h-4 text-amber-400" />
            <span>Ingest Folder</span>
          </button>

          <button
            on:click={handleBrowseArchive}
            class="px-4 py-2 bg-brand-600 hover:bg-brand-500 text-white text-xs font-bold rounded-xl transition flex items-center gap-1.5 shadow-md shadow-brand-600/20"
          >
            <Plus class="w-4 h-4" />
            <span>Ingest Archives (.zip, .rar, .7z)</span>
          </button>
        </div>
      </div>

      <!-- Ingest Progress Notification -->
      {#if $vaultStore.isIngesting}
        <div class="p-3 rounded-xl bg-brand-500/10 border border-brand-500/20 text-brand-300 text-xs flex items-center gap-2 animate-pulse">
          <RefreshCw class="w-4 h-4 animate-spin flex-shrink-0" />
          <span>{$vaultStore.ingestMessage}</span>
        </div>
      {/if}

      <!-- Search & Filters & View Switcher -->
      <div class="flex items-center justify-between gap-4 pt-1">
        <!-- Category Tabs -->
        <div class="flex items-center gap-1.5 p-1 bg-dark-950 rounded-xl border border-dark-700/60">
          {#each categories as cat}
            {@const count = getCategoryCount(cat.id)}
            <button
              class="px-3 py-1.5 text-xs font-semibold rounded-lg transition flex items-center gap-1.5 {$vaultStore.selectedCategory === cat.id ? 'bg-dark-800 text-brand-300 shadow-sm' : 'text-slate-400 hover:text-slate-200'}"
              on:click={() => vaultStore.setCategory(cat.id)}
            >
              <span>{cat.label}</span>
              <span class="text-[10px] px-1.5 py-0.2 rounded-full {$vaultStore.selectedCategory === cat.id ? 'bg-brand-500/20 text-brand-300' : 'bg-dark-900 text-slate-500'}">
                {count}
              </span>
            </button>
          {/each}
        </div>

        <div class="flex items-center gap-2.5">
          <!-- Search Bar -->
          <div class="relative w-72">
            <Search class="absolute left-3.5 top-2.5 w-4 h-4 text-slate-400" />
            <input
              type="text"
              placeholder="Search characters, authors, tags..."
              bind:value={$vaultStore.searchQuery}
              class="w-full bg-dark-950 border border-dark-600/80 rounded-xl pl-9 pr-8 py-2 text-xs text-slate-200 focus:outline-none focus:border-brand-500 transition"
            />
            {#if $vaultStore.searchQuery}
              <button
                class="absolute right-2.5 top-2.5 text-slate-500 hover:text-slate-300"
                on:click={() => vaultStore.setSearchQuery('')}
              >
                <X class="w-3.5 h-3.5" />
              </button>
            {/if}
          </div>

          <!-- Grid / List View Mode Switcher -->
          <div class="flex items-center p-1 bg-dark-950 rounded-xl border border-dark-700/60">
            <button
              class="p-1.5 rounded-lg transition {viewMode === 'grid' ? 'bg-dark-800 text-brand-300 shadow-sm' : 'text-slate-400 hover:text-slate-200'}"
              title="Grid View"
              on:click={() => (viewMode = 'grid')}
            >
              <LayoutGrid class="w-4 h-4" />
            </button>
            <button
              class="p-1.5 rounded-lg transition {viewMode === 'list' ? 'bg-dark-800 text-brand-300 shadow-sm' : 'text-slate-400 hover:text-slate-200'}"
              title="List / Table View"
              on:click={() => (viewMode = 'list')}
            >
              <List class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>

      <!-- Tag Filter Pills (if any exist) -->
      {#if allTags.length > 0}
        <div class="flex flex-wrap items-center gap-1.5 pt-1">
          <span class="text-[11px] font-semibold uppercase text-slate-500 flex items-center gap-1 mr-1">
            <Tag class="w-3 h-3" /> Tags:
          </span>
          {#each allTags as tag}
            {@const isSelected = $vaultStore.selectedTag === tag}
            <button
              class="text-[11px] px-2 py-0.5 rounded-lg border transition font-medium {isSelected ? 'bg-brand-500/20 text-brand-300 border-brand-500/40 shadow-sm' : 'bg-dark-900 text-slate-400 border-dark-700 hover:border-dark-600 hover:text-slate-200'}"
              on:click={() => vaultStore.toggleTag(tag)}
            >
              #{tag}
            </button>
          {/each}
          {#if $vaultStore.selectedTag}
            <button
              class="text-[11px] text-rose-400 hover:text-rose-300 ml-1 underline"
              on:click={() => vaultStore.toggleTag('')}
            >
              Clear Tag
            </button>
          {/if}
        </div>
      {/if}
    </header>

    <!-- Asset Cards / List View Container -->
    <div class="flex-1 overflow-y-auto p-6">
      {#if $filteredAssets.length === 0}
        <!-- Empty State -->
        <div class="h-full flex flex-col items-center justify-center p-12 text-center text-slate-500 border-2 border-dashed border-dark-700/60 rounded-3xl">
          <Package class="w-16 h-16 stroke-1 opacity-40 mb-3" />
          <h3 class="text-base font-bold text-slate-300">No Assets Found</h3>
          <p class="text-xs text-slate-500 max-w-sm mt-1.5">
            Drop character or stage archives (<code class="text-brand-400 font-mono">.zip, .rar, .7z</code>) or folders directly onto this view to ingest them into your Vault.
          </p>
          <div class="flex items-center gap-3 mt-4">
            <button
              on:click={handleBrowseArchive}
              class="px-4 py-2 bg-brand-600 hover:bg-brand-500 text-white text-xs font-bold rounded-xl transition shadow-sm"
            >
              Ingest Archives (.zip, .rar, .7z)
            </button>
          </div>
        </div>
      {:else if viewMode === 'grid'}
        <!-- Grid View -->
        <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
          {#each $filteredAssets as asset (asset.key)}
            <VaultCard
              {asset}
              onSelect={(a) => vaultStore.setSelectedAsset(a)}
            />
          {/each}
        </div>
      {:else}
        <!-- Table / List View -->
        <div class="bg-dark-850 border border-dark-700/80 rounded-2xl overflow-hidden shadow-sm">
          <table class="w-full text-left text-xs border-collapse">
            <thead>
              <tr class="bg-dark-900/80 border-b border-dark-700/80 text-[11px] font-bold uppercase tracking-wider text-slate-400 select-none">
                <th class="py-3 px-4 w-16">Preview</th>
                <th class="py-3 px-4">Name & Folder</th>
                <th class="py-3 px-4">Author</th>
                <th class="py-3 px-4">Category</th>
                <th class="py-3 px-4">Version</th>
                <th class="py-3 px-4">Size</th>
                <th class="py-3 px-4">Source</th>
                <th class="py-3 px-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-dark-700/50">
              {#each $filteredAssets as asset (asset.key)}
                <tr
                  class="hover:bg-dark-800/80 cursor-pointer transition group"
                  on:click={() => vaultStore.setSelectedAsset(asset)}
                >
                  <!-- Thumbnail -->
                  <td class="py-2.5 px-4">
                    <div class="w-10 h-10 rounded-xl bg-dark-900 border border-dark-700/80 overflow-hidden flex items-center justify-center flex-shrink-0">
                      {#if asset.preview_base64}
                        <img src={asset.preview_base64} alt={asset.display_name} class="w-full h-full object-contain" />
                      {:else}
                        <Package class="w-5 h-5 text-slate-600" />
                      {/if}
                    </div>
                  </td>

                  <!-- Name & Subtitle Key -->
                  <td class="py-2.5 px-4">
                    <div class="font-bold text-slate-100 group-hover:text-brand-300 transition-colors">
                      {asset.display_name || asset.key}
                    </div>
                    <div class="text-[11px] text-slate-500 font-mono">
                      {asset.key}
                    </div>
                  </td>

                  <!-- Author -->
                  <td class="py-2.5 px-4 text-slate-300 font-medium">
                    {asset.author || 'Unknown'}
                  </td>

                  <!-- Category -->
                  <td class="py-2.5 px-4">
                    <span class="text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full border {getCategoryBadge(asset.category)}">
                      {asset.category}
                    </span>
                  </td>

                  <!-- Version -->
                  <td class="py-2.5 px-4 text-slate-400">
                    {asset.mugen_version || asset.ikemen_version || '—'}
                  </td>

                  <!-- Size -->
                  <td class="py-2.5 px-4 text-slate-400 font-mono">
                    {formatBytes(asset.size_bytes)}
                  </td>

                  <!-- Source Package -->
                  <td class="py-2.5 px-4 text-slate-400 max-w-[140px] truncate" title={asset.source_url || asset.source_package}>
                    {asset.source_package || 'Local'}
                  </td>

                  <!-- Actions -->
                  <td class="py-2.5 px-4 text-right">
                    <button
                      class="px-2.5 py-1 text-[11px] font-semibold text-slate-300 hover:text-white bg-dark-800 hover:bg-dark-700 border border-dark-600 rounded-lg transition"
                      on:click|stopPropagation={() => vaultStore.setSelectedAsset(asset)}
                    >
                      Inspect
                    </button>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </div>
  </div>

  <!-- Slide-out Asset Inspector -->
  {#if $vaultStore.selectedAsset}
    <VaultInspector
      asset={$vaultStore.selectedAsset}
      activeVaultId={$vaultStore.activeVaultId !== 'all' ? $vaultStore.activeVaultId : 'vault-default'}
      onClose={() => vaultStore.setSelectedAsset(null)}
    />
  {/if}
</div>

<!-- Modals -->
<CreateVaultModal
  isOpen={showCreateModal}
  onClose={() => (showCreateModal = false)}
/>

<IngestPackModal
  isOpen={showPackModal}
  filePath={pendingPackPath}
  activeVaultName={activeVaultObj?.name || 'Default Vault'}
  onConfirm={handlePackConfirm}
  onCancel={() => { showPackModal = false; pendingPackPath = ''; }}
/>
