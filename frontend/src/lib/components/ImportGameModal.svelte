<script lang="ts">
  import { onMount } from 'svelte';
  import { engineStore } from '../stores/engineStore';
  import { projectStore } from '../stores/projectStore';
  import { vaultStore } from '../stores/vaultStore';
  import type { ExistingGameInspection, ImportOptions } from '../types';
  import { SelectDirectoryDialog } from '../../../wailsjs/go/main/App';
  import {
    FolderDown,
    ShieldCheck,
    Folder,
    Users,
    Mountain,
    FileCode,
    X,
    Loader2,
    CheckCircle2,
    Sparkles,
    GitBranch,
    History,
    Layers,
    Download,
    Check,
    Music,
    Type,
    AlertTriangle,
    Package,
  } from 'lucide-svelte';

  export let inspection: ExistingGameInspection;
  export let onClose: () => void;
  export let onSuccess: () => void;

  let mode: 'rebuild' | 'diff_upgrade' | 'legacy_match' = 'rebuild';

  let projectName = inspection.detectedName || 'My Imported Game';
  let targetDir = `${inspection.sourcePath}_studio`;
  let author = '';
  let importing = false;

  // Vault auto-ingestion options
  let registerInVault = true;
  let targetVaultId = 'vault-default';

  // Selected engine version for the chosen mode
  let selectedEngine = $engineStore.installed[0]?.version || 'nightly';

  // Interactive Checklist for Clean Rebuild Mode
  let includeChars = true;
  let includeStages = true;
  let includeRoster = true;
  let includeSound = true;
  let includeFonts = true;
  let includeLegacySystem = false;

  onMount(async () => {
    await engineStore.loadInstalled();
    await engineStore.loadAvailable();
    await vaultStore.loadVaults();

    // Smart engine matching
    if (inspection.detectedEngineVersion) {
      const match = $engineStore.installed.find(
        (e) => e.version.toLowerCase() === inspection.detectedEngineVersion?.toLowerCase()
      );
      if (match) {
        selectedEngine = match.version;
      } else {
        const nightlyInstalled = $engineStore.installed.find((e) => e.version.includes('nightly'));
        if (nightlyInstalled) {
          selectedEngine = nightlyInstalled.version;
        }
      }
    } else {
      const nightlyInstalled = $engineStore.installed.find((e) => e.version.includes('nightly'));
      if (nightlyInstalled) {
        selectedEngine = nightlyInstalled.version;
      }
    }
  });

  $: detectedLegacyVer = inspection.detectedEngineVersion || 'v0.99.0';

  // When switching to legacy_match, auto-switch selected engine to detected legacy version
  function setMode(newMode: 'rebuild' | 'diff_upgrade' | 'legacy_match') {
    mode = newMode;
    if (newMode === 'legacy_match') {
      selectedEngine = detectedLegacyVer;
    } else if (newMode === 'rebuild' || newMode === 'diff_upgrade') {
      const installedNightly = $engineStore.installed.find((e) => e.version.includes('nightly'));
      if (installedNightly) {
        selectedEngine = installedNightly.version;
      } else if ($engineStore.installed[0]) {
        selectedEngine = $engineStore.installed[0].version;
      } else {
        selectedEngine = 'nightly';
      }
    }
  }

  $: isEngineInstalled = $engineStore.installed.some(
    (e) => e.version.toLowerCase() === selectedEngine.toLowerCase()
  );

  $: currentDownload = $engineStore.downloads[selectedEngine];

  async function handleDownloadEngine(tag: string) {
    await engineStore.startDownload(tag);
  }

  async function handleBrowseTarget() {
    try {
      const selected = await SelectDirectoryDialog('Select Destination Directory for Imported Project');
      if (selected) {
        targetDir = selected;
      }
    } catch (e) {
      console.error(e);
    }
  }

  async function handleImport() {
    if (!projectName.trim() || !targetDir.trim()) return;

    importing = true;
    const finalTargetDir = targetDir.trim();
    const opts: ImportOptions = {
      sourceDir: inspection.sourcePath,
      targetDir: finalTargetDir,
      projectName: projectName.trim(),
      engineVersion: selectedEngine,
      author: author.trim(),
      mode,
      includeChars,
      includeStages,
      includeSound,
      includeFonts,
      includeRoster,
      includeLegacySystem,
    };

    const ok = await projectStore.importExistingWithOptions(opts);

    // If requested, auto-ingest into Vault
    if (ok && registerInVault) {
      try {
        const pathsToIngest = [`${finalTargetDir}/chars`, `${finalTargetDir}/stages`];
        await vaultStore.ingestMultiple(pathsToIngest, targetVaultId, 'auto');
      } catch (e) {
        console.warn('Vault auto-ingest notice:', e);
      }
    }

    importing = false;

    if (ok) {
      onSuccess();
    }
  }
