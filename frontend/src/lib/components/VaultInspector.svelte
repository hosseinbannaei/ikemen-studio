<script lang="ts">
  import type { VaultAsset, AssetMetadataUpdate } from '../types';
  import {
    X,
    User,
    Tag,
    Globe,
    FileText,
    Trash2,
    Save,
    ExternalLink,
    FolderOpen,
    Shield,
    Calendar,
    HardDrive,
    Sparkles
  } from 'lucide-svelte';
  import { vaultStore } from '../stores/vaultStore';
  import { OpenFolderInExplorer } from '../../../wailsjs/go/main/App';

  export let asset: VaultAsset;
  export let activeVaultId: string;
  export let onClose: () => void;

  let displayName = asset.display_name;
  let author = asset.author;
  let sourceUrl = asset.source_url;
  let license = asset.license || 'Unknown / Fan-made';
  let notes = asset.notes || '';
  let tagInput = '';
  let tags: string[] = asset.tags ? [...asset.tags] : [];
  let isSaving = false;
  let showConfirmDelete = false;

  $: {
    displayName = asset.display_name;
    author = asset.author;
    sourceUrl = asset.source_url;
    license = asset.license || 'Unknown / Fan-made';
    notes = asset.notes || '';
    tags = asset.tags ? [...asset.tags] : [];
    showConfirmDelete = false;
  }

  function addTag() {
    const trimmed = tagInput.trim().replace(/^#/, '');
    if (trimmed && !tags.includes(trimmed)) {
      tags = [...tags, trimmed];
      tagInput = '';
    }
  }

  function removeTag(tag: string) {
    tags = tags.filter((t) => t !== tag);
  }

  async function handleSave() {
    isSaving = true;
    const updates: AssetMetadataUpdate = {
      display_name: displayName,
      author,
      source_url: sourceUrl,
      license,
      tags,
      notes,
    };
    await vaultStore.updateAsset(activeVaultId, asset.key, updates);
    isSaving = false;
  }

  async function handleDelete() {
    await vaultStore.deleteAsset(activeVaultId, asset.key);
    onClose();
  }

  function formatBytes(b: number): string {
    if (!b) return '0 B';
    if (b < 1024 * 1024) return (b / 1024).toFixed(1) + ' KB';
    return (b / (1024 * 1024)).toFixed(1) + ' MB';
  }
</script>

<aside class="w-96 flex flex-col h-full bg-dark-900 border-l border-dark-700/80 shadow-2xl z-20 animate-in slide-in-from-right duration-200">
  <!-- Header -->
  <div class="p-4 border-b border-dark-700/80 flex items-center justify-between bg-dark-850">
    <div class="flex items-center gap-2">
      <Sparkles class="w-4 h-4 text-brand-400" />
      <h3 class="font-bold text-slate-200 text-sm">Asset Inspector</h3>
    </div>
    <button
      class="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-dark-700 transition"
      on:click={onClose}
    >
      <X class="w-4 h-4" />
    </button>
  </div>

  <!-- Content Form -->
  <div class="flex-1 overflow-y-auto p-5 space-y-5">
    <!-- Preview Box -->
    <div class="flex flex-col items-center p-4 rounded-2xl bg-dark-950/80 border border-dark-700/60">
      <div class="h-32 w-32 rounded-xl bg-dark-900 border border-dark-700 flex items-center justify-center overflow-hidden mb-3">
        {#if asset.preview_base64}
          <img src={asset.preview_base64} alt={asset.display_name} class="h-full w-full object-contain p-1" />
        {:else}
          <User class="w-12 h-12 text-slate-600" />
        {/if}
      </div>
      <span class="text-xs font-mono text-slate-400">{asset.key}</span>
      <span class="text-[11px] text-slate-500 mt-0.5">{formatBytes(asset.size_bytes)}</span>
    </div>

    <!-- Inputs -->
    <div class="space-y-4">
      <div>
        <label class="block text-[11px] font-bold uppercase tracking-wider text-slate-400 mb-1">Display Name</label>
        <input
          type="text"
          bind:value={displayName}
          class="w-full bg-dark-950 border border-dark-600/80 rounded-xl px-3 py-2 text-sm text-slate-200 focus:outline-none focus:border-brand-500 transition"
        />
      </div>

      <div>
        <label class="block text-[11px] font-bold uppercase tracking-wider text-slate-400 mb-1">Author</label>
        <input
          type="text"
          bind:value={author}
          class="w-full bg-dark-950 border border-dark-600/80 rounded-xl px-3 py-2 text-sm text-slate-200 focus:outline-none focus:border-brand-500 transition"
        />
      </div>

      <div>
        <label class="block text-[11px] font-bold uppercase tracking-wider text-slate-400 mb-1">Source / Download URL</label>
        <div class="relative">
          <input
            type="text"
            bind:value={sourceUrl}
            placeholder="https://..."
            class="w-full bg-dark-950 border border-dark-600/80 rounded-xl pl-3 pr-8 py-2 text-sm text-slate-200 focus:outline-none focus:border-brand-500 transition font-mono text-xs"
          />
          {#if sourceUrl}
            <a
              href={sourceUrl}
              target="_blank"
              rel="noreferrer"
              class="absolute right-2.5 top-2.5 text-slate-400 hover:text-brand-400 transition"
            >
              <ExternalLink class="w-3.5 h-3.5" />
            </a>
          {/if}
        </div>
      </div>

      <div>
        <label class="block text-[11px] font-bold uppercase tracking-wider text-slate-400 mb-1">License & Rights</label>
        <select
          bind:value={license}
          class="w-full bg-dark-950 border border-dark-600/80 rounded-xl px-3 py-2 text-sm text-slate-200 focus:outline-none focus:border-brand-500 transition"
        >
          <option value="Unknown / Fan-made">Unknown / Fan-made</option>
          <option value="Public Domain / CC0">Public Domain / CC0</option>
          <option value="Creative Commons (CC-BY)">Creative Commons (CC-BY)</option>
          <option value="Original IP / Commercial Safe">Original IP / Commercial Safe</option>
          <option value="Custom Author Permission">Custom Author Permission</option>
        </select>
      </div>

      <!-- Tags Input -->
      <div>
        <label class="block text-[11px] font-bold uppercase tracking-wider text-slate-400 mb-1">Tags</label>
        <div class="flex flex-wrap gap-1.5 mb-2">
          {#each tags as tag}
            <span class="inline-flex items-center gap-1 text-xs text-brand-300 bg-brand-950/80 border border-brand-800/60 px-2 py-0.5 rounded-lg">
              #{tag}
              <button on:click={() => removeTag(tag)} class="hover:text-rose-400">
                <X class="w-3 h-3" />
              </button>
            </span>
          {/each}
        </div>
        <div class="flex gap-2">
          <input
            type="text"
            bind:value={tagInput}
            placeholder="Add a tag..."
            on:keydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addTag(); } }}
            class="flex-1 bg-dark-950 border border-dark-600/80 rounded-xl px-3 py-1.5 text-xs text-slate-200 focus:outline-none focus:border-brand-500 transition"
          />
          <button
            on:click={addTag}
            class="px-3 py-1.5 bg-dark-700 hover:bg-dark-600 text-slate-200 text-xs font-semibold rounded-xl transition"
          >
            Add
          </button>
        </div>
      </div>

      <!-- Notes / Readme -->
      <div>
        <label class="block text-[11px] font-bold uppercase tracking-wider text-slate-400 mb-1">Notes & Readme</label>
        <textarea
          bind:value={notes}
          rows="4"
          placeholder="Release notes, AI patches, or custom credits..."
          class="w-full bg-dark-950 border border-dark-600/80 rounded-xl p-3 text-xs text-slate-300 focus:outline-none focus:border-brand-500 transition resize-none font-mono"
        ></textarea>
      </div>
    </div>
  </div>

  <!-- Bottom Actions -->
  <div class="p-4 border-t border-dark-700/80 bg-dark-850 flex items-center justify-between gap-3">
    {#if !showConfirmDelete}
      <button
        class="p-2.5 rounded-xl text-rose-400 hover:bg-rose-500/10 transition"
        title="Delete Asset from Vault"
        on:click={() => (showConfirmDelete = true)}
      >
        <Trash2 class="w-4 h-4" />
      </button>
    {:else}
      <div class="flex items-center gap-2">
        <button
          class="px-2.5 py-1 text-xs bg-rose-600 hover:bg-rose-500 text-white font-bold rounded-lg transition"
          on:click={handleDelete}
        >
          Confirm Delete
        </button>
        <button
          class="px-2 py-1 text-xs bg-dark-700 text-slate-300 rounded-lg hover:bg-dark-600 transition"
          on:click={() => (showConfirmDelete = false)}
        >
          Cancel
        </button>
      </div>
    {/if}

    <button
      class="flex-1 flex items-center justify-center gap-2 py-2.5 px-4 bg-brand-600 hover:bg-brand-500 text-white font-semibold rounded-xl text-sm transition shadow-lg shadow-brand-600/20 disabled:opacity-50"
      disabled={isSaving}
      on:click={handleSave}
    >
      <Save class="w-4 h-4" />
      <span>{isSaving ? 'Saving...' : 'Save Changes'}</span>
    </button>
  </div>
</aside>
