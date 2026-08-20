<script lang="ts">
  import { onMount } from 'svelte';
  import { engineStore } from '../stores/engineStore';
  import { settingsStore } from '../stores/settingsStore';
  import { OpenFolderInExplorer } from '../../../wailsjs/go/main/App';
  import {
    Download,
    Trash2,
    RefreshCw,
    CheckCircle2,
    Layers,
    ExternalLink,
    ChevronDown,
    ChevronUp,
    HardDrive,
    XCircle,
    Loader2,
    Sparkles,
    Folder,
    FolderOpen,
  } from 'lucide-svelte';

  let activeSubTab: 'installed' | 'available' = 'installed';
  let filterChannel: 'all' | 'stable' | 'nightly' = 'all';
  let expandedRelease: string | null = null;

  onMount(() => {
    settingsStore.load();
    engineStore.loadInstalled();
    engineStore.loadAvailable();
  });

  function handleOpenEnginesFolder() {
    if ($settingsStore.enginesDir) {
      OpenFolderInExplorer($settingsStore.enginesDir);
    }
  }

  function isInstalled(tag: string): boolean {
    const download = $engineStore.downloads[tag];
    if (download && (download.status === 'downloading' || download.status === 'extracting')) {
      return false;
    }
    return $engineStore.installed.some(
      (e) => e.version.toLowerCase() === tag.toLowerCase()
    );
  }

  function formatBytes(bytes: number): string {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }

  function formatDate(dateStr: string): string {
    if (!dateStr) return '';
    try {
      return new Date(dateStr).toLocaleDateString(undefined, {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
      });
    } catch {
      return dateStr;
    }
  }

  $: filteredAvailable = $engineStore.available.filter((rel) => {
    if (filterChannel === 'stable') return !rel.isPrerelease && !rel.tag.toLowerCase().includes('nightly');
    if (filterChannel === 'nightly') return rel.isPrerelease || rel.tag.toLowerCase().includes('nightly');
    return true;
  });
</script>