</script>

<div class="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
  <div class="bg-dark-800 border border-dark-600/80 rounded-2xl w-full max-w-2xl shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-150 flex flex-col max-h-[92vh]">
    <!-- Header -->
    <div class="p-5 border-b border-dark-600/60 flex items-center justify-between bg-dark-850">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 flex items-center justify-center flex-shrink-0">
          <FolderDown class="w-5 h-5" />
        </div>
        <div>
          <h2 class="text-base font-bold text-slate-100">Import Existing MUGEN / Ikemen Game</h2>
          <p class="text-xs text-slate-400">Migrate legacy folder into a managed Ikemen Studio workspace</p>
        </div>
      </div>
      <button
        type="button"
        disabled={importing}
        class="p-2 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-dark-700/60 transition disabled:opacity-50"
        on:click={onClose}
      >
        <X class="w-5 h-5" />
      </button>
    </div>

    <!-- Content -->
    <div class="p-6 space-y-5 overflow-y-auto flex-1 text-xs">
      <!-- Safe Copy Notice & Detected Version -->
      <div class="p-4 rounded-xl bg-indigo-950/30 border border-indigo-500/30 flex items-center justify-between gap-3">
        <div class="flex items-center gap-3">
          <ShieldCheck class="w-5 h-5 text-indigo-400 flex-shrink-0" />
          <div class="space-y-0.5">
            <div class="font-bold text-slate-200">Non-Destructive Migration</div>
            <div class="text-[11px] text-slate-400">Your original game folder will remain <strong>100% untouched</strong>.</div>
          </div>
        </div>
        <div class="px-3 py-1.5 rounded-lg bg-dark-900/80 border border-dark-600 font-mono text-[11px] text-purple-300">
          Detected: <strong class="text-slate-100">{detectedLegacyVer}</strong>
        </div>
      </div>

      <!-- Discovered Game Stats -->
      <div class="grid grid-cols-3 gap-2.5">
        <div class="p-3 rounded-xl bg-dark-900 border border-dark-600/60 flex items-center gap-2.5">
          <Users class="w-4 h-4 text-cyan-400" />
          <div>
            <div class="text-[10px] text-slate-400">Fighters</div>
            <div class="text-xs font-bold text-slate-100">{inspection.characterCount} found</div>
          </div>
        </div>
        <div class="p-3 rounded-xl bg-dark-900 border border-dark-600/60 flex items-center gap-2.5">
          <Mountain class="w-4 h-4 text-emerald-400" />
          <div>
            <div class="text-[10px] text-slate-400">Stages</div>
            <div class="text-xs font-bold text-slate-100">{inspection.stageCount} found</div>
          </div>
        </div>
        <div class="p-3 rounded-xl bg-dark-900 border border-dark-600/60 flex items-center gap-2.5">
          <FileCode class="w-4 h-4 text-amber-400" />
          <div>
            <div class="text-[10px] text-slate-400">Roster</div>
            <div class="text-xs font-bold text-slate-100">{inspection.hasSelectDef ? 'select.def' : 'default'}</div>
          </div>
        </div>
      </div>

      <!-- Migration Mode Selector -->
      <div class="space-y-2">
        <label class="block text-[11px] font-bold uppercase tracking-wider text-slate-400">Migration Strategy</label>
        <div class="grid grid-cols-3 gap-2.5">
          <!-- Mode 1: Clean Rebuild -->
          <button
            type="button"
            class="p-3 rounded-xl border text-left flex flex-col justify-between gap-2 transition {
              mode === 'rebuild'
                ? 'border-emerald-500 bg-emerald-950/20 text-emerald-300 ring-1 ring-emerald-500/30'
                : 'border-dark-600/80 bg-dark-900/60 text-slate-400 hover:text-slate-200'
            }"
            on:click={() => setMode('rebuild')}
          >
            <div class="flex items-center gap-2">
              <Sparkles class="w-4 h-4 {mode === 'rebuild' ? 'text-emerald-400' : 'text-slate-400'}" />
              <span class="font-bold text-xs">Clean Rebuild</span>
            </div>
            <p class="text-[10px] text-slate-400 leading-tight">
              Scaffold modern engine & port pure assets with checklist. 100% crash-free.
            </p>
          </button>

          <!-- Mode 2: Smart Diff Upgrade -->
          <button
            type="button"
            class="p-3 rounded-xl border text-left flex flex-col justify-between gap-2 transition {
              mode === 'diff_upgrade'
                ? 'border-indigo-500 bg-indigo-950/20 text-indigo-300 ring-1 ring-indigo-500/30'
                : 'border-dark-600/80 bg-dark-900/60 text-slate-400 hover:text-slate-200'
            }"
            on:click={() => setMode('diff_upgrade')}
          >
            <div class="flex items-center gap-2">
              <GitBranch class="w-4 h-4 {mode === 'diff_upgrade' ? 'text-indigo-400' : 'text-slate-400'}" />
              <span class="font-bold text-xs">Smart Upgrade</span>
            </div>
            <p class="text-[10px] text-slate-400 leading-tight">
              Replaces old stock engine files, preserves all user modifications.
            </p>
          </button>

          <!-- Mode 3: Preserve Legacy -->
          <button
            type="button"
            class="p-3 rounded-xl border text-left flex flex-col justify-between gap-2 transition {
              mode === 'legacy_match'
                ? 'border-purple-500 bg-purple-950/20 text-purple-300 ring-1 ring-purple-500/30'
                : 'border-dark-600/80 bg-dark-900/60 text-slate-400 hover:text-slate-200'
            }"
            on:click={() => setMode('legacy_match')}
          >
            <div class="flex items-center gap-2">
              <History class="w-4 h-4 {mode === 'legacy_match' ? 'text-purple-400' : 'text-slate-400'}" />
              <span class="font-bold text-xs">Preserve Legacy</span>
            </div>
            <p class="text-[10px] text-slate-400 leading-tight">
              Pair with clean official {detectedLegacyVer} engine release.
            </p>
          </button>
        </div>
      </div>

      <!-- Checklist (in Rebuild Mode) -->
      {#if mode === 'rebuild'}
        <div class="p-4 rounded-xl bg-dark-900 border border-dark-600/70 space-y-3">
          <span class="block text-[11px] font-bold uppercase tracking-wider text-slate-300">
            Assets Review Checklist
          </span>
          <div class="grid grid-cols-2 gap-2.5">
            <label class="flex items-center gap-2.5 p-2 rounded-lg bg-dark-800 border border-dark-600/50 cursor-pointer">
              <input type="checkbox" bind:checked={includeChars} class="rounded border-dark-600 bg-dark-900 text-indigo-600" />
              <span class="text-xs text-slate-200">Fighters ({inspection.characterCount})</span>
            </label>

            <label class="flex items-center gap-2.5 p-2 rounded-lg bg-dark-800 border border-dark-600/50 cursor-pointer">
              <input type="checkbox" bind:checked={includeStages} class="rounded border-dark-600 bg-dark-900 text-indigo-600" />
              <span class="text-xs text-slate-200">Stages ({inspection.stageCount})</span>
            </label>

            <label class="flex items-center gap-2.5 p-2 rounded-lg bg-dark-800 border border-dark-600/50 cursor-pointer">
              <input type="checkbox" bind:checked={includeRoster} class="rounded border-dark-600 bg-dark-900 text-indigo-600" />
              <span class="text-xs text-slate-200">Roster Order (select.def)</span>
            </label>

            <label class="flex items-center gap-2.5 p-2 rounded-lg bg-dark-800 border border-dark-600/50 cursor-pointer">
              <input type="checkbox" bind:checked={includeSound} class="rounded border-dark-600 bg-dark-900 text-indigo-600" />
              <span class="text-xs text-slate-200">Sound & BGM Tracks</span>
            </label>

            <label class="flex items-center gap-2.5 p-2 rounded-lg bg-dark-800 border border-dark-600/50 cursor-pointer">
              <input type="checkbox" bind:checked={includeFonts} class="rounded border-dark-600 bg-dark-900 text-indigo-600" />
              <span class="text-xs text-slate-200">Custom Fonts</span>
            </label>

            <label class="flex items-center gap-2.5 p-2 rounded-lg bg-dark-800 border border-dark-600/50 cursor-pointer">
              <input type="checkbox" bind:checked={includeLegacySystem} class="rounded border-dark-600 bg-dark-900 text-indigo-600" />
              <span class="text-xs text-slate-300">Legacy System Scripts</span>
            </label>
          </div>
          {#if !includeLegacySystem}
            <div class="text-[10px] text-emerald-400 flex items-center gap-1.5 pt-1">
              <CheckCircle2 class="w-3.5 h-3.5" />
              <span>Studio will supply clean modern engine system files (prevents character selection errors)</span>
            </div>
          {/if}
        </div>
      {/if}

      <!-- Target Engine Selection & Inline Downloader -->
      <div class="p-4 rounded-xl bg-dark-900 border border-dark-600/70 space-y-3">
        <div class="flex items-center justify-between">
          <label class="text-[11px] font-bold uppercase tracking-wider text-slate-300">Target Official Engine</label>
          {#if isEngineInstalled}
            <span class="flex items-center gap-1 text-[11px] font-semibold text-emerald-400">
              <Check class="w-3.5 h-3.5" />
              <span>Installed & Ready</span>
            </span>
          {:else}
            <span class="flex items-center gap-1 text-[11px] font-semibold text-amber-400">
              <AlertTriangle class="w-3.5 h-3.5" />
              <span>Download Required</span>
            </span>
          {/if}
        </div>

        <div class="flex gap-2">
          <select
            bind:value={selectedEngine}
            class="flex-1 bg-dark-800 border border-dark-600 rounded-xl px-3.5 py-2.5 text-xs text-slate-100 font-mono focus:outline-none focus:border-indigo-500"
          >
            {#each $engineStore.available as rel}
              <option value={rel.tag}>
                {rel.tag} ({rel.isPrerelease ? 'nightly/pre' : 'stable'}) {$engineStore.installed.some((e) => e.version === rel.tag) ? '✓' : ''}
              </option>
            {/each}
            {#if $engineStore.available.length === 0}
              {#each $engineStore.installed as eng}
                <option value={eng.version}>{eng.version} (installed)</option>
              {/each}
              <option value="nightly">nightly</option>
              <option value="v0.99.0">v0.99.0</option>
            {/if}
          </select>

          {#if !isEngineInstalled}
            <button
              type="button"
              disabled={currentDownload?.status === 'downloading' || currentDownload?.status === 'extracting'}
              class="px-4 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white font-bold text-xs flex items-center gap-2 transition"
              on:click={() => handleDownloadEngine(selectedEngine)}
            >
              {#if currentDownload?.status === 'downloading' || currentDownload?.status === 'extracting'}
                <Loader2 class="w-3.5 h-3.5 animate-spin" />
                <span>Downloading...</span>
              {:else}
                <Download class="w-3.5 h-3.5" />
                <span>Get Engine</span>
              {/if}
            </button>
          {/if}
        </div>

        {#if currentDownload && (currentDownload.status === 'downloading' || currentDownload.status === 'extracting')}
          <div class="p-3 rounded-lg bg-dark-800 border border-dark-600/50 space-y-1.5">
            <div class="flex justify-between text-[10px] text-slate-400">
              <span class="capitalize font-semibold text-indigo-300">{currentDownload.status}...</span>
              <span class="font-mono text-slate-200">{currentDownload.percent}%</span>
            </div>
            <div class="w-full h-1.5 bg-dark-900 rounded-full overflow-hidden">
              <div class="h-full bg-indigo-500 rounded-full transition-all duration-200" style="width: {currentDownload.percent}%"></div>
            </div>
          </div>
        {/if}
      </div>

      <!-- Project Name & Target Folder Form Inputs -->
      <div class="space-y-3">
        <div>
          <label class="block text-xs font-semibold text-slate-300 mb-1">Project Name</label>
          <input
            type="text"
            bind:value={projectName}
            placeholder="My Fighter Game"
            class="w-full bg-dark-900 border border-dark-600 rounded-xl px-3.5 py-2 text-xs text-slate-100 focus:outline-none focus:border-indigo-500"
          />
        </div>

        <div>
          <label class="block text-xs font-semibold text-slate-300 mb-1">Destination Directory</label>
          <div class="flex gap-2">
            <input
              type="text"
              bind:value={targetDir}
              class="flex-1 bg-dark-900 border border-dark-600 rounded-xl px-3.5 py-2 text-xs text-slate-100 font-mono focus:outline-none focus:border-indigo-500"
            />
            <button
              type="button"
              class="px-3.5 py-2 bg-dark-700 hover:bg-dark-600 border border-dark-600 rounded-xl text-xs font-semibold text-slate-200 flex items-center gap-1.5 transition"
              on:click={handleBrowseTarget}
            >
              <Folder class="w-4 h-4 text-indigo-400" />
              <span>Browse</span>
            </button>
          </div>
        </div>

        <div>
          <label class="block text-xs font-semibold text-slate-300 mb-1">Author (Optional)</label>
          <input
            type="text"
            bind:value={author}
            placeholder="Your Name / Studio"
            class="w-full bg-dark-900 border border-dark-600 rounded-xl px-3.5 py-2 text-xs text-slate-100 focus:outline-none focus:border-indigo-500"
          />
        </div>
      </div>

      <!-- Asset Vault Integration Option -->
      <div class="p-4 rounded-xl bg-dark-900 border border-dark-600/70 space-y-3">
        <label class="flex items-center gap-2.5 cursor-pointer">
          <input
            type="checkbox"
            bind:checked={registerInVault}
            class="rounded border-dark-600 bg-dark-800 text-brand-500 focus:ring-brand-500"
          />
          <div class="space-y-0.5">
            <span class="font-bold text-xs text-slate-200 flex items-center gap-1.5">
              <Package class="w-3.5 h-3.5 text-brand-400" />
              <span>Import Fighters & Stages into Asset Vault</span>
            </span>
            <p class="text-[11px] text-slate-400">
              Extracts SFF portraits, character metadata, and indexes them into your central Vault library.
            </p>
          </div>
        </label>

        {#if registerInVault}
          <div class="pt-2 pl-6 flex items-center gap-3">
            <label class="text-[11px] text-slate-400 font-semibold">Target Vault:</label>
            <select
              bind:value={targetVaultId}
              class="bg-dark-800 border border-dark-600 rounded-xl px-3 py-1.5 text-xs text-slate-200 focus:outline-none focus:border-brand-500"
            >
              {#each $vaultStore.vaults as v}
                <option value={v.id}>{v.name} {v.is_default ? '(Default)' : ''}</option>
              {/each}
              {#if $vaultStore.vaults.length === 0}
                <option value="vault-default">Default Vault</option>
              {/if}
            </select>
          </div>
        {/if}
      </div>
    </div>

    <!-- Footer -->
    <div class="p-4 border-t border-dark-600/60 bg-dark-850 flex items-center justify-between">
      <button
        type="button"
        disabled={importing}
        class="px-4 py-2 rounded-xl text-slate-400 hover:text-slate-200 text-xs font-semibold hover:bg-dark-700/60 transition disabled:opacity-50"
        on:click={onClose}
      >
        Cancel
      </button>

      <button
        type="button"
        disabled={importing || !isEngineInstalled || !projectName.trim() || !targetDir.trim()}
        class="px-6 py-2.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 text-white text-xs font-bold shadow-md shadow-emerald-950/50 flex items-center gap-2 transition"
        on:click={handleImport}
      >
        {#if importing}
          <Loader2 class="w-3.5 h-3.5 animate-spin" />
          <span>Migrating & Scaffolding...</span>
        {:else if !isEngineInstalled}
          <Download class="w-4 h-4" />
          <span>Download Engine to Proceed</span>
        {:else}
          <FolderDown class="w-4 h-4" />
          <span>Import & Adopt Project</span>
        {/if}
      </button>
    </div>
  </div>
</div>
