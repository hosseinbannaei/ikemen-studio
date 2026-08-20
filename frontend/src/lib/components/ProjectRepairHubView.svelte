<script lang="ts">
  import { onMount } from 'svelte';
  import { projectStore } from '../stores/projectStore';
  import { engineStore } from '../stores/engineStore';
  import type {
    ProjectDiffSummary,
    ConfigInspectionResult,
    VerificationReport,
  } from '../types';
  import {
    Wrench,
    ShieldAlert,
    ShieldCheck,
    CheckCircle2,
    AlertTriangle,
    FileText,
    Play,
    RotateCcw,
    Folder,
    FolderOpen,
    Sliders,
    Sparkles,
    Check,
    Loader2,
    Layers,
    ArrowLeft,
    RefreshCw,
    X,
    Cpu,
  } from 'lucide-svelte';

  export let onBackToWorkspace: () => void;

  let activeTab: 'config' | 'sync' | 'logs' = 'config';

  // Config Doctor State
  let configInspection: ConfigInspectionResult | null = null;
  let loadingConfig = false;
  let fixingConfig = false;

  // Granular Sync State
  let diffSummary: ProjectDiffSummary | null = null;
  let loadingDiff = false;
  let syncingAssets = false;
  let lastSyncReport: VerificationReport | null = null;

  let syncStockChars = true;
  let syncStockStages = false;
  let syncScreenpack = true;
  let syncFonts = false;
  let syncRuntime = true;
  let syncResetConfig = true;

  // Logs State
  let logContent = '';
  let loadingLogs = false;

  onMount(async () => {
    await refreshAll();
  });

  async function refreshAll() {
    await loadConfigDoctor();
    await loadDiffSummary();
    await loadLogsViewer();
  }

  async function loadConfigDoctor() {
    loadingConfig = true;
    configInspection = await projectStore.inspectConfig();
    loadingConfig = false;
  }

  async function loadDiffSummary() {
    loadingDiff = true;
    diffSummary = await projectStore.inspectDiff();
    loadingDiff = false;
  }

  async function loadLogsViewer() {
    loadingLogs = true;
    logContent = await projectStore.getLogs();
    loadingLogs = false;
  }

  async function handleAutoRepairConfig() {
    fixingConfig = true;
    const ok = await projectStore.repairConfig();
    fixingConfig = false;
    if (ok) {
      await loadConfigDoctor();
    }
  }

  async function handleResetConfig() {
    fixingConfig = true;
    const ok = await projectStore.resetConfig();
    fixingConfig = false;
    if (ok) {
      await loadConfigDoctor();
    }
  }

  async function handlePerformSync() {
    if (!$projectStore.current) return;
    syncingAssets = true;
    lastSyncReport = await projectStore.syncAssets({
      projectDir: $projectStore.current.path,
      syncStockChars,
      syncStockStages,
      syncScreenpack,
      syncFonts,
      syncRuntime,
      resetConfig: syncResetConfig,
    });
    syncingAssets = false;
    await refreshAll();
  }

  async function handleRunStandardVerify() {
    const report = await projectStore.verifyAndRepair();
    lastSyncReport = report;
    await loadLogsViewer();
  }

  function handleRelaunch() {
    projectStore.dismissCrash();
    projectStore.launch();
  }
</script>

