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
    Shuffle,
    ArrowUpDown,
    CheckSquare,
    Square,
    Copy,
    ListFilter,
  } from 'lucide-svelte';

  export let onBackToWorkspace: () => void;

  let loading = true;
  let saving = false;
  let exportingToVault = false;
  let roster: ProjectRoster | null = null;
  let selectedSlotIndex: number | null = null;

  // Grid dimension overrides (synced from system.def)
  let gridColumns = 8;
  let gridRows = 10;
  let density: 'arcade' | 'medium' | 'large' = 'arcade';

  // Hovered slot for arcade live preview
  let hoveredSlotIndex: number | null = null;

  // Sidebar library state
  let libraryTab: 'project' | 'vault' | 'stages' = 'project';
  let librarySearch = '';
  let vaultFilter = 'all';

  // Drag and drop state
  let draggedSlotIndex: number | null = null;
  let draggedLibraryItem: { type: 'char' | 'vault'; name: string; asset?: any } | null = null;

  // Active subtab in roster view
  let activeView: 'grid' | 'extra_stages' = 'grid';

  // Multi-selection state for Library sidebar
  let selectedLibraryChars = new Set<string>();

  // Multi-selection state for Grid slots
  let selectedGridSlots = new Set<number>();
  let isGridSelectMode = false;

  // Right-click context menu
  let contextMenu: { x: number; y: number; slotIndex: number } | null = null;

  function handleSlotRightClick(e: MouseEvent, idx: number) {
    e.preventDefault();
    ensureSlotCapacity(idx);
    roster && (roster.slots = [...roster.slots]);
    selectedSlotIndex = idx;
    selectedGridSlots = new Set();
    contextMenu = { x: e.clientX, y: e.clientY, slotIndex: idx };
  }

  function closeContextMenu() {
    contextMenu = null;
  }

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
      gridColumns = Math.max(1, data.grid?.columns || 8);
      gridRows = Math.max(1, data.grid?.rows || 10);
      if (data.slots.length > 0) {
        selectedSlotIndex = 0;
      }
    }
    loading = false;
  }

  $: totalGridCells = Math.max(gridColumns * gridRows, roster?.slots.length || 0);

  // Character frequency map to detect duplicates
  $: characterCounts = (() => {
    const counts: Record<string, number> = {};
    if (!roster) return counts;
    for (const slot of roster.slots) {
      if (slot.character && slot.type !== 'empty' && slot.type !== 'randomselect') {
        const key = slot.character.toLowerCase();
        counts[key] = (counts[key] || 0) + 1;
      }
    }
    return counts;
  })();

  function getSlotAt(idx: number): RosterCharacterSlot {
    if (roster && idx < roster.slots.length) {
      return roster.slots[idx];
    }
    return {
      index: idx,
      type: 'empty',
      character: '',
      include_in_arcade: true,
    };
  }

  function ensureSlotCapacity(targetIdx: number) {
    if (!roster) return;
    while (roster.slots.length <= targetIdx) {
      roster.slots.push({
        index: roster.slots.length,
        type: 'empty',
        character: '',
        include_in_arcade: true,
      });
    }
  }

  // --- Bulk Library Actions ---
  function toggleLibraryCharSelect(charName: string) {
    if (selectedLibraryChars.has(charName)) {
      selectedLibraryChars.delete(charName);
    } else {
      selectedLibraryChars.add(charName);
    }
    selectedLibraryChars = new Set(selectedLibraryChars);
  }

  function selectAllLibraryChars() {
    if (libraryTab === 'project') {
      selectedLibraryChars = new Set(filteredProjectChars.map((c) => c.name));
    } else if (libraryTab === 'vault') {
      selectedLibraryChars = new Set(filteredVaultAssets.map((a) => getAssetName(a.key)));
    }
  }

  function deselectAllLibraryChars() {
    selectedLibraryChars = new Set();
  }

  function addSelectedLibraryCharsToRoster() {
    if (!roster || selectedLibraryChars.size === 0) return;

    for (const name of selectedLibraryChars) {
      if (libraryTab === 'project') {
        const char = roster.available_characters.find((c) => c.name === name);
        if (char) {
          addCharacterToNextSlot(char.name, char.display_name, char.author, char.portrait_base64);
        }
      } else if (libraryTab === 'vault') {
        const asset = $vaultStore.assets.find((a) => getAssetName(a.key) === name);
        if (asset) {
          handleAddVaultCharToProject(asset);
        }
      }
    }
    deselectAllLibraryChars();
  }

  // --- Roster Grid Actions ---
  function autoPopulateAllProjectCharacters() {
    if (!roster || !roster.available_characters) return;
    if (roster.available_characters.length === 0) return;

    roster.slots = roster.available_characters.map((char, i) => ({
      index: i,
      type: 'character' as const,
      character: char.name,
      display_name: char.display_name || char.name,
      author: char.author || 'Unknown',
      portrait_base64: char.portrait_base64 || '',
      include_in_arcade: true,
      order: 1,
    }));

    gridRows = Math.max(1, Math.ceil(roster.slots.length / gridColumns));
    selectedSlotIndex = 0;
    roster.slots = [...roster.slots];
  }

  function clearEntireRoster() {
    if (!roster) return;
    roster.slots = [];
    selectedSlotIndex = null;
    selectedGridSlots = new Set();
    gridRows = Math.max(1, roster.grid?.rows || 10);
  }

  function shuffleRosterLayout() {
    if (!roster || roster.slots.length <= 1) return;
    // Extract non-empty slots
    const characterSlots = roster.slots.filter((s) => s.type === 'character' || s.type === 'randomselect' || s.type === 'disabled');
    // Fisher-Yates shuffle
    for (let i = characterSlots.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [characterSlots[i], characterSlots[j]] = [characterSlots[j], characterSlots[i]];
    }

    // Place back into slots
    let charIdx = 0;
    for (let i = 0; i < roster.slots.length; i++) {
      if (roster.slots[i].type !== 'empty') {
        const item = characterSlots[charIdx++];
        roster.slots[i] = { ...item, index: i };
      }
    }
    roster.slots = [...roster.slots];
  }

  function sortRosterBy(criteria: 'name_asc' | 'name_desc' | 'author' | 'order') {
    if (!roster || roster.slots.length <= 1) return;
    const activeSlots = roster.slots.filter((s) => s.type !== 'empty');
    activeSlots.sort((a, b) => {
      if (criteria === 'name_asc') {
        return (a.display_name || a.character || '').localeCompare(b.display_name || b.character || '');
      }
      if (criteria === 'name_desc') {
        return (b.display_name || b.character || '').localeCompare(a.display_name || a.character || '');
      }
      if (criteria === 'author') {
        return (a.author || '').localeCompare(b.author || '');
      }
      if (criteria === 'order') {
        return (a.order || 0) - (b.order || 0);
      }
      return 0;
    });

    let activeIdx = 0;
    for (let i = 0; i < roster.slots.length; i++) {
      if (roster.slots[i].type !== 'empty') {
        const item = activeSlots[activeIdx++];
        roster.slots[i] = { ...item, index: i };
      }
    }
    roster.slots = [...roster.slots];
  }

  // --- Grid Multi-Select actions ---
  function toggleGridSlotSelect(idx: number) {
    if (selectedGridSlots.has(idx)) {
      selectedGridSlots.delete(idx);
    } else {
      selectedGridSlots.add(idx);
    }
    selectedGridSlots = new Set(selectedGridSlots);
  }

  function bulkClearSelectedGridSlots() {
    if (!roster || selectedGridSlots.size === 0) return;
    for (const idx of selectedGridSlots) {
      if (roster.slots[idx]) {
        roster.slots[idx] = {
          index: idx,
          type: 'empty',
          character: '',
          include_in_arcade: true,
        };
      }
    }
    roster.slots = [...roster.slots];
    selectedGridSlots = new Set();
  }

  function bulkDisableSelectedGridSlots() {
    if (!roster || selectedGridSlots.size === 0) return;
    for (const idx of selectedGridSlots) {
      if (roster.slots[idx] && roster.slots[idx].type === 'character') {
        roster.slots[idx].type = 'disabled';
      }
    }
    roster.slots = [...roster.slots];
    selectedGridSlots = new Set();
  }

  function fillRemainingWithRandom() {
    if (!roster) return;
    const maxCapacity = gridColumns * gridRows;
    ensureSlotCapacity(maxCapacity - 1);
    for (let i = 0; i < maxCapacity; i++) {
      if (roster.slots[i].type === 'empty') {
        roster.slots[i] = {
          index: i,
          type: 'randomselect',
          character: 'randomselect',
          display_name: 'Random Select',
          include_in_arcade: true,
        };
      }
    }
    roster.slots = [...roster.slots];
  }

  function trimTrailingEmptySlots() {
    if (!roster) return;
    while (roster.slots.length > 0 && roster.slots[roster.slots.length - 1].type === 'empty') {
      roster.slots.pop();
    }
    gridRows = Math.max(1, Math.ceil(roster.slots.length / gridColumns));
    roster.slots = [...roster.slots];
  }

  function matchMotifGrid() {
    if (!roster) return;
    gridColumns = Math.max(1, roster.grid?.columns || 8);
    gridRows = Math.max(1, roster.grid?.rows || 10);
  }

  function selectOrActivateSlot(idx: number, e?: MouseEvent) {
    ensureSlotCapacity(idx);
    roster && (roster.slots = [...roster.slots]);
    if (isGridSelectMode || (e && (e.shiftKey || e.ctrlKey || e.metaKey))) {
      toggleGridSlotSelect(idx);
      selectedSlotIndex = idx;
    } else {
      selectedSlotIndex = idx;
      selectedGridSlots = new Set();
    }
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

  $: selectedSlot = roster && selectedSlotIndex !== null && selectedSlotIndex < roster.slots.length
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
    if (!roster) return;
    ensureSlotCapacity(index);
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
          {#if libraryTab !== 'stages'}
            <button
              type="button"
              class="text-[10px] text-indigo-400 hover:text-indigo-300 font-semibold flex items-center gap-1"
              on:click={() => {
                if (selectedLibraryChars.size > 0) deselectAllLibraryChars();
                else selectAllLibraryChars();
              }}
            >
              {#if selectedLibraryChars.size > 0}
                <CheckSquare class="w-3 h-3" />
                <span>Clear ({selectedLibraryChars.size})</span>
              {:else}
                <Square class="w-3 h-3" />
                <span>Select All</span>
              {/if}
            </button>
          {/if}
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
              {@const isSelected = selectedLibraryChars.has(char.name)}
              <div
                class="group p-2 rounded-xl border flex items-center justify-between gap-2.5 transition cursor-grab active:cursor-grabbing shadow-sm {
                  isSelected
                    ? 'bg-indigo-950/40 border-indigo-500/80 shadow-indigo-950/40'
                    : 'bg-dark-800 hover:bg-dark-750 border-dark-600/60 hover:border-indigo-500/40'
                }"
                draggable="true"
                on:dragstart={(e) => handleLibraryDragStart(e, 'char', char.name)}
              >
                <!-- Selection Checkbox & Preview -->
                <div class="flex items-center gap-2 min-w-0">
                  <button
                    type="button"
                    class="p-0.5 rounded text-slate-500 hover:text-indigo-400"
                    on:click|stopPropagation={() => toggleLibraryCharSelect(char.name)}
                  >
                    {#if isSelected}
                      <CheckSquare class="w-4 h-4 text-indigo-400" />
                    {:else}
                      <Square class="w-4 h-4 text-dark-500 group-hover:text-slate-400" />
                    {/if}
                  </button>

                  <div class="w-10 h-10 rounded-xl bg-dark-900 border border-dark-600/80 overflow-hidden flex items-center justify-center flex-shrink-0 bg-gradient-to-br from-indigo-950/40 to-dark-900 shadow-inner relative">
                    {#if char.portrait_base64}
                      <img src={char.portrait_base64} alt={char.name} class="w-full h-full object-cover object-top" />
                    {:else}
                      <Users class="w-5 h-5 text-indigo-400" />
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
              {@const isSelected = selectedLibraryChars.has(charName)}
              <div
                class="group p-2 rounded-xl border flex items-center justify-between gap-2.5 transition cursor-grab active:cursor-grabbing shadow-sm {
                  isSelected
                    ? 'bg-brand-950/40 border-brand-500/80'
                    : 'bg-dark-800 hover:bg-dark-750 border-dark-600/60 hover:border-brand-500/40'
                }"
                draggable="true"
                on:dragstart={(e) => handleLibraryDragStart(e, 'vault', charName, asset)}
              >
                <div class="flex items-center gap-2 min-w-0">
                  <button
                    type="button"
                    class="p-0.5 rounded text-slate-500 hover:text-brand-400"
                    on:click|stopPropagation={() => toggleLibraryCharSelect(charName)}
                  >
                    {#if isSelected}
                      <CheckSquare class="w-4 h-4 text-brand-400" />
                    {:else}
                      <Square class="w-4 h-4 text-dark-500 group-hover:text-slate-400" />
                    {/if}
                  </button>

                  <div class="w-10 h-10 rounded-xl bg-dark-900 border border-dark-600/80 overflow-hidden flex items-center justify-center flex-shrink-0 relative shadow-inner">
                    {#if asset.preview_base64}
                      <img src={asset.preview_base64} alt={charName} class="w-full h-full object-cover object-top" />
                    {:else}
                      <Sparkles class="w-5 h-5 text-brand-400" />
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

      <!-- Sticky Bottom Multi-Select Bar for Library -->
      {#if selectedLibraryChars.size > 0}
        <div class="p-3 border-t border-dark-600/60 bg-dark-900 flex items-center justify-between gap-2">
          <div class="text-[11px] font-bold text-slate-200 flex items-center gap-1.5">
            <span class="w-2 h-2 rounded-full bg-indigo-400 animate-pulse"></span>
            <span>{selectedLibraryChars.size} selected</span>
          </div>
          <div class="flex items-center gap-1.5">
            <button
              type="button"
              class="px-3 py-1.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold transition shadow-md shadow-indigo-950/50"
              on:click={addSelectedLibraryCharsToRoster}
            >
              + Add to Grid
            </button>
            <button
              type="button"
              class="p-1.5 rounded-xl text-slate-400 hover:text-white hover:bg-dark-800"
              on:click={deselectAllLibraryChars}
            >
              <X class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      {/if}
    </aside>

    <!-- 2. Middle Center Work Area (Interactive Grid or Extra Stages) -->
    <main class="flex-1 flex flex-col min-w-0 bg-dark-900 overflow-y-auto p-5 relative">
      {#if loading}
        <div class="flex-1 flex items-center justify-center">
          <div class="text-xs text-slate-400 animate-pulse">Loading select.def roster...</div>
        </div>
      {:else if activeView === 'grid' && roster}
        <!-- Grid Toolbar -->
        <div class="flex flex-wrap items-center justify-between gap-3 pb-3 mb-4 border-b border-dark-700/60 flex-shrink-0">
          <div class="flex items-center gap-3">
            <span class="text-xs font-bold uppercase tracking-wider text-slate-300 flex items-center gap-1.5">
              <Grid class="w-3.5 h-3.5 text-indigo-400" />
              <span>Select Screen Matrix</span>
            </span>

            <!-- Grid Dimensions -->
            <div class="flex items-center gap-2 bg-dark-800 px-3 py-1 rounded-xl border border-dark-600/60 text-xs">
              <span class="text-slate-400 text-[11px]">Cols:</span>
              <button
                type="button"
                class="px-1.5 py-0.2 rounded hover:bg-dark-700 font-bold text-slate-300"
                on:click={() => (gridColumns = Math.max(1, gridColumns - 1))}
              >-</button>
              <span class="font-mono text-indigo-400 font-bold">{gridColumns}</span>
              <button
                type="button"
                class="px-1.5 py-0.2 rounded hover:bg-dark-700 font-bold text-slate-300"
                on:click={() => (gridColumns = gridColumns + 1)}
              >+</button>

              <span class="text-dark-600 mx-1">|</span>

              <span class="text-slate-400 text-[11px]">Rows:</span>
              <button
                type="button"
                class="px-1.5 py-0.2 rounded hover:bg-dark-700 font-bold text-slate-300"
                on:click={() => (gridRows = Math.max(1, gridRows - 1))}
              >-</button>
              <span class="font-mono text-indigo-400 font-bold">{gridRows}</span>
              <button
                type="button"
                class="px-1.5 py-0.2 rounded hover:bg-dark-700 font-bold text-slate-300"
                on:click={() => (gridRows = gridRows + 1)}
              >+</button>

              <span class="text-[10px] text-slate-500 font-mono ml-1">
                ({gridColumns * gridRows} total cells)
              </span>
            </div>

            <!-- Density Switcher -->
            <div class="flex items-center p-0.5 bg-dark-800 rounded-xl border border-dark-600/60 text-xs">
              <button
                type="button"
                class="px-2.5 py-1 rounded-lg text-[11px] font-semibold transition {density === 'arcade' ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:text-slate-200'}"
                on:click={() => (density = 'arcade')}
              >
                Arcade (Compact)
              </button>
              <button
                type="button"
                class="px-2.5 py-1 rounded-lg text-[11px] font-semibold transition {density === 'medium' ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:text-slate-200'}"
                on:click={() => (density = 'medium')}
              >
                Medium
              </button>
              <button
                type="button"
                class="px-2.5 py-1 rounded-lg text-[11px] font-semibold transition {density === 'large' ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:text-slate-200'}"
                on:click={() => (density = 'large')}
              >
                Large
              </button>
            </div>
          </div>

          <!-- Quick Fill & Roster Tools -->
          <div class="flex items-center gap-2">
            <!-- Auto Populate All Project Characters -->
            <button
              type="button"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-indigo-600/20 hover:bg-indigo-600/30 border border-indigo-500/40 text-xs font-semibold text-indigo-300 transition"
              title="Populate roster with all fighters found in project chars/"
              on:click={autoPopulateAllProjectCharacters}
            >
              <Sparkles class="w-3.5 h-3.5 text-indigo-400" />
              <span>Auto-Populate All</span>
            </button>

            <!-- Shuffle Grid -->
            <button
              type="button"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-dark-800 hover:bg-dark-700 border border-dark-600/60 text-xs font-semibold text-slate-300 transition"
              title="Shuffle character placement on the matrix"
              on:click={shuffleRosterLayout}
            >
              <Shuffle class="w-3.5 h-3.5 text-cyan-400" />
              <span>Shuffle</span>
            </button>

            <!-- Sort Dropdown -->
            <div class="flex items-center gap-1 bg-dark-800 px-2.5 py-1.5 rounded-xl border border-dark-600/60 text-xs">
              <ArrowUpDown class="w-3.5 h-3.5 text-slate-400" />
              <select
                class="bg-transparent text-slate-300 text-xs font-medium focus:outline-none cursor-pointer"
                on:change={(e) => sortRosterBy(e.currentTarget.value as any)}
              >
                <option value="" disabled selected class="bg-dark-900">Sort Fighters...</option>
                <option value="name_asc" class="bg-dark-900">Name (A-Z)</option>
                <option value="name_desc" class="bg-dark-900">Name (Z-A)</option>
                <option value="author" class="bg-dark-900">By Author</option>
                <option value="order" class="bg-dark-900">By Boss Order</option>
              </select>
            </div>

            <button
              type="button"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-dark-800 hover:bg-dark-700 border border-dark-600/60 text-xs font-semibold text-purple-300 transition"
              title="Fill all remaining empty slots on screen with Random Select"
              on:click={fillRemainingWithRandom}
            >
              <HelpCircle class="w-3.5 h-3.5 text-purple-400" />
              <span>Fill with ?</span>
            </button>

            <button
              type="button"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-dark-800 hover:bg-dark-700 border border-dark-600/60 text-xs font-semibold text-slate-300 transition"
              title="Remove trailing empty slots and shrink grid rows"
              on:click={trimTrailingEmptySlots}
            >
              <span>Trim Empty</span>
            </button>

            <!-- Multi Select Mode Toggle -->
            <button
              type="button"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl border text-xs font-semibold transition {
                isGridSelectMode
                  ? 'bg-indigo-600 text-white border-indigo-500 shadow'
                  : 'bg-dark-800 hover:bg-dark-700 border-dark-600/60 text-slate-300'
              }"
              title="Toggle multi-select mode on grid slots"
              on:click={() => {
                isGridSelectMode = !isGridSelectMode;
                if (!isGridSelectMode) selectedGridSlots = new Set();
              }}
            >
              <CheckSquare class="w-3.5 h-3.5" />
              <span>Select Mode</span>
            </button>

            <!-- Clear Roster -->
            <button
              type="button"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-rose-600/10 hover:bg-rose-600/20 border border-rose-500/30 text-xs font-semibold text-rose-300 transition"
              title="Clear all character slots from the roster"
              on:click={clearEntireRoster}
            >
              <RotateCcw class="w-3.5 h-3.5" />
              <span>Clear</span>
            </button>
          </div>
        </div>

        <!-- The Interactive Grid Matrix -->
        <div class="flex-1 overflow-auto flex justify-center items-start">
          <div
            class="grid gap-2.5 p-3 bg-dark-950/80 border border-dark-700/70 rounded-2xl shadow-2xl {
              density === 'arcade' ? 'max-w-4xl' : density === 'medium' ? 'max-w-5xl' : 'w-full'
            }"
            style="grid-template-columns: repeat({gridColumns}, minmax(0, 1fr));"
          >
            {#each Array(totalGridCells) as _, idx}
              {@const slot = getSlotAt(idx)}
              {@const isVirtual = idx >= (roster?.slots.length || 0)}
              {@const isMultiSelected = selectedGridSlots.has(idx)}
              {@const dupCount = slot.character ? characterCounts[slot.character.toLowerCase()] || 0 : 0}
              <!-- svelte-ignore a11y_click_to_play -->
              <div
                class="group relative rounded-xl border transition-all duration-150 cursor-pointer flex flex-col justify-between overflow-hidden select-none {
                  density === 'arcade' ? 'aspect-[4/3] min-h-[64px]' : density === 'medium' ? 'aspect-square min-h-[88px]' : 'aspect-square min-h-[120px]'
                } {
                  isMultiSelected
                    ? 'border-cyan-400 ring-2 ring-cyan-400/80 bg-cyan-950/40 shadow-lg shadow-cyan-950/60'
                    : selectedSlotIndex === idx
                    ? 'border-indigo-500 ring-2 ring-indigo-500/60 shadow-lg shadow-indigo-950/60'
                    : isVirtual || slot.type === 'empty'
                    ? 'border-dark-700/60 bg-dark-900/40 hover:bg-dark-800/80 hover:border-dark-500'
                    : 'border-dark-600/80 bg-dark-850 hover:border-slate-400/80 shadow-md'
                } {
                  slot.type === 'disabled' ? 'opacity-50 grayscale' : ''
                }"
                draggable="true"
                on:click={(e) => selectOrActivateSlot(idx, e)}
                on:contextmenu={(e) => handleSlotRightClick(e, idx)}
                on:mouseenter={() => (hoveredSlotIndex = idx)}
                on:mouseleave={() => (hoveredSlotIndex = null)}
                on:dragstart={(e) => handleSlotDragStart(e, idx)}
                on:dragover={handleSlotDragOver}
                on:drop={(e) => handleSlotDrop(e, idx)}
              >
                {#if slot.type === 'character' || slot.type === 'disabled'}
                  <!-- Full Edge-to-Edge Fighter Portrait -->
                  {#if slot.portrait_base64}
                    <img
                      src={slot.portrait_base64}
                      alt={slot.character}
                      class="absolute inset-0 w-full h-full object-cover object-top group-hover:scale-105 transition-transform duration-200 pointer-events-none"
                    />
                  {:else}
                    <div class="absolute inset-0 bg-gradient-to-br from-indigo-950/70 to-dark-900 flex items-center justify-center pointer-events-none">
                      <Users class="w-8 h-8 text-indigo-400/40" />
                    </div>
                  {/if}

                  <!-- Bottom Dark Gradient Overlay for Crisp Text Readability -->
                  <div class="absolute inset-0 bg-gradient-to-t from-dark-950/95 via-dark-950/30 to-transparent pointer-events-none"></div>

                  <!-- Top Index & Status Badges -->
                  <div class="relative z-10 flex items-center justify-between p-1.5 w-full pointer-events-none">
                    <span class="font-mono text-[9px] px-1 py-0.2 rounded bg-black/70 backdrop-blur-xs {selectedSlotIndex === idx ? 'text-indigo-400 font-bold border border-indigo-500/40' : 'text-slate-300'}">
                      #{idx + 1}
                    </span>

                    <div class="flex items-center gap-1">
                      <!-- Duplicate indicator badge -->
                      {#if dupCount > 1}
                        <span class="px-1 py-0.2 rounded bg-amber-500/90 text-black font-extrabold text-[8px] shadow" title="This character is placed in {dupCount} slots">
                          {dupCount}x
                        </span>
                      {/if}

                      {#if slot.type === 'disabled'}
                        <span class="px-1 py-0.2 rounded bg-rose-600/90 text-white font-bold text-[8px] shadow">
                          OFF
                        </span>
                      {:else if slot.order && slot.order > 0}
                        <span class="px-1.5 py-0.2 rounded bg-amber-500 text-black font-extrabold text-[8px] flex items-center gap-0.5 shadow">
                          <Crown class="w-2.5 h-2.5 fill-black" />
                          <span>{slot.order}</span>
                        </span>
                      {/if}
                    </div>
                  </div>

                  <!-- Bottom Fighter Name & Stage Subtitle -->
                  <div class="relative z-10 p-1.5 text-center pointer-events-none">
                    <div class="text-[10px] font-bold text-white leading-tight drop-shadow truncate">
                      {slot.display_name || slot.character}
                    </div>
                    {#if slot.home_stage && density !== 'arcade'}
                      <div class="text-[8px] text-emerald-400 truncate flex items-center justify-center gap-0.5">
                        <Mountain class="w-2 h-2" />
                        <span>{slot.home_stage.replace(/^stages\//i, '').replace(/\.def$/i, '')}</span>
                      </div>
                    {/if}
                  </div>

                {:else if slot.type === 'randomselect'}
                  <!-- Random Select Tile -->
                  <div class="absolute inset-0 bg-gradient-to-br from-purple-950/80 via-indigo-950/60 to-dark-900 flex flex-col items-center justify-center pointer-events-none">
                    <div class="text-xl font-black text-purple-400 drop-shadow animate-pulse">?</div>
                  </div>
                  <div class="relative z-10 flex items-center justify-between p-1.5 w-full pointer-events-none">
                    <span class="font-mono text-[9px] px-1 py-0.2 rounded bg-black/60 text-purple-300">#{idx + 1}</span>
                    <span class="px-1 py-0.2 rounded bg-purple-500/30 text-purple-300 font-bold text-[8px]">RANDOM</span>
                  </div>
                  <div class="relative z-10 p-1.5 text-center text-[10px] font-bold text-purple-300 pointer-events-none">
                    Random Select
                  </div>

                {:else}
                  <!-- Empty Slot Tile -->
                  <div class="relative z-10 flex items-center justify-between p-1.5 w-full pointer-events-none">
                    <span class="font-mono text-[9px] text-slate-600">#{idx + 1}</span>
                  </div>
                  <div class="flex-1 flex flex-col items-center justify-center pointer-events-none">
                    <Plus class="w-4 h-4 text-slate-600 group-hover:text-slate-400 transition" />
                    <span class="text-[8px] text-slate-600 group-hover:text-slate-400 mt-0.5 font-medium">Empty Slot</span>
                  </div>
                  <div class="p-1"></div>
                {/if}
              </div>
            {/each}
          </div>
        </div>

        <!-- Floating Bulk Action Toolbar for Selected Grid Slots -->
        {#if selectedGridSlots.size > 0}
          <div class="fixed bottom-6 left-1/2 -translate-x-1/2 z-30 bg-dark-850/95 border border-indigo-500/60 backdrop-blur-xl shadow-2xl rounded-2xl px-5 py-3 flex items-center gap-4 animate-in slide-in-from-bottom-4">
            <div class="flex items-center gap-2 pr-3 border-r border-dark-700">
              <span class="w-2.5 h-2.5 rounded-full bg-indigo-400 animate-pulse"></span>
              <span class="text-xs font-bold text-slate-100">{selectedGridSlots.size} slots selected</span>
            </div>

            <button
              type="button"
              class="px-3 py-1.5 rounded-xl bg-dark-800 hover:bg-dark-750 border border-dark-600 text-xs font-semibold text-slate-200 transition"
              on:click={bulkClearSelectedGridSlots}
            >
              Clear to Empty
            </button>

            <button
              type="button"
              class="px-3 py-1.5 rounded-xl bg-rose-600/20 hover:bg-rose-600/30 border border-rose-500/40 text-xs font-bold text-rose-300 transition"
              on:click={bulkDisableSelectedGridSlots}
            >
              Disable Selected
            </button>

            <button
              type="button"
              class="px-2.5 py-1.5 rounded-xl text-slate-400 hover:text-slate-200 text-xs font-semibold hover:bg-dark-700 transition"
              on:click={() => (selectedGridSlots = new Set())}
            >
              Deselect All
            </button>
          </div>
        {/if}
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

    <!-- Right-click Context Menu -->
    {#if contextMenu}
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div class="fixed inset-0 z-40" on:click={closeContextMenu}></div>
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div
        class="fixed z-50 bg-dark-850 border border-dark-600 rounded-xl shadow-2xl py-1.5 min-w-[170px]"
        style="left: {contextMenu.x}px; top: {Math.min(contextMenu.y, window.innerHeight - 130)}px;"
        on:click|stopPropagation
      >
        <div class="px-3 py-1.5 border-b border-dark-700 mb-1">
          <span class="text-[10px] font-bold uppercase tracking-wider text-slate-500">Slot #{contextMenu.slotIndex + 1}</span>
        </div>
        <button
          type="button"
          class="w-full text-left px-3 py-2 text-xs font-semibold text-slate-200 hover:bg-dark-700 flex items-center gap-2 transition"
          on:click={() => { clearSlot(contextMenu!.slotIndex); closeContextMenu(); }}
        >
          <X class="w-3.5 h-3.5 text-slate-400" />
          Clear to Empty
        </button>
        {#if roster && roster.slots[contextMenu.slotIndex]?.type === 'character'}
          <button
            type="button"
            class="w-full text-left px-3 py-2 text-xs font-semibold text-purple-300 hover:bg-dark-700 flex items-center gap-2 transition"
            on:click={() => { if (contextMenu !== null && roster) { roster.slots[contextMenu.slotIndex] = { index: contextMenu.slotIndex, type: 'randomselect', character: 'randomselect', display_name: 'Random Select', include_in_arcade: true }; roster.slots = [...roster.slots]; } closeContextMenu(); }}
          >
            <HelpCircle class="w-3.5 h-3.5 text-purple-400" />
            Set Random Select
          </button>
          <button
            type="button"
            class="w-full text-left px-3 py-2 text-xs font-semibold text-amber-300 hover:bg-dark-700 flex items-center gap-2 transition"
            on:click={() => { if (contextMenu !== null) { toggleSlotDisabled(contextMenu.slotIndex); } closeContextMenu(); }}
          >
            <EyeOff class="w-3.5 h-3.5 text-amber-400" />
            Toggle Disabled
          </button>
        {/if}
        <div class="border-t border-dark-700 mt-1 pt-1">
          <button
            type="button"
            class="w-full text-left px-3 py-2 text-xs font-semibold text-rose-400 hover:bg-rose-500/10 flex items-center gap-2 transition"
            on:click={() => { deleteSlot(contextMenu!.slotIndex); closeContextMenu(); }}
          >
            <Trash2 class="w-3.5 h-3.5" />
            Delete Slot
          </button>
        </div>
      </div>
    {/if}

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
              value={selectedSlot.type}
              on:change={(e) => {
                if (selectedSlotIndex !== null && roster && roster.slots[selectedSlotIndex]) {
                  const newType = e.currentTarget.value as RosterCharacterSlot['type'];
                  const s = roster.slots[selectedSlotIndex];
                  if (newType === 'empty') {
                    roster.slots[selectedSlotIndex] = { index: selectedSlotIndex, type: 'empty', character: '', include_in_arcade: true };
                  } else if (newType === 'randomselect') {
                    roster.slots[selectedSlotIndex] = { index: selectedSlotIndex, type: 'randomselect', character: 'randomselect', display_name: 'Random Select', include_in_arcade: true };
                  } else {
                    roster.slots[selectedSlotIndex] = { ...s, type: newType };
                  }
                  roster.slots = [...roster.slots];
                }
              }}
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
                  on:change={() => { roster && (roster.slots = [...roster.slots]); }}
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
                  value={selectedSlot.music || ''}
                  placeholder="sound/my_theme.mp3"
                  class="w-full bg-dark-800 border border-dark-600 rounded-xl px-3 py-2 text-xs text-slate-200 placeholder:text-slate-600 focus:outline-none focus:border-cyan-500"
                  on:input={(e) => {
                    if (selectedSlotIndex !== null && roster?.slots[selectedSlotIndex]) {
                      roster.slots[selectedSlotIndex].music = e.currentTarget.value;
                      roster.slots = [...roster.slots];
                    }
                  }}
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
                    value={selectedSlot.order ?? 0}
                    class="w-20 bg-dark-800 border border-dark-600 rounded-xl px-3 py-2 text-xs text-slate-200 font-mono focus:outline-none focus:border-amber-500"
                    on:input={(e) => {
                      if (selectedSlotIndex !== null && roster?.slots[selectedSlotIndex]) {
                        roster.slots[selectedSlotIndex].order = parseInt(e.currentTarget.value) || 0;
                        roster.slots = [...roster.slots];
                      }
                    }}
                  />
                  <span class="text-[11px] text-slate-500">
                    {(selectedSlot.order ?? 0) === 0 ? 'Normal Fighter' : `Boss Tier ${selectedSlot.order}`}
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
