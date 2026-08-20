<script lang="ts">
  import type { VaultAsset } from '../types';
  import {
    User,
    Layers,
    Tag,
    Globe,
    FileText,
    Trash2,
    Sliders,
    Sparkles,
    ShieldAlert,
    ExternalLink
  } from 'lucide-svelte';
  import { vaultStore } from '../stores/vaultStore';

  export let asset: VaultAsset;
  export let onSelect: (asset: VaultAsset) => void = () => {};
  export let onQuickAdd: ((asset: VaultAsset) => void) | null = null;

  function formatSize(bytes: number): string {
    if (!bytes) return '0 B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  }

  function getCategoryColor(cat: string) {
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

  function getCategoryLabel(cat: string) {
    switch (cat) {
      case 'fighters':
        return 'Fighter';
      case 'stages':
        return 'Stage';
      case 'motifs':
        return 'Screenpack';
      case 'sounds':
        return 'Audio';
      default:
        return cat;
    }
  }
</script>

<div
  role="button"
  tabindex="0"
  class="group relative flex flex-col rounded-2xl bg-dark-800/90 border border-dark-600/70 hover:border-brand-500/60 hover:shadow-xl hover:shadow-brand-500/5 transition-all duration-200 overflow-hidden cursor-pointer text-left"
  on:click={() => onSelect(asset)}
  on:keydown={(e) => { if (e.key === 'Enter' || e.key === ' ') onSelect(asset); }}
>
  <!-- Card Header & Image Preview -->
  <div class="relative h-36 w-full bg-dark-900/90 flex items-center justify-center overflow-hidden border-b border-dark-700/60">
    {#if asset.preview_base64}
      <img
        src={asset.preview_base64}
        alt={asset.display_name}
        class="h-full w-full object-contain p-2 group-hover:scale-105 transition-transform duration-300"
      />
    {:else}
      <div class="flex flex-col items-center justify-center gap-1.5 text-slate-500">
        <User class="w-10 h-10 stroke-1 opacity-40" />
        <span class="text-[10px] font-medium tracking-wider uppercase opacity-60">Preview</span>
      </div>
    {/if}

    <!-- Badges overlay -->
    <div class="absolute top-2.5 left-2.5 flex items-center gap-1.5">
      <span class="text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full border backdrop-blur-md shadow-sm {getCategoryColor(asset.category)}">
        {getCategoryLabel(asset.category)}
      </span>
      {#if asset.mugen_version}
        <span class="text-[10px] font-medium text-slate-300 bg-dark-900/80 border border-dark-600/60 px-1.5 py-0.5 rounded-md backdrop-blur-md">
          {asset.mugen_version}
        </span>
      {/if}
    </div>

    <!-- Size badge -->
    <div class="absolute bottom-2 right-2.5">
      <span class="text-[10px] text-slate-400 bg-dark-900/80 px-1.5 py-0.5 rounded border border-dark-700/50 backdrop-blur-md">
        {formatSize(asset.size_bytes)}
      </span>
    </div>
  </div>

  <!-- Card Body -->
  <div class="p-4 flex-1 flex flex-col justify-between space-y-3">
    <div>
      <div class="flex items-start justify-between gap-2">
        <h4 class="font-bold text-slate-100 text-sm truncate group-hover:text-brand-300 transition-colors" title={asset.display_name}>
          {asset.display_name || asset.key}
        </h4>
      </div>
      <div class="text-[10px] text-slate-500 font-mono truncate mt-0.5" title={asset.key}>
        {asset.key}
      </div>
      <p class="text-xs text-slate-400 flex items-center gap-1 mt-1 truncate">
        <span class="opacity-60">by</span>
        <span class="text-slate-300 font-medium">{asset.author || 'Unknown'}</span>
      </p>
    </div>

    <!-- Tags Row -->
    {#if asset.tags && asset.tags.length > 0}
      <div class="flex flex-wrap gap-1 items-center">
        {#each asset.tags.slice(0, 3) as tag}
          <span class="text-[10px] text-slate-400 bg-dark-900 px-1.5 py-0.5 rounded-md border border-dark-700/70 truncate max-w-[90px]">
            #{tag}
          </span>
        {/each}
        {#if asset.tags.length > 3}
          <span class="text-[10px] text-slate-500">+{asset.tags.length - 3}</span>
        {/if}
      </div>
    {/if}

    <!-- Bottom Actions / Quick Add -->
    <div class="pt-2 border-t border-dark-700/50 flex items-center justify-between text-xs text-slate-400">
      {#if asset.source_url}
        <span class="flex items-center gap-1 text-[11px] text-brand-400 hover:text-brand-300 truncate max-w-[120px]" title={asset.source_url}>
          <Globe class="w-3 h-3 flex-shrink-0" />
          <span class="truncate">Source</span>
        </span>
      {:else}
        <span class="text-[11px] text-slate-500 truncate">{asset.source_package || 'Local'}</span>
      {/if}

      {#if onQuickAdd}
        <button
          class="px-2.5 py-1 text-xs font-semibold text-white bg-brand-600 hover:bg-brand-500 rounded-lg transition-colors flex items-center gap-1 shadow-sm"
          on:click|stopPropagation={() => onQuickAdd && onQuickAdd(asset)}
        >
          <span>+ Add</span>
        </button>
      {/if}
    </div>
  </div>
</div>
