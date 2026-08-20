<script lang="ts">
  import { X, Package, FolderPlus, Layers, Sparkles } from 'lucide-svelte';

  export let isOpen = false;
  export let filePath = '';
  export let activeVaultName = 'Default Vault';
  export let onConfirm: (mode: 'new_vault' | 'current_vault') => void;
  export let onCancel: () => void;

  $: fileName = filePath.split('/').pop()?.split('\\').pop() || 'Archive';
</script>

{#if isOpen}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-in fade-in duration-150">
    <div class="relative w-full max-w-lg rounded-2xl bg-dark-850 border border-dark-600/80 shadow-2xl p-6 space-y-5 animate-in zoom-in-95 duration-150">
      <!-- Header -->
      <div class="flex items-center justify-between border-b border-dark-700/70 pb-4">
        <div class="flex items-center gap-2.5">
          <div class="p-2 rounded-xl bg-purple-500/10 text-purple-400 border border-purple-500/20">
            <Package class="w-5 h-5" />
          </div>
          <div>
            <h3 class="font-bold text-slate-100 text-base">Package Ingestion Strategy</h3>
            <p class="text-xs text-slate-400">Choose how to organize the assets inside <span class="text-slate-200 font-mono">{fileName}</span></p>
          </div>
        </div>
        <button
          class="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-dark-700 transition"
          on:click={onCancel}
        >
          <X class="w-4 h-4" />
        </button>
      </div>

      <!-- Option Cards -->
      <div class="space-y-3">
        <!-- Option 1: Standalone Vault -->
        <button
          type="button"
          class="w-full text-left p-4 rounded-xl bg-dark-900 border border-dark-600/70 hover:border-brand-500/70 hover:bg-dark-800/80 transition group cursor-pointer"
          on:click={() => onConfirm('new_vault')}
        >
          <div class="flex items-start gap-3">
            <div class="p-2 rounded-lg bg-brand-500/10 text-brand-400 border border-brand-500/20 mt-0.5 group-hover:scale-105 transition-transform">
              <FolderPlus class="w-4 h-4" />
            </div>
            <div class="flex-1">
              <div class="flex items-center justify-between">
                <h4 class="font-bold text-slate-200 text-sm group-hover:text-brand-300 transition-colors">
                  Create Dedicated Standalone Vault
                </h4>
                <span class="text-[10px] font-bold text-brand-400 bg-brand-500/10 px-2 py-0.5 rounded-full border border-brand-500/20">
                  Recommended for Games
                </span>
              </div>
              <p class="text-xs text-slate-400 mt-1">
                Creates a new vault named after the archive. Best for full game conversions or large curated collections.
              </p>
            </div>
          </div>
        </button>

        <!-- Option 2: Extract to Current Vault -->
        <button
          type="button"
          class="w-full text-left p-4 rounded-xl bg-dark-900 border border-dark-600/70 hover:border-brand-500/70 hover:bg-dark-800/80 transition group cursor-pointer"
          on:click={() => onConfirm('current_vault')}
        >
          <div class="flex items-start gap-3">
            <div class="p-2 rounded-lg bg-slate-700/40 text-slate-300 border border-slate-600/30 mt-0.5 group-hover:scale-105 transition-transform">
              <Layers class="w-4 h-4" />
            </div>
            <div class="flex-1">
              <h4 class="font-bold text-slate-200 text-sm group-hover:text-brand-300 transition-colors">
                Extract into "{activeVaultName}"
              </h4>
              <p class="text-xs text-slate-400 mt-1">
                Dissects all characters and stages into individual items inside your current vault, tagged with the origin package name.
              </p>
            </div>
          </div>
        </button>
      </div>

      <!-- Footer -->
      <div class="flex justify-end pt-2">
        <button
          type="button"
          class="px-4 py-2 text-xs font-semibold text-slate-400 hover:text-slate-200 transition"
          on:click={onCancel}
        >
          Cancel
        </button>
      </div>
    </div>
  </div>
{/if}