<div class="p-8 max-w-6xl mx-auto space-y-6">
  <!-- Header -->
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
    <div>
      <h1 class="text-2xl font-black tracking-tight text-slate-100">Engine Versions</h1>
      <p class="text-xs text-slate-400 mt-0.5">Discover, download, and manage local Ikemen GO engine runtimes</p>
    </div>

    <div class="flex items-center gap-2.5">
      <button
        type="button"
        class="flex items-center gap-2 px-3.5 py-2 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600/80 text-slate-200 text-xs font-semibold shadow-sm transition"
        on:click={handleOpenEnginesFolder}
        title="Open local engines storage folder"
      >
        <FolderOpen class="w-4 h-4 text-amber-400" />
        <span>Open Engines Folder</span>
      </button>

      <button
        type="button"
        class="flex items-center gap-2 px-3.5 py-2 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600/80 text-slate-200 text-xs font-semibold shadow-sm transition"
        on:click={() => {
          engineStore.loadInstalled();
          engineStore.loadAvailable();
        }}
      >
        <RefreshCw class="w-4 h-4 {$engineStore.loadingAvailable || $engineStore.loadingInstalled ? 'animate-spin text-indigo-400' : 'text-indigo-400'}" />
        <span>Refresh Releases</span>
      </button>
    </div>
  </div>

  <!-- Sub Tabs & Channel Filters -->
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-dark-600/60 pb-3">
    <div class="flex gap-4">
      <button
        type="button"
        class="pb-2 text-xs font-bold tracking-wide border-b-2 transition {
          activeSubTab === 'installed'
            ? 'border-indigo-500 text-indigo-400'
            : 'border-transparent text-slate-400 hover:text-slate-200'
        }"
        on:click={() => (activeSubTab = 'installed')}
      >
        Installed Runtimes ({$engineStore.installed.length})
      </button>
      <button
        type="button"
        class="pb-2 text-xs font-bold tracking-wide border-b-2 transition {
          activeSubTab === 'available'
            ? 'border-indigo-500 text-indigo-400'
            : 'border-transparent text-slate-400 hover:text-slate-200'
        }"
        on:click={() => (activeSubTab = 'available')}
      >
        GitHub Releases ({$engineStore.available.length})
      </button>
    </div>

    {#if activeSubTab === 'available'}
      <div class="flex items-center gap-1.5">
        <button
          type="button"
          class="px-2.5 py-1 text-[11px] font-semibold rounded-lg transition {filterChannel === 'all' ? 'bg-indigo-600 text-white shadow-sm' : 'bg-dark-700 text-slate-400 hover:text-slate-200'}"
          on:click={() => (filterChannel = 'all')}
        >
          All
        </button>
        <button
          type="button"
          class="px-2.5 py-1 text-[11px] font-semibold rounded-lg transition {filterChannel === 'stable' ? 'bg-indigo-600 text-white shadow-sm' : 'bg-dark-700 text-slate-400 hover:text-slate-200'}"
          on:click={() => (filterChannel = 'stable')}
        >
          Stable
        </button>
        <button
          type="button"
          class="px-2.5 py-1 text-[11px] font-semibold rounded-lg transition {filterChannel === 'nightly' ? 'bg-indigo-600 text-white shadow-sm' : 'bg-dark-700 text-slate-400 hover:text-slate-200'}"
          on:click={() => (filterChannel = 'nightly')}
        >
          Nightly / Pre
        </button>
      </div>
    {/if}
  </div>

  <!-- Content List -->
  {#if activeSubTab === 'installed'}
    {#if $engineStore.loadingInstalled && $engineStore.installed.length === 0}
      <div class="py-16 text-center text-slate-400 flex flex-col items-center gap-2">
        <RefreshCw class="w-6 h-6 animate-spin text-indigo-400" />
        <span class="text-xs">Scanning engines directory...</span>
      </div>
    {:else if $engineStore.installed.length === 0}
      <div class="py-16 text-center text-slate-400 flex flex-col items-center gap-3 bg-dark-800/40 rounded-2xl border border-dark-600/40 p-8">
        <HardDrive class="w-10 h-10 text-slate-600" />
        <div class="space-y-1">
          <div class="text-sm font-bold text-slate-300">No engines installed yet</div>
          <div class="text-xs text-slate-500">Switch to the "GitHub Releases" tab to download an Ikemen GO version.</div>
        </div>
        <button
          type="button"
          class="mt-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold rounded-xl shadow-md transition"
          on:click={() => (activeSubTab = 'available')}
        >
          Browse Releases
        </button>
      </div>
    {:else}
      <div class="grid grid-cols-1 gap-3">
        {#each $engineStore.installed as engine (engine.version)}
          <div class="p-5 rounded-2xl bg-dark-800/80 border border-dark-600/60 flex flex-col sm:flex-row sm:items-center justify-between gap-4 shadow-sm hover:border-dark-600 transition">
            <div class="space-y-1">
              <div class="flex items-center gap-2.5 flex-wrap">
                <span class="font-bold text-base text-slate-100">{engine.version}</span>
                <span
                  class="text-[10px] font-mono uppercase px-2 py-0.5 rounded-full font-semibold {
                    engine.channel === 'stable'
                      ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                      : 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                  }"
                >
                  {engine.channel}
                </span>
                {#if engine.executablePath}
                  <span class="text-[11px] text-emerald-400 flex items-center gap-1 font-semibold">
                    <CheckCircle2 class="w-3.5 h-3.5" /> Ready
                  </span>
                {/if}
              </div>
              <div class="text-xs text-slate-400 font-mono flex items-center gap-3 flex-wrap">
                <span>Size: {formatBytes(engine.size)}</span>
                <span>&bull;</span>
                <span class="truncate max-w-lg">Path: {engine.path}</span>
              </div>
            </div>

            <div class="flex items-center gap-2 self-end sm:self-center">
              <button
                type="button"
                class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold text-rose-400 hover:bg-rose-950/40 hover:border-rose-800/50 border border-transparent transition"
                on:click={() => {
                  if (confirm(`Are you sure you want to delete engine version ${engine.version}?`)) {
                    engineStore.removeEngine(engine.version);
                  }
                }}
              >
                <Trash2 class="w-3.5 h-3.5" />
                Delete
              </button>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  {:else}
    <!-- Available Releases Tab -->
    {#if $engineStore.loadingAvailable && $engineStore.available.length === 0}
      <div class="py-16 text-center text-slate-400 flex flex-col items-center gap-2">
        <RefreshCw class="w-6 h-6 animate-spin text-indigo-400" />
        <span class="text-xs">Fetching releases from GitHub...</span>
      </div>
    {:else if filteredAvailable.length === 0}
      <div class="py-16 text-center text-slate-400">No releases found matching the filter.</div>
    {:else}
      <div class="grid grid-cols-1 gap-3">
        {#each filteredAvailable as release (release.tag)}
          {@const download = $engineStore.downloads[release.tag]}
          {@const isDownloading = download && (download.status === 'downloading' || download.status === 'extracting')}
          {@const installed = isInstalled(release.tag)}
          <div class="p-5 rounded-2xl bg-dark-800/80 border border-dark-600/60 space-y-3.5 shadow-sm">
            <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
              <div class="space-y-1">
                <div class="flex items-center gap-2 flex-wrap">
                  <span class="font-bold text-base text-slate-100">{release.tag}</span>
                  {#if release.name && release.name !== release.tag}
                    <span class="text-xs text-slate-300 font-semibold">{release.name}</span>
                  {/if}
                  <span
                    class="text-[10px] font-mono uppercase px-2 py-0.5 rounded-full font-semibold {
                      release.isPrerelease || release.tag.toLowerCase().includes('nightly')
                        ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                        : 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                    }"
                  >
                    {release.isPrerelease || release.tag.toLowerCase().includes('nightly') ? 'Nightly / Pre' : 'Stable'}
                  </span>
                  {#if installed}
                    <span class="text-[10px] px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 font-semibold">
                      Installed
                    </span>
                  {/if}
                </div>
                <div class="text-xs text-slate-400 flex items-center gap-2">
                  <span>Published {formatDate(release.publishedAt)}</span>
                  {#if release.htmlUrl}
                    <span>&bull;</span>
                    <a
                      href={release.htmlUrl}
                      target="_blank"
                      rel="noreferrer"
                      class="text-indigo-400 hover:underline inline-flex items-center gap-1"
                    >
                      GitHub <ExternalLink class="w-3 h-3" />
                    </a>
                  {/if}
                </div>
              </div>

              <!-- Action Controls -->
              <div class="self-end sm:self-auto">
                {#if isDownloading}
                  <div class="flex items-center gap-2">
                    <div class="flex items-center gap-2 px-3 py-1.5 rounded-xl bg-indigo-950/60 text-indigo-300 text-xs font-semibold border border-indigo-700/40 shadow-sm">
                      <Loader2 class="w-3.5 h-3.5 animate-spin text-indigo-400" />
                      <span class="capitalize">{download.status}... {Math.round(download.percent)}%</span>
                    </div>
                    <button
                      type="button"
                      class="flex items-center gap-1 px-3 py-1.5 rounded-xl bg-rose-950/60 hover:bg-rose-900 border border-rose-800/40 text-rose-300 text-xs font-bold transition"
                      title="Cancel Download"
                      on:click={() => engineStore.cancelDownload(release.tag)}
                    >
                      <XCircle class="w-3.5 h-3.5" />
                      <span>Cancel</span>
                    </button>
                  </div>
                {:else if installed}
                  <div class="flex items-center gap-1.5 px-3.5 py-1.5 rounded-xl bg-dark-700/60 text-emerald-400 text-xs font-bold border border-dark-600/40">
                    <CheckCircle2 class="w-4 h-4" /> Ready
                  </div>
                {:else}
                  <button
                    type="button"
                    class="flex items-center gap-1.5 px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold shadow-md transition"
                    on:click={() => engineStore.startDownload(release.tag)}
                  >
                    <Download class="w-3.5 h-3.5" />
                    Download
                  </button>
                {/if}
              </div>
            </div>

            <!-- Progress Bar -->
            {#if isDownloading}
              <div class="space-y-1">
                <div class="w-full bg-dark-700 rounded-full h-2 overflow-hidden shadow-inner">
                  <div
                    class="bg-gradient-to-r from-indigo-500 to-purple-500 h-2 rounded-full transition-all duration-200"
                    style="width: {download.percent}%"
                  ></div>
                </div>
                <div class="flex justify-between text-[11px] text-slate-400 font-mono">
                  <span>{download.status === 'downloading' ? 'Downloading archive package...' : 'Extracting & configuring binary...'}</span>
                  <span>{Math.round(download.percent)}%</span>
                </div>
              </div>
            {/if}

            <!-- Release Notes -->
            {#if release.body}
              <div>
                <button
                  type="button"
                  class="text-xs text-slate-400 hover:text-slate-200 flex items-center gap-1 font-medium"
                  on:click={() => (expandedRelease = expandedRelease === release.tag ? null : release.tag)}
                >
                  {#if expandedRelease === release.tag}
                    <ChevronUp class="w-3.5 h-3.5" /> Hide Release Notes
                  {:else}
                    <ChevronDown class="w-3.5 h-3.5" /> Show Release Notes
                  {/if}
                </button>
                {#if expandedRelease === release.tag}
                  <div class="mt-2 p-3.5 rounded-xl bg-dark-900/80 text-xs text-slate-300 max-h-48 overflow-y-auto whitespace-pre-wrap font-mono leading-relaxed border border-dark-600/40">
                    {release.body}
                  </div>
                {/if}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  {/if}
</div>
