<script lang="ts">
  import { onMount } from 'svelte';
  import { projectStore } from '../stores/projectStore';
  import { vaultStore } from '../stores/vaultStore';
  import type { ProjectRoster, RosterCharacterSlot, RosterAvailableCharacter, VaultAsset } from '../types';
  import {
    Users,
    Grid,
    Sparkles,
    Save,
    RotateCcw,
    Plus,
    Trash2,
    Music,
    Mountain,
    HelpCircle,
    EyeOff,
    Check,
    Search,
    Package,
    ArrowLeft,
    Sliders,
    Layers,
    DownloadCloud,
    FolderPlus,
    X,
    GripVertical,
    CheckCircle2,
    Crown,
    Volume2,
  } from 'lucide-svelte';

  export let onBackToWorkspace: () => void;

  let loading = true;
  let saving = false;
  let exportingToVault = false;
  let roster: ProjectRoster | null = null;
  let selectedSlotIndex: number | null = null;

  // Grid dimension overrides (can override system.def visually)
  let gridColumns = 5;
  let gridRows = 2;

  // Sidebar library state
  let libraryTab: 'project' | 'vault' | 'stages' = 'project';
  let librarySearch = '';
  let vaultFilter = 'all';

  // Drag and drop state
  let draggedSlotIndex: number | null = null;
  let draggedLibraryItem: { type: 'char' | 'vault'; name: string; asset?: any } | null = null;

  // Active subtab in roster view
  let activeView: 'grid' | 'extra_stages' = 'grid';

  onMount(async () => {
    await loadRosterData();
    await vaultStore.loadVaults();
    await vaultStore.loadAssets();
  });

  async function loadRosterData() {
    loading = true;
    const data = await projectStore.loadRoster();
    if (data) {
      roster = data;
      gridColumns = Math.max(1, data.grid?.columns || 5);
      const totalSlots = data.slots.length;
      gridRows = Math.max(2, Math.ceil(totalSlots / gridColumns));
      if (data.slots.length > 0) {
        selectedSlotIndex = 0;
      }
    }
    loading = false;
  }

  async function handleSave() {
    if (!roster) return;
    saving = true;
    await projectStore.saveRoster(roster);
    saving = false;
  }

  async function handleExportToVault() {
    if (!$projectStore.current) return;
    exportingToVault = true;
    const targetVault = $vaultStore.vaults[0]?.id || 'vault-default';
    await projectStore.exportToVault(targetVault);
    await vaultStore.loadAssets();
    exportingToVault = false;
  }

  $: selectedSlot = roster && selectedSlotIndex !== null && roster.slots[selectedSlotIndex]
    ? roster.slots[selectedSlotIndex]
    : null;

  function getAssetName(key: string): string {
    const parts = key.replace(/\\/g, '/').split('/').filter(Boolean);
    return parts[parts.length - 1] || key;
  }

  $: filteredProjectChars = (roster?.available_characters || []).filter((c) =>
    c.name.toLowerCase().includes(librarySearch.toLowerCase()) ||
    c.display_name.toLowerCase().includes(librarySearch.toLowerCase()) ||
    c.author.toLowerCase().includes(librarySearch.toLowerCase())
  );

  $: filteredVaultAssets = ($vaultStore.assets || []).filter((a) => {
    if (a.category !== 'fighters') return false;
    const q = librarySearch.toLowerCase();
    const name = getAssetName(a.key);
    return (
      name.toLowerCase().includes(q) ||
      (a.display_name && a.display_name.toLowerCase().includes(q)) ||
      (a.author && a.author.toLowerCase().includes(q))
    );
  });

  $: filteredAvailableStages = (roster?.available_stages || []).filter((s) =>
    s.toLowerCase().includes(librarySearch.toLowerCase())
  );

  // --- Slot manipulation actions ---

  function addEmptySlot() {
    if (!roster) return;
    const nextIdx = roster.slots.length;
    roster.slots.push({
      index: nextIdx,
      type: 'empty',
      character: '',
      include_in_arcade: true,
    });
    roster.slots = [...roster.slots];
    selectedSlotIndex = nextIdx;
  }

  function addRandomSelectSlot() {
    if (!roster) return;
    const nextIdx = roster.slots.length;
    roster.slots.push({
      index: nextIdx,
      type: 'randomselect',
      character: 'randomselect',
      display_name: 'Random Select',
      include_in_arcade: true,
    });
    roster.slots = [...roster.slots];
    selectedSlotIndex = nextIdx;
  }

  function addCharacterToNextSlot(charName: string, displayName?: string, author?: string, portrait?: string) {
    if (!roster) return;
    // Find first empty slot or append
    let targetIdx = roster.slots.findIndex((s) => s.type === 'empty');
    if (targetIdx === -1) {
      targetIdx = roster.slots.length;
      roster.slots.push({
        index: targetIdx,
        type: 'character',
        character: charName,
        display_name: displayName || charName,
        author: author || 'Unknown',
        portrait_base64: portrait || '',
        include_in_arcade: true,
        order: 1,
      });
    } else {
      roster.slots[targetIdx] = {
        index: targetIdx,
        type: 'character',
        character: charName,
        display_name: displayName || charName,
        author: author || 'Unknown',
        portrait_base64: portrait || '',
        include_in_arcade: true,
        order: 1,
      };
    }
    roster.slots = [...roster.slots];
    selectedSlotIndex = targetIdx;
  }

  async function handleAddVaultCharToProject(asset: VaultAsset) {
    if (!$projectStore.current) return;
    const vaultId = (asset as any).vault_id || $vaultStore.activeVaultId || 'vault-default';
    await vaultStore.linkToProject($projectStore.current.path, vaultId, asset.key);
    const charName = getAssetName(asset.key);
    addCharacterToNextSlot(charName, asset.display_name, asset.author, asset.preview_base64);
  }

  function deleteSlot(index: number) {
    if (!roster) return;
    roster.slots.splice(index, 1);
    // Re-index remaining
    roster.slots.forEach((s, i) => (s.index = i));
    roster.slots = [...roster.slots];
    if (selectedSlotIndex !== null && selectedSlotIndex >= roster.slots.length) {
      selectedSlotIndex = roster.slots.length > 0 ? roster.slots.length - 1 : null;
    }
  }

  function clearSlot(index: number) {
    if (!roster || !roster.slots[index]) return;
    roster.slots[index] = {
      index,
      type: 'empty',
      character: '',
      include_in_arcade: true,
    };
    roster.slots = [...roster.slots];
  }

  function toggleSlotDisabled(index: number) {
    if (!roster || !roster.slots[index]) return;
    const s = roster.slots[index];
    if (s.type === 'disabled') {
      s.type = 'character';
    } else if (s.type === 'character') {
      s.type = 'disabled';
    }
    roster.slots = [...roster.slots];
  }

  // --- Drag and Drop handlers ---

  function handleSlotDragStart(e: DragEvent, index: number) {
    draggedSlotIndex = index;
    draggedLibraryItem = null;
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move';
      e.dataTransfer.setData('text/plain', `slot:${index}`);
    }
  }

  function handleLibraryDragStart(e: DragEvent, type: 'char' | 'vault', name: string, asset?: any) {
    draggedSlotIndex = null;
    draggedLibraryItem = { type, name, asset };
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'copy';
      e.dataTransfer.setData('text/plain', `lib:${name}`);
    }
  }

  function handleSlotDragOver(e: DragEvent) {
    e.preventDefault();
    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = draggedSlotIndex !== null ? 'move' : 'copy';
    }
  }

  function handleSlotDrop(e: DragEvent, targetIndex: number) {
    e.preventDefault();
    if (!roster) return;

    if (draggedSlotIndex !== null && draggedSlotIndex !== targetIndex) {
      // Swap two slots
      const temp = { ...roster.slots[draggedSlotIndex], index: targetIndex };
      roster.slots[draggedSlotIndex] = { ...roster.slots[targetIndex], index: draggedSlotIndex };
      roster.slots[targetIndex] = temp;
      roster.slots = [...roster.slots];
      selectedSlotIndex = targetIndex;
    } else if (draggedLibraryItem) {
      // Dropping from library onto slot
      if (draggedLibraryItem.type === 'char') {
        const char = roster.available_characters.find((c) => c.name === draggedLibraryItem?.name);
        roster.slots[targetIndex] = {
          index: targetIndex,
          type: 'character',
          character: draggedLibraryItem.name,
          display_name: char?.display_name || draggedLibraryItem.name,
          author: char?.author || 'Unknown',
          portrait_base64: char?.portrait_base64 || '',
          include_in_arcade: true,
          order: 1,
        };
      } else if (draggedLibraryItem.type === 'vault' && draggedLibraryItem.asset) {
        const asset = draggedLibraryItem.asset as VaultAsset;
        const charName = getAssetName(asset.key);
        if ($projectStore.current) {
          const vaultId = (asset as any).vault_id || $vaultStore.activeVaultId || 'vault-default';
          vaultStore.linkToProject($projectStore.current.path, vaultId, asset.key);
        }
        roster.slots[targetIndex] = {
          index: targetIndex,
          type: 'character',
          character: charName,
          display_name: asset.display_name || charName,
          author: asset.author || 'Unknown',
          portrait_base64: asset.preview_base64 || '',
          include_in_arcade: true,
          order: 1,
        };
      }
      roster.slots = [...roster.slots];
      selectedSlotIndex = targetIndex;
    }

    draggedSlotIndex = null;
    draggedLibraryItem = null;
  }

  function addExtraStage(stagePath: string) {
    if (!roster) return;
    if (!roster.extra_stages.includes(stagePath)) {
      roster.extra_stages.push(stagePath);
      roster.extra_stages = [...roster.extra_stages];
    }
  }

  function removeExtraStage(index: number) {
    if (!roster) return;
    roster.extra_stages.splice(index, 1);
    roster.extra_stages = [...roster.extra_stages];
  }
