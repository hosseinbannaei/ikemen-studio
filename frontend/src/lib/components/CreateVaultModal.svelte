<script lang="ts">
  import { X, FolderPlus, FolderOpen, HardDrive, Sparkles } from 'lucide-svelte';
  import { vaultStore } from '../stores/vaultStore';

  export let isOpen = false;
  export let onClose: () => void;

  let mode: 'create' | 'register' = 'create';
  let vaultName = '';
  let vaultDescription = '';
  let customPath = '';
  let isSubmitting = false;

  async function handleBrowseFolder() {
    const selected = await vaultStore.browseFolder();
    if (selected) {
      customPath = selected;
    }
  }

  async function handleSubmit() {
    if (mode === 'create') {
      if (!vaultName.trim()) return;
      isSubmitting = true;
      const ok = await vaultStore.create(vaultName.trim(), vaultDescription.trim(), customPath.trim());
      isSubmitting = false;
      if (ok) {
        resetAndClose();
      }
    } else {
      if (!customPath.trim()) return;
      isSubmitting = true;
      const ok = await vaultStore.register(customPath.trim());
      isSubmitting = false;
      if (ok) {
        resetAndClose();
      }
    }
  }

  function resetAndClose() {
    vaultName = '';
    vaultDescription = '';
    customPath = '';
    onClose();
  }
</script>

{#if isOpen}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-in fade-in duration-150">
    <div class="relative w-full max-w-md rounded-2xl bg-dark-850 border border-dark-600/80 shadow-2xl p-6 space-y-5 animate-in zoom-in-95 duration-150">
      <!-- Header -->
      <div class="flex items-center justify-between border-b border-dark-700/70 pb-4">
        <div class="flex items-center gap-2.5">
          <div class="p-2 rounded-xl bg-brand-500/10 text-brand-400 border border-brand-500/20">
            <FolderPlus class="w-5 h-5" />
          </div>
          <div>
            <h3 class="font-bold text-slate-100 text-base">Manage Asset Vaults</h3>
            <p class="text-xs text-slate-400">Create a new vault or mount an existing directory</p>
          </div>
        </div>
        <button
          class="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-dark-700 transition"
          on:click={resetAndClose}
        >
          <X class="w-4 h-4" />
        </button>
      </div>

      <!-- Mode Selector -->
      <div class="grid grid-cols-2 gap-2 p-1 bg-dark-950 rounded-xl border border-dark-700/60">
        <button
          class="py-2 text-xs font-semibold rounded-lg transition {mode === 'create' ? 'bg-dark-800 text-brand-300 shadow-sm' : 'text-slate-400 hover:text-slate-200'}"
          on:click={() => (mode = 'create')}
        >
          Create New Vault
        </button>
        <button
          class="py-2 text-xs font-semibold rounded-lg transition {mode === 'register' ? 'bg-dark-800 text-brand-300 shadow-sm' : 'text-slate-400 hover:text-slate-200'}"
          on:click={() => (mode = 'register')}
        >
          Mount Existing Vault
        </button>
      </div>

      <!-- Body Form -->
      <div class="space-y-4">
        {#if mode === 'create'}
          <div>
            <label class="block text-xs font-semibold text-slate-300 mb-1.5">Vault Name *</label>
            <input
              type="text"
              bind:value={vaultName}
              placeholder="e.g. Capcom Fighters or Anime Stages"
              class="w-full bg-dark-900 border border-dark-600/80 rounded-xl px-3.5 py-2.5 text-sm text-slate-200 focus:outline-none focus:border-brand-500 transition"
            />
          </div>

          <div>
            <label class="block text-xs font-semibold text-slate-300 mb-1.5">Description (Optional)</label>
            <input
              type="text"
              bind:value={vaultDescription}
              placeholder="e.g. Curated high-res characters"
              class="w-full bg-dark-900 border border-dark-600/80 rounded-xl px-3.5 py-2.5 text-sm text-slate-200 focus:outline-none focus:border-brand-500 transition"
            />
          </div>

          <div>
            <label class="block text-xs font-semibold text-slate-300 mb-1.5">Storage Location (Leave blank for default)</label>
            <div class="flex gap-2">
              <input
                type="text"
                bind:value={customPath}
                placeholder="Default app data directory"
                class="flex-1 bg-dark-900 border border-dark-600/80 rounded-xl px-3.5 py-2 text-xs font-mono text-slate-300 focus:outline-none focus:border-brand-500 transition"
              />
              <button
                type="button"
                on:click={handleBrowseFolder}
                class="px-3 py-2 bg-dark-750 hover:bg-dark-700 border border-dark-600 text-slate-300 rounded-xl text-xs font-semibold flex items-center gap-1.5 transition"
              >
                <FolderOpen class="w-3.5 h-3.5" />
                Browse
              </button>
            </div>
          </div>
        {:else}
          <div>
            <label class="block text-xs font-semibold text-slate-300 mb-1.5">Existing Vault Folder Path *</label>
            <div class="flex gap-2">
              <input
                type="text"
                bind:value={customPath}
                placeholder="/path/to/existing/vault"
                class="flex-1 bg-dark-900 border border-dark-600/80 rounded-xl px-3.5 py-2 text-xs font-mono text-slate-300 focus:outline-none focus:border-brand-500 transition"
              />
              <button
                type="button"
                on:click={handleBrowseFolder}
                class="px-3 py-2 bg-dark-750 hover:bg-dark-700 border border-dark-600 text-slate-300 rounded-xl text-xs font-semibold flex items-center gap-1.5 transition"
              >
                <FolderOpen class="w-3.5 h-3.5" />
                Browse
              </button>
            </div>
            <p class="text-[11px] text-slate-400 mt-2">
              Select a directory containing a valid <code class="text-brand-400 font-mono">vault.json</code> manifest.
            </p>
          </div>
        {/if}
      </div>

      <!-- Footer Buttons -->
      <div class="flex items-center justify-end gap-2.5 pt-2 border-t border-dark-700/60">
        <button
          type="button"
          class="px-4 py-2 text-xs font-semibold text-slate-400 hover:text-slate-200 transition"
          on:click={resetAndClose}
        >
          Cancel
        </button>
        <button
          type="button"
          disabled={isSubmitting || (mode === 'create' ? !vaultName.trim() : !customPath.trim())}
          on:click={handleSubmit}
          class="px-5 py-2.5 bg-brand-600 hover:bg-brand-500 disabled:opacity-50 text-white text-xs font-bold rounded-xl transition shadow-md shadow-brand-600/20"
        >
          {isSubmitting ? 'Processing...' : mode === 'create' ? 'Create Vault' : 'Mount Vault'}
        </button>
      </div>
    </div>
  </div>
{/if}
