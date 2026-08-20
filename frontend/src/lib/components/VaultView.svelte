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
    List,
    CheckSquare,
    Square,
    ArrowUpDown,
    DownloadCloud,
    FolderDown,
    Link2,
    Check
  } from 'lucide-svelte';
  import { vaultStore, filteredAssets } from '../stores/vaultStore';
  import { projectStore } from '../stores/projectStore';
  import { toastStore } from '../stores/toastStore';
  import VaultCard from './VaultCard.svelte';
  import VaultInspector from './VaultInspector.svelte';
  import CreateVaultModal from './CreateVaultModal.svelte';
  import IngestPackModal from './IngestPackModal.svelte';

  let showCreateModal = false;
  let showPackModal = false;
  let pendingPackPath = '';
  let isDraggingOver = false;
  let viewMode: 'grid' | 'list' = 'grid';

  // Multi-selection state
  let selectedKeys = new Set<string>();
  let sortBy: 'name_asc' | 'name_desc' | 'size_desc' | 'category' = 'name_asc';
  let isPerformingBulkAction = false;

  onMount(async () => {
    await vaultStore.loadVaults();
    await vaultStore.loadAssets();
  });

  function toggleSelectKey(key: string) {
    if (selectedKeys.has(key)) {
      selectedKeys.delete(key);
    } else {
      selectedKeys.add(key);
    }
    selectedKeys = new Set(selectedKeys);
  }

  function selectAll() {
    selectedKeys = new Set($filteredAssets.map((a) => a.key));
  }

  function deselectAll() {
    selectedKeys = new Set();
  }

  async function handleBulkAddToProject() {
    if (!$projectStore.current) {
      toastStore.warning('No Project Open', 'Please open or create a project first.');
      return;
    }
    if (selectedKeys.size === 0) return;

    isPerformingBulkAction = true;
    let addedCount = 0;
    const targetDir = $projectStore.current.path;

    for (const key of selectedKeys) {
      const asset = $vaultStore.assets.find((a) => a.key === key);
      const vaultId = (asset as any)?.vault_id || $vaultStore.activeVaultId || 'vault-default';
      const ok = await vaultStore.linkToProject(targetDir, vaultId, key);
      if (ok) addedCount++;
    }

    isPerformingBulkAction = false;
    toastStore.success('Bulk Import Complete', `Added ${addedCount} asset(s) to project`);
    deselectAll();
  }

  async function handleBulkDelete() {
    if (selectedKeys.size === 0) return;
    if (!confirm(`Are you sure you want to permanently delete ${selectedKeys.size} selected item(s) from the Vault?`)) {
      return;
    }

    isPerformingBulkAction = true;
    let deletedCount = 0;

    for (const key of selectedKeys) {
      const asset = $vaultStore.assets.find((a) => a.key === key);
      const vaultId = (asset as any)?.vault_id || $vaultStore.activeVaultId || 'vault-default';
      const ok = await vaultStore.deleteAsset(vaultId, key);
      if (ok) deletedCount++;
    }

    isPerformingBulkAction = false;
    toastStore.info('Bulk Delete Finished', `Removed ${deletedCount} asset(s) from vault`);
    deselectAll();
  }

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

  // Sorted and filtered list
  $: sortedAssets = [...$filteredAssets].sort((a, b) => {
    const nameA = (a.display_name || a.key).toLowerCase();
    const nameB = (b.display_name || b.key).toLowerCase();
    if (sortBy === 'name_asc') return nameA.localeCompare(nameB);
    if (sortBy === 'name_desc') return nameB.localeCompare(nameA);
    if (sortBy === 'size_desc') return (b.size_bytes || 0) - (a.size_bytes || 0);
    if (sortBy === 'category') return a.category.localeCompare(b.category);
    return 0;
  });

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
          <!-- Sort Dropdown -->
          <div class="flex items-center gap-1.5 bg-dark-950 px-2.5 py-1.5 rounded-xl border border-dark-700/60 text-xs">
            <ArrowUpDown class="w-3.5 h-3.5 text-slate-400" />
            <select
              bind:value={sortBy}
              class="bg-transparent text-slate-300 text-xs font-medium focus:outline-none cursor-pointer"
            >
              <option value="name_asc" class="bg-dark-900">Name (A-Z)</option>
              <option value="name_desc" class="bg-dark-900">Name (Z-A)</option>
              <option value="size_desc" class="bg-dark-900">Largest Size</option>
              <option value="category" class="bg-dark-900">Category</option>
            </select>
          </div>

          <!-- Select All / Deselect Toggle -->
          <button
            type="button"
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl border transition text-xs font-semibold {
              selectedKeys.size > 0 && selectedKeys.size === $filteredAssets.length
                ? 'bg-brand-600 text-white border-brand-500'
                : selectedKeys.size > 0
                ? 'bg-brand-500/20 text-brand-300 border-brand-500/40'
                : 'bg-dark-950 text-slate-400 border-dark-700/60 hover:text-slate-200'
            }"
            on:click={() => {
              if (selectedKeys.size === $filteredAssets.length) {
                deselectAll();
              } else {
                selectAll();
              }
            }}
          >
            {#if selectedKeys.size === $filteredAssets.length && $filteredAssets.length > 0}
              <CheckSquare class="w-3.5 h-3.5" />
              <span>All Selected</span>
            {:else if selectedKeys.size > 0}
              <CheckSquare class="w-3.5 h-3.5 text-brand-400" />
              <span>{selectedKeys.size} Selected</span>
            {:else}
              <Square class="w-3.5 h-3.5" />
              <span>Select All</span>
            {/if}
          </button>

          <!-- Search Bar -->
          <div class="relative w-64">
            <Search class="absolute left-3.5 top-2.5 w-4 h-4 text-slate-400" />
            <input
              type="text"
              placeholder="Search characters, authors..."
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
    <div class="flex-1 overflow-y-auto p-6 relative">
      {#if sortedAssets.length === 0}
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
          {#each sortedAssets as asset (asset.key)}
            <VaultCard
              {asset}
              selected={selectedKeys.has(asset.key)}
              onToggleSelect={(a) => toggleSelectKey(a.key)}
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
                <th class="py-3 px-3 w-10 text-center">
                  <button
                    type="button"
                    class="text-slate-400 hover:text-white"
                    on:click={() => {
                      if (selectedKeys.size === sortedAssets.length) deselectAll();
                      else selectAll();
                    }}
                  >
                    {#if selectedKeys.size === sortedAssets.length && sortedAssets.length > 0}
                      <CheckSquare class="w-4 h-4 text-brand-400" />
                    {:else}
                      <Square class="w-4 h-4" />
                    {/if}
                  </button>
                </th>
                <th class="py-3 px-3 w-14">Preview</th>
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
              {#each sortedAssets as asset (asset.key)}
                {@const isSelected = selectedKeys.has(asset.key)}
                <tr
                  class="cursor-pointer transition group {isSelected ? 'bg-brand-950/30 hover:bg-brand-950/40' : 'hover:bg-dark-800/80'}"
                  on:click={() => vaultStore.setSelectedAsset(asset)}
                >
                  <!-- Selection Checkbox -->
                  <td class="py-2.5 px-3 text-center" on:click|stopPropagation={() => toggleSelectKey(asset.key)}>
                    <button type="button" class="text-slate-400 hover:text-white">
                      {#if isSelected}
                        <CheckSquare class="w-4 h-4 text-brand-400" />
                      {:else}
                        <Square class="w-4 h-4" />
                      {/if}
                    </button>
                  </td>

                  <!-- Thumbnail -->
                  <td class="py-2.5 px-3">
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
                    <div class="flex items-center justify-end gap-1.5">
                      <button
                        class="p-1.5 text-slate-500 hover:text-rose-400 hover:bg-rose-500/10 rounded-lg transition"
                        title="Delete from Vault"
                        on:click|stopPropagation={() => {
                          if (confirm(`Delete ${asset.display_name || asset.key} from Vault?`)) {
                            vaultStore.deleteAsset($vaultStore.activeVaultId, asset.key);
                          }
                        }}
                      >
                        <Trash2 class="w-3.5 h-3.5" />
                      </button>
                      <button
                        class="px-2.5 py-1 text-[11px] font-semibold text-slate-300 hover:text-white bg-dark-800 hover:bg-dark-700 border border-dark-600 rounded-lg transition"
                        on:click|stopPropagation={() => vaultStore.setSelectedAsset(asset)}
                      >
                        Inspect
                      </button>
                    </div>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}

      <!-- Floating Bulk Action Toolbar -->
      {#if selectedKeys.size > 0}
        <div class="fixed bottom-6 left-1/2 -translate-x-1/2 z-30 bg-dark-850/95 border border-brand-500/60 backdrop-blur-xl shadow-2xl rounded-2xl px-5 py-3 flex items-center gap-4 animate-in slide-in-from-bottom-4">
          <div class="flex items-center gap-2 pr-3 border-r border-dark-700">
            <span class="w-2.5 h-2.5 rounded-full bg-brand-400 animate-pulse"></span>
            <span class="text-xs font-bold text-slate-100">{selectedKeys.size} item(s) selected</span>
          </div>

          {#if $projectStore.current}
            <button
              type="button"
              disabled={isPerformingBulkAction}
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-brand-600 hover:bg-brand-500 text-white text-xs font-bold transition shadow-md shadow-brand-600/20 disabled:opacity-50"
              on:click={handleBulkAddToProject}
            >
              <FolderDown class="w-3.5 h-3.5" />
              <span>Add to {$projectStore.current.name}</span>
            </button>
          {/if}

          <button
            type="button"
            disabled={isPerformingBulkAction}
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-rose-600/20 hover:bg-rose-600/30 border border-rose-500/40 text-rose-300 text-xs font-bold transition disabled:opacity-50"
            on:click={handleBulkDelete}
          >
            <Trash2 class="w-3.5 h-3.5" />
            <span>Delete Selected</span>
          </button>

          <button
            type="button"
            class="px-2.5 py-1.5 rounded-xl text-slate-400 hover:text-slate-200 text-xs font-semibold hover:bg-dark-700 transition"
            on:click={deselectAll}
          >
            Deselect All
          </button>
        </div>
      {/if}
    </div>
  </div>

  <!-- Slide-out Asset Inspector -->
  {#if $vaultStore.selectedAsset}
    <VaultInspector
      asset={$vaultStore.selectedAsset}
      activeVaultId={$vaultStore.activeVaultId}
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