</script>

<div class="h-full flex flex-col bg-dark-900 text-slate-100 overflow-hidden select-none">
  <!-- Top Navigation & Actions Bar -->
  <header class="p-4 bg-dark-850 border-b border-dark-600/60 flex items-center justify-between gap-4 flex-shrink-0 z-10">
    <div class="flex items-center gap-3">
      <button
        type="button"
        class="p-2 rounded-xl bg-dark-800 hover:bg-dark-700 border border-dark-600/80 text-slate-300 hover:text-white transition flex items-center gap-1.5 text-xs font-semibold"
        on:click={onBackToWorkspace}
      >
        <ArrowLeft class="w-4 h-4" />
        <span>Workspace</span>
      </button>

      <div class="h-5 w-px bg-dark-600/60"></div>

      <div class="flex items-center gap-2.5">
        <div class="w-8 h-8 rounded-xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center text-white shadow-md shadow-indigo-500/20">
          <Users class="w-4 h-4" />
        </div>
        <div>
          <h1 class="text-sm font-bold text-slate-100 flex items-center gap-2">
            <span>Roster & Select Screen Manager</span>
            <span class="text-[10px] font-mono px-2 py-0.5 rounded-full bg-indigo-500/10 text-indigo-400 border border-indigo-500/30">
              select.def
            </span>
          </h1>
          <p class="text-[11px] text-slate-400">
            {roster?.slots.length || 0} Character Slots &bull; {roster?.extra_stages.length || 0} Extra Stages
          </p>
        </div>
      </div>
    </div>

    <!-- Center Navigation Tabs -->
    <div class="flex items-center bg-dark-800 p-1 rounded-xl border border-dark-600/60 text-xs">
      <button
        type="button"
        class="flex items-center gap-2 px-3 py-1.5 rounded-lg font-semibold transition {
          activeView === 'grid' ? 'bg-indigo-600 text-white shadow' : 'text-slate-400 hover:text-slate-200'
        }"
        on:click={() => (activeView = 'grid')}
      >
        <Grid class="w-3.5 h-3.5" />
        <span>Fighter Grid</span>
      </button>

      <button
        type="button"
        class="flex items-center gap-2 px-3 py-1.5 rounded-lg font-semibold transition {
          activeView === 'extra_stages' ? 'bg-indigo-600 text-white shadow' : 'text-slate-400 hover:text-slate-200'
        }"
        on:click={() => (activeView = 'extra_stages')}
      >
        <Mountain class="w-3.5 h-3.5" />
        <span>Extra Stages ({roster?.extra_stages.length || 0})</span>
      </button>
    </div>

    <!-- Right Controls: Sync to Vault & Save -->
    <div class="flex items-center gap-2.5">
      <button
        type="button"
        disabled={exportingToVault}
        class="flex items-center gap-2 px-3.5 py-2 rounded-xl bg-dark-800 hover:bg-dark-700 border border-dark-600/80 text-brand-300 text-xs font-semibold transition disabled:opacity-50"
        on:click={handleExportToVault}
        title="Extract and index all project fighters and stages into the Vault"
      >
        <DownloadCloud class="w-3.5 h-3.5 text-brand-400" />
        <span>{exportingToVault ? 'Exporting...' : 'Sync All to Vault'}</span>
      </button>

      <button
        type="button"
        disabled={saving || loading}
        class="flex items-center gap-2 px-5 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold shadow-lg shadow-indigo-950/50 transition disabled:opacity-50"
        on:click={handleSave}
      >
        <Save class="w-3.5 h-3.5" />
        <span>{saving ? 'Saving...' : 'Save select.def'}</span>
      </button>
    </div>
  </header>

  <!-- Main Work Area (3 Columns: Library Sidebar, Visual Grid, Slot Inspector) -->
  <div class="flex-1 flex min-h-0 overflow-hidden">
    <!-- 1. Left Library Sidebar (Fighter / Stage Picker) -->
    <aside class="w-72 bg-dark-850 border-r border-dark-600/60 flex flex-col flex-shrink-0 z-10">
      <!-- Library Header & Tabs -->
      <div class="p-3 border-b border-dark-600/40 space-y-2.5">
        <div class="flex items-center justify-between text-xs">
          <span class="font-bold text-[11px] uppercase tracking-wider text-slate-400">Asset Library</span>
          <span class="text-[10px] font-mono text-slate-500">Drag to place</span>
        </div>

        <div class="grid grid-cols-3 gap-1 bg-dark-800 p-1 rounded-xl border border-dark-600/60 text-xs">
          <button
            type="button"
            class="py-1.5 px-2 rounded-lg font-semibold text-center transition {
              libraryTab === 'project' ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:text-slate-200'
            }"
            on:click={() => (libraryTab = 'project')}
          >
            Project ({roster?.available_characters.length || 0})
          </button>
          <button
            type="button"
            class="py-1.5 px-2 rounded-lg font-semibold text-center transition {
              libraryTab === 'vault' ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:text-slate-200'
            }"
            on:click={() => (libraryTab = 'vault')}
          >
            Vault
          </button>
          <button
            type="button"
            class="py-1.5 px-2 rounded-lg font-semibold text-center transition {
              libraryTab === 'stages' ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:text-slate-200'
            }"
            on:click={() => (libraryTab = 'stages')}
          >
            Stages
          </button>
        </div>

        <!-- Search Input -->
        <div class="relative">
          <Search class="w-3.5 h-3.5 text-slate-500 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            bind:value={librarySearch}
            placeholder="Search fighters..."
            class="w-full bg-dark-900 border border-dark-600/70 rounded-xl pl-9 pr-3 py-1.5 text-xs text-slate-200 placeholder:text-slate-500 focus:outline-none focus:border-indigo-500"
          />
        </div>
      </div>

      <!-- Library List Content -->
      <div class="flex-1 overflow-y-auto p-2.5 space-y-1.5">
        {#if libraryTab === 'project'}
          {#if filteredProjectChars.length === 0}
            <div class="p-6 text-center text-xs text-slate-500">
              No project characters found matching filter.
            </div>
          {:else}
            {#each filteredProjectChars as char}
              <div
                class="group p-2 rounded-xl bg-dark-800 hover:bg-dark-750 border border-dark-600/60 hover:border-indigo-500/40 flex items-center justify-between gap-2.5 transition cursor-grab active:cursor-grabbing shadow-sm"
                draggable="true"
                on:dragstart={(e) => handleLibraryDragStart(e, 'char', char.name)}
              >
                <div class="flex items-center gap-2.5 min-w-0">
                  <div class="w-8 h-8 rounded-lg bg-dark-900 border border-dark-600/80 overflow-hidden flex items-center justify-center flex-shrink-0 bg-gradient-to-br from-indigo-950/40 to-dark-900">
                    {#if char.portrait_base64}
                      <img src={char.portrait_base64} alt={char.name} class="w-full h-full object-contain" />
                    {:else}
                      <Users class="w-4 h-4 text-slate-500" />
                    {/if}
                  </div>
                  <div class="min-w-0">
                    <div class="text-xs font-bold text-slate-200 truncate group-hover:text-indigo-300 transition">
                      {char.display_name}
                    </div>
                    <div class="text-[10px] text-slate-400 truncate">
                      {char.author}
                    </div>
                  </div>
                </div>

                <button
                  type="button"
                  class="p-1.5 rounded-lg bg-dark-700 hover:bg-indigo-600 text-slate-300 hover:text-white transition flex-shrink-0"
                  title="Add to next slot"
                  on:click={() => addCharacterToNextSlot(char.name, char.display_name, char.author, char.portrait_base64)}
                >
                  <Plus class="w-3.5 h-3.5" />
                </button>
              </div>
            {/each}
          {/if}
        {:else if libraryTab === 'vault'}
          {#if filteredVaultAssets.length === 0}
            <div class="p-6 text-center text-xs text-slate-500">
              No fighters found in Vault. Ingest characters to populate your central library.
            </div>
          {:else}
            {#each filteredVaultAssets as asset}
              {@const charName = getAssetName(asset.key)}
              <div
                class="group p-2 rounded-xl bg-dark-800 hover:bg-dark-750 border border-dark-600/60 hover:border-brand-500/40 flex items-center justify-between gap-2.5 transition cursor-grab active:cursor-grabbing shadow-sm"
                draggable="true"
                on:dragstart={(e) => handleLibraryDragStart(e, 'vault', charName, asset)}
              >
                <div class="flex items-center gap-2.5 min-w-0">
                  <div class="w-8 h-8 rounded-lg bg-dark-900 border border-dark-600/80 overflow-hidden flex items-center justify-center flex-shrink-0">
                    {#if asset.preview_base64}
                      <img src={asset.preview_base64} alt={charName} class="w-full h-full object-contain" />
                    {:else}
                      <Sparkles class="w-4 h-4 text-brand-400" />
                    {/if}
                  </div>
                  <div class="min-w-0">
                    <div class="text-xs font-bold text-slate-200 truncate group-hover:text-brand-300 transition">
                      {asset.display_name || charName}
                    </div>
                    <div class="text-[10px] text-slate-400 truncate">
                      {asset.author || 'Vault Asset'}
                    </div>
                  </div>
                </div>

                <button
                  type="button"
                  class="p-1.5 rounded-lg bg-dark-700 hover:bg-brand-600 text-slate-300 hover:text-white transition flex-shrink-0"
                  title="Link from Vault to Roster"
                  on:click={() => handleAddVaultCharToProject(asset)}
                >
                  <FolderPlus class="w-3.5 h-3.5" />
                </button>
              </div>
            {/each}
          {/if}
        {:else if libraryTab === 'stages'}
          {#if filteredAvailableStages.length === 0}
            <div class="p-6 text-center text-xs text-slate-500">
              No stages found in stages/ folder.
            </div>
          {:else}
            {#each filteredAvailableStages as stage}
              <div
                class="group p-2 rounded-xl bg-dark-800 hover:bg-dark-750 border border-dark-600/60 hover:border-emerald-500/40 flex items-center justify-between gap-2.5 transition shadow-sm"
              >
                <div class="flex items-center gap-2 min-w-0">
                  <Mountain class="w-4 h-4 text-emerald-400 flex-shrink-0" />
                  <span class="text-xs font-medium text-slate-300 truncate group-hover:text-emerald-300">
                    {stage}
                  </span>
                </div>
                <button
                  type="button"
                  class="p-1.5 rounded-lg bg-dark-700 hover:bg-emerald-600 text-slate-300 hover:text-white transition flex-shrink-0"
                  title="Add to Extra Stages"
                  on:click={() => addExtraStage(stage)}
                >
                  <Plus class="w-3.5 h-3.5" />
                </button>
              </div>
            {/each}
          {/if}
        {/if}
      </div>
    </aside>

    <!-- 2. Middle Center Work Area (Interactive Grid or Extra Stages) -->
    <main class="flex-1 flex flex-col min-w-0 bg-dark-900 overflow-y-auto p-6">
      {#if loading}
        <div class="flex-1 flex items-center justify-center">
          <div class="text-xs text-slate-400 animate-pulse">Loading select.def roster...</div>
        </div>
      {:else if activeView === 'grid' && roster}
        <!-- Grid Toolbar -->
        <div class="flex items-center justify-between pb-4 mb-4 border-b border-dark-700/60">
          <div class="flex items-center gap-3">
            <span class="text-xs font-bold uppercase tracking-wider text-slate-400">Select Screen Grid</span>
            <div class="flex items-center gap-1.5 bg-dark-800 px-3 py-1 rounded-xl border border-dark-600/60 text-xs">
              <span class="text-slate-400">Cols:</span>
              <button
                type="button"
                class="px-1.5 py-0.5 rounded hover:bg-dark-700 font-bold"
                on:click={() => (gridColumns = Math.max(1, gridColumns - 1))}
              >-</button>
              <span class="font-mono text-indigo-400 font-bold px-1">{gridColumns}</span>
              <button
                type="button"
                class="px-1.5 py-0.5 rounded hover:bg-dark-700 font-bold"
                on:click={() => (gridColumns = gridColumns + 1)}
              >+</button>
            </div>
          </div>

          <div class="flex items-center gap-2">
            <button
              type="button"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-dark-800 hover:bg-dark-700 border border-dark-600/60 text-xs font-semibold text-slate-300 transition"
              on:click={addEmptySlot}
            >
              <Plus class="w-3.5 h-3.5 text-indigo-400" />
              <span>Empty Slot</span>
            </button>

            <button
              type="button"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-dark-800 hover:bg-dark-700 border border-dark-600/60 text-xs font-semibold text-purple-300 transition"
              on:click={addRandomSelectSlot}
            >
              <HelpCircle class="w-3.5 h-3.5 text-purple-400" />
              <span>Random Select</span>
            </button>
          </div>
        </div>

        <!-- The Interactive Grid -->
        <div
          class="grid gap-3 flex-1 auto-rows-max"
          style="grid-template-columns: repeat({gridColumns}, minmax(0, 1fr));"
        >
          {#each roster.slots as slot, idx}
            <!-- svelte-ignore a11y_click_to_play -->
            <div
              class="group relative rounded-2xl border transition-all duration-150 cursor-pointer aspect-square flex flex-col justify-between p-2.5 {
                selectedSlotIndex === idx
                  ? 'border-indigo-500 bg-indigo-950/30 ring-2 ring-indigo-500/40 shadow-lg shadow-indigo-950/40'
                  : 'border-dark-600/70 bg-dark-800/80 hover:bg-dark-800 hover:border-slate-500/60'
              } {
                slot.type === 'disabled' ? 'opacity-50 grayscale' : ''
              }"
              draggable="true"
              on:click={() => (selectedSlotIndex = idx)}
              on:dragstart={(e) => handleSlotDragStart(e, idx)}
              on:dragover={handleSlotDragOver}
              on:drop={(e) => handleSlotDrop(e, idx)}
            >
              <!-- Slot Top Index & Type Badge -->
              <div class="flex items-center justify-between w-full text-[10px] text-slate-400">
                <span class="font-mono font-bold">#{idx + 1}</span>

                {#if slot.type === 'randomselect'}
                  <span class="px-1.5 py-0.5 rounded bg-purple-500/20 text-purple-300 border border-purple-500/30 font-bold text-[9px]">
                    RANDOM
                  </span>
                {:else if slot.type === 'disabled'}
                  <span class="px-1.5 py-0.5 rounded bg-rose-500/20 text-rose-300 border border-rose-500/30 font-bold text-[9px]">
                    DISABLED
                  </span>
                {:else if slot.type === 'empty'}
                  <span class="px-1.5 py-0.5 rounded bg-dark-700 text-slate-500 font-mono text-[9px]">
                    EMPTY
                  </span>
                {:else if slot.order && slot.order > 0}
                  <span class="px-1.5 py-0.5 rounded bg-amber-500/20 text-amber-300 border border-amber-500/30 font-bold text-[9px] flex items-center gap-1">
                    <Crown class="w-2.5 h-2.5" />
                    <span>Boss {slot.order}</span>
                  </span>
                {/if}
              </div>

              <!-- Slot Center: Thumbnail / Icon -->
              <div class="flex-1 flex items-center justify-center my-1 overflow-hidden">
                {#if slot.type === 'character' || slot.type === 'disabled'}
                  {#if slot.portrait_base64}
                    <img
                      src={slot.portrait_base64}
                      alt={slot.character}
                      class="max-h-full max-w-full object-contain rounded drop-shadow-md group-hover:scale-105 transition duration-150"
                    />
                  {:else}
                    <div class="w-12 h-12 rounded-xl bg-indigo-950/30 border border-indigo-500/20 flex items-center justify-center text-indigo-400">
                      <Users class="w-6 h-6" />
                    </div>
                  {/if}
                {:else if slot.type === 'randomselect'}
                  <div class="w-12 h-12 rounded-2xl bg-gradient-to-br from-purple-500/20 to-indigo-500/20 border border-purple-500/30 flex items-center justify-center text-purple-400 shadow-inner">
                    <HelpCircle class="w-7 h-7" />
                  </div>
                {:else}
                  <div class="w-10 h-10 rounded-xl border border-dashed border-dark-600 flex items-center justify-center text-slate-600 group-hover:text-slate-400 group-hover:border-slate-500 transition">
                    <Plus class="w-5 h-5" />
                  </div>
                {/if}
              </div>

              <!-- Slot Bottom: Character Name -->
              <div class="w-full text-center">
                {#if slot.type === 'character' || slot.type === 'disabled'}
                  <div class="text-xs font-bold text-slate-100 truncate">
                    {slot.display_name || slot.character}
                  </div>
                  {#if slot.home_stage}
                    <div class="text-[9px] text-emerald-400 truncate flex items-center justify-center gap-1">
                      <Mountain class="w-2.5 h-2.5 flex-shrink-0" />
                      <span>{slot.home_stage.replace(/^stages\//i, '').replace(/\.def$/i, '')}</span>
                    </div>
                  {/if}
                {:else if slot.type === 'randomselect'}
                  <div class="text-[11px] font-bold text-purple-300">Random Select</div>
                {:else}
                  <div class="text-[10px] text-slate-500">Drop fighter here</div>
                {/if}
              </div>
            </div>
          {/each}
        </div>
      {:else if activeView === 'extra_stages' && roster}
        <!-- Extra Stages View -->
        <div class="space-y-4 max-w-2xl">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-sm font-bold text-slate-100">Extra Stage Roster ([ExtraStages])</h2>
              <p class="text-xs text-slate-400">Stages selectable in Versus, Training, and Survival modes.</p>
            </div>
          </div>

          <div class="space-y-2">
            {#each roster.extra_stages as stage, sIdx}
              <div class="p-3 rounded-xl bg-dark-800 border border-dark-600/70 flex items-center justify-between gap-3">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-lg bg-emerald-950/40 border border-emerald-500/30 flex items-center justify-center text-emerald-400">
                    <Mountain class="w-4 h-4" />
                  </div>
                  <div>
                    <div class="text-xs font-bold text-slate-200">{stage}</div>
                    <div class="text-[10px] text-slate-500 font-mono">Stage #{sIdx + 1}</div>
                  </div>
                </div>

                <button
                  type="button"
                  class="p-2 rounded-lg text-slate-500 hover:text-rose-400 hover:bg-rose-500/10 transition"
                  on:click={() => removeExtraStage(sIdx)}
                >
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>
            {/each}

            {#if roster.extra_stages.length === 0}
              <div class="p-8 text-center rounded-xl bg-dark-800/40 border border-dashed border-dark-600 text-xs text-slate-500">
                No extra stages registered. Pick stages from the left library to add them.
              </div>
            {/if}
          </div>
        </div>
      {/if}
    </main>

    <!-- 3. Right Slot Inspector Drawer -->
    {#if selectedSlot && roster}
      <aside class="w-80 bg-dark-850 border-l border-dark-600/60 p-5 flex flex-col justify-between overflow-y-auto z-10">
        <div class="space-y-5">
          <!-- Inspector Header -->
          <div class="flex items-center justify-between pb-3 border-b border-dark-600/60">
            <div class="flex items-center gap-2">
              <Sliders class="w-4 h-4 text-indigo-400" />
              <span class="text-xs font-bold uppercase tracking-wider text-slate-200">
                Slot #{selectedSlot.index + 1} Inspector
              </span>
            </div>
            <button
              type="button"
              class="p-1 rounded text-slate-500 hover:text-slate-300"
              on:click={() => (selectedSlotIndex = null)}
            >
              <X class="w-4 h-4" />
            </button>
          </div>

          <!-- Slot Type Switcher -->
          <div class="space-y-1.5">
            <label class="block text-[11px] font-bold uppercase tracking-wider text-slate-400">Slot Type</label>
            <select
              bind:value={selectedSlot.type}
              class="w-full bg-dark-800 border border-dark-600 rounded-xl px-3 py-2 text-xs text-slate-100 font-semibold focus:outline-none focus:border-indigo-500"
            >
              <option value="character">Character</option>
              <option value="randomselect">Random Select (?)</option>
              <option value="empty">Empty Slot</option>
              <option value="disabled">Disabled (Commented)</option>
            </select>
          </div>

          <!-- If Character: Details & Config -->
          {#if selectedSlot.type === 'character' || selectedSlot.type === 'disabled'}
            <div class="space-y-4">
              <!-- Portrait Thumbnail Preview -->
              <div class="p-3 rounded-2xl bg-dark-900 border border-dark-600/80 flex items-center gap-3">
                <div class="w-14 h-14 rounded-xl bg-dark-800 border border-dark-600 flex items-center justify-center overflow-hidden flex-shrink-0">
                  {#if selectedSlot.portrait_base64}
                    <img src={selectedSlot.portrait_base64} alt={selectedSlot.character} class="w-full h-full object-contain" />
                  {:else}
                    <Users class="w-6 h-6 text-slate-500" />
                  {/if}
                </div>
                <div class="min-w-0">
                  <div class="text-xs font-bold text-slate-100 truncate">{selectedSlot.display_name || selectedSlot.character}</div>
                  <div class="text-[11px] text-slate-400 truncate">chars/{selectedSlot.character}</div>
                  <div class="text-[10px] text-indigo-400 truncate">{selectedSlot.author}</div>
                </div>
              </div>

              <!-- Home Stage Selector -->
              <div class="space-y-1.5">
                <label class="block text-[11px] font-bold uppercase tracking-wider text-slate-400 flex items-center gap-1.5">
                  <Mountain class="w-3.5 h-3.5 text-emerald-400" />
                  <span>Assigned Home Stage</span>
                </label>
                <select
                  bind:value={selectedSlot.home_stage}
                  class="w-full bg-dark-800 border border-dark-600 rounded-xl px-3 py-2 text-xs text-slate-200 focus:outline-none focus:border-emerald-500"
                >
                  <option value="">Default (No specific stage)</option>
                  {#each roster.available_stages as st}
                    <option value={st}>{st}</option>
                  {/each}
                </select>
              </div>

              <!-- Custom BGM -->
              <div class="space-y-1.5">
                <label class="block text-[11px] font-bold uppercase tracking-wider text-slate-400 flex items-center gap-1.5">
                  <Volume2 class="w-3.5 h-3.5 text-cyan-400" />
                  <span>Custom BGM Track</span>
                </label>
                <input
                  type="text"
                  bind:value={selectedSlot.music}
                  placeholder="sound/my_theme.mp3"
                  class="w-full bg-dark-800 border border-dark-600 rounded-xl px-3 py-2 text-xs text-slate-200 placeholder:text-slate-600 focus:outline-none focus:border-cyan-500"
                />
              </div>

              <!-- Arcade Boss Order -->
              <div class="space-y-1.5">
                <label class="block text-[11px] font-bold uppercase tracking-wider text-slate-400 flex items-center gap-1.5">
                  <Crown class="w-3.5 h-3.5 text-amber-400" />
                  <span>Arcade Order (Boss Priority)</span>
                </label>
                <div class="flex items-center gap-2">
                  <input
                    type="number"
                    min="0"
                    max="10"
                    bind:value={selectedSlot.order}
                    class="w-20 bg-dark-800 border border-dark-600 rounded-xl px-3 py-2 text-xs text-slate-200 font-mono focus:outline-none focus:border-amber-500"
                  />
                  <span class="text-[11px] text-slate-500">
                    {selectedSlot.order === 0 ? 'Normal Fighter' : `Boss Tier ${selectedSlot.order}`}
                  </span>
                </div>
              </div>

              <!-- Disabled / Enabled Toggle -->
              <div class="pt-2">
                <button
                  type="button"
                  class="w-full py-2 px-3 rounded-xl border text-xs font-semibold flex items-center justify-center gap-2 transition {
                    selectedSlot.type === 'disabled'
                      ? 'bg-rose-500/10 border-rose-500/30 text-rose-400 hover:bg-rose-500/20'
                      : 'bg-dark-800 border-dark-600 text-slate-300 hover:bg-dark-700'
                  }"
                  on:click={() => selectedSlotIndex !== null && toggleSlotDisabled(selectedSlotIndex)}
                >
                  <EyeOff class="w-3.5 h-3.5" />
                  <span>{selectedSlot.type === 'disabled' ? 'Enable Character' : 'Disable Character (Keep in list)'}</span>
                </button>
              </div>
            </div>
          {/if}
        </div>

        <!-- Slot Danger Zone Actions -->
        <div class="pt-4 border-t border-dark-600/60 space-y-2">
          <button
            type="button"
            class="w-full py-2 px-3 rounded-xl bg-dark-800 hover:bg-dark-700 border border-dark-600 text-slate-300 text-xs font-semibold transition"
            on:click={() => selectedSlotIndex !== null && clearSlot(selectedSlotIndex)}
          >
            Clear Slot to Empty
          </button>

          <button
            type="button"
            class="w-full py-2 px-3 rounded-xl bg-rose-600/10 hover:bg-rose-600/20 border border-rose-500/30 text-rose-400 text-xs font-semibold flex items-center justify-center gap-2 transition"
            on:click={() => selectedSlotIndex !== null && deleteSlot(selectedSlotIndex)}
          >
            <Trash2 class="w-3.5 h-3.5" />
            <span>Delete Slot Completely</span>
          </button>
        </div>
      </aside>
    {/if}
  </div>
</div>