<div class="p-8 max-w-6xl mx-auto space-y-6">
  <!-- Top Hub Banner -->
  <div class="p-6 rounded-2xl bg-dark-800 border border-dark-600/70 shadow-lg relative overflow-hidden flex flex-col md:flex-row md:items-center justify-between gap-6">
    <div class="space-y-2">
      <div class="flex items-center gap-3 flex-wrap">
        <button
          type="button"
          class="p-2 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600 text-slate-300 hover:text-white transition"
          on:click={onBackToWorkspace}
          title="Back to Workspace"
        >
          <ArrowLeft class="w-4 h-4" />
        </button>

        <div class="w-10 h-10 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-400 flex items-center justify-center flex-shrink-0">
          <Wrench class="w-5 h-5" />
        </div>

        <div>
          <div class="flex items-center gap-2.5">
            <h1 class="text-2xl font-black text-slate-100">Maintenance & Repair Hub</h1>
            {#if $projectStore.current}
              <span class="text-xs font-mono px-2.5 py-0.5 rounded-full bg-purple-500/10 text-purple-300 border border-purple-500/30 font-semibold">
                {$projectStore.current.engine.version}
              </span>
            {/if}
          </div>
          <p class="text-xs text-slate-400 mt-0.5">
            Diagnose crashes, repair config.ini, resolve legacy syntax errors, and sync clean engine assets
          </p>
        </div>
      </div>
    </div>

    <!-- Quick Action Launch -->
    <div class="flex items-center gap-2.5">
      <button
        type="button"
        class="px-4 py-2.5 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600 text-slate-200 text-xs font-semibold flex items-center gap-2 transition shadow-sm"
        on:click={refreshAll}
      >
        <RefreshCw class="w-4 h-4 text-indigo-400 {loadingConfig || loadingDiff ? 'animate-spin' : ''}" />
        <span>Refresh</span>
      </button>

      <button
        type="button"
        class="px-5 py-2.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-bold shadow-md shadow-emerald-950/50 flex items-center gap-2 transition"
        on:click={handleRelaunch}
      >
        <Play class="w-4 h-4 fill-current" />
        <span>Test / Run Game</span>
      </button>
    </div>
  </div>

  <!-- Navigation Tabs -->
  <div class="flex items-center gap-3 border-b border-dark-600/60 pb-3">
    <button
      type="button"
      class="pb-2 text-xs font-bold tracking-wide border-b-2 transition flex items-center gap-2 {
        activeTab === 'config'
          ? 'border-amber-500 text-amber-400'
          : 'border-transparent text-slate-400 hover:text-slate-200'
      }"
      on:click={() => (activeTab = 'config')}
    >
      <Sliders class="w-4 h-4" />
      <span>Config Doctor & Settings</span>
      {#if configInspection && !configInspection.isValid}
        <span class="w-2 h-2 rounded-full bg-rose-500 animate-pulse"></span>
      {/if}
    </button>

    <button
      type="button"
      class="pb-2 text-xs font-bold tracking-wide border-b-2 transition flex items-center gap-2 {
        activeTab === 'sync'
          ? 'border-indigo-500 text-indigo-400'
          : 'border-transparent text-slate-400 hover:text-slate-200'
      }"
      on:click={() => (activeTab = 'sync')}
    >
      <Sparkles class="w-4 h-4" />
      <span>Asset Replacement Wizard</span>
      {#if diffSummary && diffSummary.totalDiscrepancies > 0}
        <span class="px-1.5 py-0.2 rounded-full bg-indigo-500/20 text-indigo-300 font-mono text-[10px]">
          {diffSummary.totalDiscrepancies}
        </span>
      {/if}
    </button>

    <button
      type="button"
      class="pb-2 text-xs font-bold tracking-wide border-b-2 transition flex items-center gap-2 {
        activeTab === 'logs'
          ? 'border-purple-500 text-purple-400'
          : 'border-transparent text-slate-400 hover:text-slate-200'
      }"
      on:click={() => (activeTab = 'logs')}
    >
      <FileText class="w-4 h-4" />
      <span>Audit & Crash Logs</span>
    </button>
  </div>

  <!-- TAB 1: Config Doctor -->
  {#if activeTab === 'config'}
    <div class="space-y-6 animate-in fade-in duration-150">
      <!-- Config Health Status Banner -->
      <div class="p-5 rounded-2xl bg-dark-800 border {configInspection?.isValid ? 'border-emerald-500/40 bg-emerald-950/10' : 'border-amber-500/40 bg-amber-950/10'} shadow-sm space-y-3">
        <div class="flex items-start justify-between gap-4">
          <div class="flex items-start gap-3">
            {#if configInspection?.isValid}
              <CheckCircle2 class="w-5 h-5 text-emerald-400 flex-shrink-0 mt-0.5" />
              <div>
                <h3 class="text-sm font-bold text-slate-100">Configuration is Valid & Compatible</h3>
                <p class="text-xs text-slate-400 mt-0.5">Your save/config.ini parameters match modern Ikemen GO engine requirements.</p>
              </div>
            {:else}
              <AlertTriangle class="w-5 h-5 text-amber-400 flex-shrink-0 mt-0.5" />
              <div>
                <h3 class="text-sm font-bold text-slate-100">Incompatible or Obsolete Config Keys Detected</h3>
                <p class="text-xs text-slate-400 mt-0.5">Found parameters in save/config.ini that may cause OpenGL / display crashes.</p>
              </div>
            {/if}
          </div>

          <div class="flex items-center gap-2">
            <button
              type="button"
              disabled={fixingConfig}
              class="px-4 py-2 rounded-xl bg-amber-600 hover:bg-amber-500 disabled:opacity-50 text-white text-xs font-bold shadow-md flex items-center gap-1.5 transition"
              on:click={handleAutoRepairConfig}
              title="Fix invalid RenderMode and normalize config values"
            >
              {#if fixingConfig}
                <Loader2 class="w-3.5 h-3.5 animate-spin" />
                <span>Fixing...</span>
              {:else}
                <Wrench class="w-3.5 h-3.5" />
                <span>Auto-Fix Incompatible Keys</span>
              {/if}
            </button>

            <button
              type="button"
              disabled={fixingConfig}
              class="px-4 py-2 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600 text-slate-200 text-xs font-semibold flex items-center gap-1.5 transition"
              on:click={handleResetConfig}
              title="Restore clean engine default config.ini"
            >
              <RotateCcw class="w-3.5 h-3.5 text-rose-400" />
              <span>Reset to Defaults</span>
            </button>
          </div>
        </div>

        <!-- Detected Issues List -->
        {#if configInspection && configInspection.issues.length > 0}
          <div class="pt-3 border-t border-dark-600/40 space-y-2">
            <div class="text-[11px] font-bold uppercase tracking-wider text-slate-400">Identified Parameter Issues:</div>
            <div class="space-y-1.5">
              {#each configInspection.issues as issue}
                <div class="p-3 rounded-xl bg-dark-900 border border-dark-600/60 flex items-start justify-between gap-3 text-xs">
                  <div class="space-y-0.5">
                    <div class="font-mono font-bold text-rose-300">
                      {issue.key} = <span class="line-through opacity-70">{issue.currentValue}</span> &rarr; <span class="text-emerald-400">{issue.suggestedValue}</span>
                    </div>
                    <p class="text-[11px] text-slate-400">{issue.description}</p>
                  </div>
                  <span class="px-2 py-0.5 rounded text-[10px] font-bold uppercase font-mono {issue.severity === 'error' ? 'bg-rose-500/20 text-rose-300' : 'bg-amber-500/20 text-amber-300'}">
                    {issue.severity}
                  </span>
                </div>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    </div>

  <!-- TAB 2: Asset Replacement Wizard -->
  {:else if activeTab === 'sync'}
    <div class="space-y-6 animate-in fade-in duration-150">
      <!-- Info Banner -->
      <div class="p-4 rounded-2xl bg-indigo-950/30 border border-indigo-500/30 flex items-start gap-3">
        <Sparkles class="w-5 h-5 text-indigo-400 flex-shrink-0 mt-0.5" />
        <div class="space-y-1 text-xs">
          <h3 class="font-bold text-slate-100">Selective Engine Asset Synchronizer</h3>
          <p class="text-slate-300 leading-relaxed">
            Legacy projects often contain outdated stock files (like 0.99 <code class="text-purple-300">chars/kfm_zss/</code>) that crash modern ZSS parsers. Select components below to replace them with clean modern versions from your engine.
          </p>
          <p class="text-emerald-400 font-semibold pt-1">
            ✓ Your custom fighters, stages, music, and roster (<code class="text-emerald-300">data/select.def</code>) are strictly protected and will NEVER be deleted.
          </p>
        </div>
      </div>

      <!-- Component Replacement Options -->
      <div class="p-6 rounded-2xl bg-dark-800 border border-dark-600/70 space-y-4 shadow-sm">
        <div class="flex items-center justify-between">
          <span class="text-xs font-bold uppercase tracking-wider text-slate-300">Select Components to Synchronize & Replace</span>
          <button
            type="button"
            class="px-5 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold shadow-md shadow-indigo-950/50 flex items-center gap-2 transition"
            disabled={syncingAssets}
            on:click={handlePerformSync}
          >
            {#if syncingAssets}
              <Loader2 class="w-3.5 h-3.5 animate-spin" />
              <span>Synchronizing...</span>
            {:else}
              <Check class="w-3.5 h-3.5" />
              <span>Sync & Replace Selected</span>
            {/if}
          </button>
        </div>

        <div class="space-y-3">
          <!-- 1. Stock Characters -->
          <label class="p-4 rounded-xl border flex items-start gap-3.5 cursor-pointer transition {syncStockChars ? 'bg-dark-900 border-indigo-500/60 ring-1 ring-indigo-500/20' : 'bg-dark-900/60 border-dark-600/60'}">
            <input type="checkbox" bind:checked={syncStockChars} class="rounded border-dark-600 bg-dark-800 text-indigo-600 mt-1" />
            <div class="space-y-1 flex-1">
              <div class="flex items-center justify-between">
                <span class="text-xs font-bold text-slate-100">Default Characters (KFM & KFM ZSS)</span>
                <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-purple-500/10 text-purple-300 border border-purple-500/30">Fixes kfm.zss:659 crash</span>
              </div>
              <p class="text-[11px] text-slate-400">Replaces outdated 0.99 stock Kung Fu Man characters with modern ZSS syntax versions from the engine.</p>
            </div>
          </label>

          <!-- 2. Screenpack & System Scripts -->
          <label class="p-4 rounded-xl border flex items-start gap-3.5 cursor-pointer transition {syncScreenpack ? 'bg-dark-900 border-indigo-500/60 ring-1 ring-indigo-500/20' : 'bg-dark-900/60 border-dark-600/60'}">
            <input type="checkbox" bind:checked={syncScreenpack} class="rounded border-dark-600 bg-dark-800 text-indigo-600 mt-1" />
            <div class="space-y-1 flex-1">
              <div class="flex items-center justify-between">
                <span class="text-xs font-bold text-slate-100">System Scripts & Screenpack (data/common1.cns.zss, dgl.zss, fight.def)</span>
                <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-300 border border-emerald-500/30">Recommended</span>
              </div>
              <p class="text-[11px] text-slate-400">Replaces core fight definitions and common ZSS scripts while preserving your roster (select.def).</p>
            </div>
          </label>

          <!-- 3. Engine Core Lua Runtime -->
          <label class="p-4 rounded-xl border flex items-start gap-3.5 cursor-pointer transition {syncRuntime ? 'bg-dark-900 border-indigo-500/60 ring-1 ring-indigo-500/20' : 'bg-dark-900/60 border-dark-600/60'}">
            <input type="checkbox" bind:checked={syncRuntime} class="rounded border-dark-600 bg-dark-800 text-indigo-600 mt-1" />
            <div class="space-y-1 flex-1">
              <span class="text-xs font-bold text-slate-100">Engine Lua VM & Shaders (external/ & lib/)</span>
              <p class="text-[11px] text-slate-400">Restores clean Lua script execution environment and modern display shaders.</p>
            </div>
          </label>

          <!-- 4. Default Stages -->
          <label class="p-4 rounded-xl border flex items-start gap-3.5 cursor-pointer transition {syncStockStages ? 'bg-dark-900 border-indigo-500/60 ring-1 ring-indigo-500/20' : 'bg-dark-900/60 border-dark-600/60'}">
            <input type="checkbox" bind:checked={syncStockStages} class="rounded border-dark-600 bg-dark-800 text-indigo-600 mt-1" />
            <div class="space-y-1 flex-1">
              <span class="text-xs font-bold text-slate-100">Default Stages (stages/stage0.def)</span>
              <p class="text-[11px] text-slate-400">Ensures standard training and demo stages exist.</p>
            </div>
          </label>

          <!-- 5. System Fonts -->
          <label class="p-4 rounded-xl border flex items-start gap-3.5 cursor-pointer transition {syncFonts ? 'bg-dark-900 border-indigo-500/60 ring-1 ring-indigo-500/20' : 'bg-dark-900/60 border-dark-600/60'}">
            <input type="checkbox" bind:checked={syncFonts} class="rounded border-dark-600 bg-dark-800 text-indigo-600 mt-1" />
            <div class="space-y-1 flex-1">
              <span class="text-xs font-bold text-slate-100">Engine Fonts (font/)</span>
              <p class="text-[11px] text-slate-400">Restores engine UI and bitmap font assets.</p>
            </div>
          </label>

          <!-- 6. Reset Config -->
          <label class="p-4 rounded-xl border flex items-start gap-3.5 cursor-pointer transition {syncResetConfig ? 'bg-dark-900 border-indigo-500/60 ring-1 ring-indigo-500/20' : 'bg-dark-900/60 border-dark-600/60'}">
            <input type="checkbox" bind:checked={syncResetConfig} class="rounded border-dark-600 bg-dark-800 text-indigo-600 mt-1" />
            <div class="space-y-1 flex-1">
              <span class="text-xs font-bold text-slate-100">Reset save/config.ini to Clean Engine Defaults</span>
              <p class="text-[11px] text-slate-400">Eliminates legacy render mode syntax and corrupted display resolutions.</p>
            </div>
          </label>
        </div>
      </div>
    </div>

  <!-- TAB 3: Audit & Crash Logs -->
  {:else if activeTab === 'logs'}
    <div class="space-y-6 animate-in fade-in duration-150">
      <div class="p-6 rounded-2xl bg-dark-800 border border-dark-600/70 space-y-4 shadow-sm">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <FileText class="w-4 h-4 text-purple-400" />
            <span class="text-xs font-bold uppercase tracking-wider text-slate-300">Live Engine & Diagnostic Logs</span>
          </div>

          <div class="flex items-center gap-2">
            <button
              type="button"
              class="px-3 py-1.5 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600 text-slate-200 text-xs font-semibold flex items-center gap-1.5 transition"
              on:click={loadLogsViewer}
            >
              <RefreshCw class="w-3.5 h-3.5 {loadingLogs ? 'animate-spin' : ''}" />
              <span>Refresh Log</span>
            </button>

            <button
              type="button"
              class="px-3 py-1.5 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600 text-slate-200 text-xs font-semibold flex items-center gap-1.5 transition"
              on:click={() => projectStore.openLogs()}
            >
              <FolderOpen class="w-3.5 h-3.5 text-amber-400" />
              <span>Open Logs Folder</span>
            </button>
          </div>
        </div>

        <div class="p-4 rounded-xl bg-dark-900 border border-dark-600/60 font-mono text-xs text-slate-300 whitespace-pre-wrap max-h-96 overflow-y-auto leading-relaxed shadow-inner">
          {logContent}
        </div>
      </div>
    </div>
  {/if}
</div>
