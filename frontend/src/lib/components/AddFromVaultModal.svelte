<script lang="ts">
  import { onMount } from 'svelte';
  import type { VaultAsset, AssetCategory } from '../types';
  import {
    X,
    Search,
    Package,
    Check,
    Plus,
    User,
    Layers,
    Tag,
    HardDrive,
    Sparkles,
    ShieldAlert
  } from 'lucide-svelte';
  import { vaultStore, filteredAssets } from '../stores/vaultStore';
  import { projectStore } from '../stores/projectStore';

  export let isOpen = false;
  export let projectDir = '';
  export let targetCategory: AssetCategory = 'fighters';
  export let onClose: () => void;

  let selectedKeys: string[] = [];
  let isAdding = false;

  onMount(async () => {
    await vaultStore.loadVaults();
    await vaultStore.loadAssets('all');
    vaultStore.setCategory(targetCategory);
  });

  $: {
    if (isOpen) {
      vaultStore.setCategory(targetCategory);
      selectedKeys = [];
    }
  }

  function toggleSelect(asset: VaultAsset) {
    if (selectedKeys.includes(asset.key)) {
      selectedKeys = selectedKeys.filter((k) => k !== asset.key);
    } else {
      selectedKeys = [...selectedKeys, asset.key];
    }
  }

  async function handleAddSelected() {
    if (selectedKeys.length === 0 || !projectDir) return;
    isAdding = true;

    for (const key of selectedKeys) {
      // Find asset to determine its owning vault or default
      const asset = $filteredAssets.find((a) => a.key === key);
      const vaultId = $vaultStore.activeVaultId !== 'all' ? $vaultStore.activeVaultId : 'vault-default';
      await vaultStore.linkToProject(projectDir, vaultId, key);
    }

    isAdding = false;
    onClose();
  }
</script>

{#if isOpen}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-in fade-in duration-150">
    <div class="relative w-full max-w-3xl h-[600px] flex flex-col rounded-2xl bg-dark-850 border border-dark-600/80 shadow-2xl p-6 space-y-4 animate-in zoom-in-95 duration-150">
      <!-- Header -->
      <div class="flex items-center justify-between border-b border-dark-700/70 pb-3">
        <div class="flex items-center gap-2.5">
          <div class="p-2 rounded-xl bg-brand-500/10 text-brand-400 border border-brand-500/20">
            <Package class="w-5 h-5" />
          </div>
          <div>
            <h3 class="font-bold text-slate-100 text-base">Add {targetCategory === 'fighters' ? 'Fighters' : 'Stages'} from Vault</h3>
            <p class="text-xs text-slate-400">Select assets from your vaults to link into this project's roster</p>
          </div>
        </div>
        <button
          class="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-dark-700 transition"
          on:click={onClose}
        >
          <X class="w-4 h-4" />
        </button>
      </div>

      <!-- Controls: Search & Vault Select -->
      <div class="flex items-center gap-3">
        <div class="relative flex-1">
          <Search class="absolute left-3 top-2.5 w-4 h-4 text-slate-400" />
          <input
            type="text"
            placeholder="Search characters, authors, tags..."
            bind:value={$vaultStore.searchQuery}
            class="w-full bg-dark-900 border border-dark-600/80 rounded-xl pl-9 pr-4 py-2 text-xs text-slate-200 focus:outline-none focus:border-brand-500 transition"
          />
        </div>
        <select
          value={$vaultStore.activeVaultId}
          on:change={(e) => vaultStore.selectVault(e.currentTarget.value)}
          class="bg-dark-900 border border-dark-600/80 rounded-xl px-3 py-2 text-xs text-slate-200 focus:outline-none focus:border-brand-500 transition"
        >
          <option value="all">All Vaults</option>
          {#each $vaultStore.vaults as v}
            <option value={v.id}>{v.name} ({v.asset_count})</option>
          {/each}
        </select>
      </div>

      <!-- Asset Grid -->
      <div class="flex-1 overflow-y-auto pr-1">
        {#if $filteredAssets.length === 0}
          <div class="h-full flex flex-col items-center justify-center p-8 text-center text-slate-500">
            <Package class="w-12 h-12 stroke-1 opacity-40 mb-2" />
            <p class="text-sm font-semibold text-slate-400">No matching assets in Vault</p>
            <p class="text-xs text-slate-500 max-w-sm mt-1">
              Add characters or stages into your Vault library first, or adjust your search filters.
            </p>
          </div>
        {:else}
          <div class="grid grid-cols-3 gap-3">
            {#each $filteredAssets as asset}
              {@const isSelected = selectedKeys.includes(asset.key)}
              <div
                role="button"
                tabindex="0"
                class="relative flex items-center gap-3 p-2.5 rounded-xl border transition-all cursor-pointer text-left {isSelected ? 'bg-brand-500/10 border-brand-500 shadow-md shadow-brand-500/10' : 'bg-dark-900/90 border-dark-600/70 hover:border-dark-500'}"
                on:click={() => toggleSelect(asset)}
                on:keydown={(e) => { if (e.key === 'Enter' || e.key === ' ') toggleSelect(asset); }}
              >
                <!-- Thumbnail -->
                <div class="h-12 w-12 rounded-lg bg-dark-950 border border-dark-700 flex items-center justify-center overflow-hidden flex-shrink-0">
                  {#if asset.preview_base64}
                    <img src={asset.preview_base64} alt={asset.display_name} class="h-full w-full object-contain p-0.5" />
                  {:else}
                    <User class="w-5 h-5 text-slate-600" />
                  {/if}
                </div>

                <!-- Info -->
                <div class="flex-1 min-w-0">
                  <h5 class="text-xs font-bold text-slate-200 truncate">{asset.display_name}</h5>
                  <p class="text-[11px] text-slate-400 truncate mt-0.5">{asset.author || 'Unknown'}</p>
                </div>

                <!-- Checkmark -->
                <div class="w-5 h-5 rounded-md flex items-center justify-center border transition {isSelected ? 'bg-brand-500 border-brand-400 text-white' : 'border-dark-600 bg-dark-800 text-transparent'}">
                  <Check class="w-3.5 h-3.5" />
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-between pt-3 border-t border-dark-700/70">
        <span class="text-xs text-slate-400">
          {selectedKeys.length} asset{selectedKeys.length === 1 ? '' : 's'} selected
        </span>

        <div class="flex items-center gap-2.5">
          <button
            type="button"
            class="px-4 py-2 text-xs font-semibold text-slate-400 hover:text-slate-200 transition"
            on:click={onClose}
          >
            Cancel
          </button>
          <button
            type="button"
            disabled={selectedKeys.length === 0 || isAdding}
            on:click={handleAddSelected}
            class="px-5 py-2 bg-brand-600 hover:bg-brand-500 disabled:opacity-50 text-white text-xs font-bold rounded-xl transition shadow-md shadow-brand-600/20 flex items-center gap-1.5"
          >
            <Plus class="w-4 h-4" />
            <span>{isAdding ? 'Adding...' : 'Add to Roster'}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
