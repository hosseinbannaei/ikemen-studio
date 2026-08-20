<script lang="ts">
  import { onMount } from 'svelte';
  import { projectStore } from '../stores/projectStore';
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
    FolderOpen,
    Sliders,
    Sparkles,
    Check,
    Loader2,
    ArrowLeft,
    RefreshCw,
    X,
    Activity,
  } from 'lucide-svelte';

  export let onBackToWorkspace: () => void;

  let activeTab: 'analysis' | 'sync' | 'config' | 'logs' = 'analysis';

  // State
  let configInspection: ConfigInspectionResult | null = null;
  let diffSummary: ProjectDiffSummary | null = null;
  let logContent = '';
  let lastSyncReport: VerificationReport | null = null;

  // Loading / Animation States
  let isScanning = false;
  let isExecuting = false;
  let executionStep = '';
  let executionPhaseIndex = 0;
  let showSuccessBanner = false;

  // Sensitive Confirmation Modal
  let showConfirmModal = false;
  let confirmTitle = '';
  let confirmDescription = '';
  let confirmAffectedItems: string[] = [];
  let confirmActionFn: (() => Promise<void>) | null = null;

  // Granular Sync Checkboxes
  let syncStockChars = true;
  let syncStockStages = false;
  let syncScreenpack = true;
  let syncSound = true;
  let syncFonts = false;
  let syncRuntime = true;
  let syncResetConfig = true;

  const executionStepsSequence = [
    'Scanning directory hierarchy...',
    'Comparing against clean engine baseline...',
    'Resolving missing audio & ZSS scripts...',
    'Normalizing OpenGL 3.3 config keys...',
    'Verifying final project integrity...',
  ];

  onMount(async () => {
    await performFullScan();
  });

  async function performFullScan() {
    isScanning = true;
    executionStep = 'Analyzing project files...';
    
    const [cfg, diff, logs] = await Promise.all([
      projectStore.inspectConfig(),
      projectStore.inspectDiff(),
      projectStore.getLogs(),
    ]);

    configInspection = cfg;
    diffSummary = diff;
    logContent = logs;
    isScanning = false;
  }

  async function runWithAnimatedPhases(taskFn: () => Promise<void>) {
    isExecuting = true;
    showSuccessBanner = false;

    executionStep = 'Repairing assets...';
    await taskFn();

    executionStep = 'Verifying final project integrity...';
    await performFullScan();
    
    isExecuting = false;
    showSuccessBanner = true;
  }

  // Action 1: Non-destructive Quick Fix (Config Only)
  async function handleAutoRepairConfigOnly() {
    await runWithAnimatedPhases(async () => {
      await projectStore.repairConfig();
    });
  }

  // Action 2: Non-destructive Restore Missing Files Only
  function requestRestoreMissingOnly() {
    confirmTitle = 'Restore Missing Core Files';
    
    let hasMissing = false;
    let missingItems = [];
    if (diffSummary) {
      for (const cat of diffSummary.categories) {
        if (cat.status === 'missing' || cat.status === 'outdated') {
          hasMissing = true;
          if (cat.category === 'screenpack') missingItems.push('Missing data/*.zss scripts (training.zss, logo.zss)');
          if (cat.category === 'sound') missingItems.push('Missing sound/*.mp3 files (Title, Select, Versus)');
          if (cat.category === 'runtime') missingItems.push('Missing external/ & lib/ files');
        }
      }
    }
    
    if (configInspection && !configInspection.isValid) {
      missingItems.push('Config OpenGL 3.3 normalization');
    }
    
    if (missingItems.length === 0) {
      confirmDescription = 'No missing core files detected. Proceeding will force-sync base system scripts and audio.';
      confirmAffectedItems = ['Force re-sync of missing assets only'];
    } else {
      confirmDescription = 'This will ONLY copy missing stock files from the engine. None of your existing custom assets will be modified.';
      confirmAffectedItems = missingItems;
    }
    confirmActionFn = async () => {
      await runWithAnimatedPhases(async () => {
        if (!$projectStore.current) return;
        lastSyncReport = await projectStore.syncAssets({
          projectDir: $projectStore.current.path,
          syncStockChars: false,
          syncStockStages: false,
          syncScreenpack: true,
          syncSound: true,
          syncFonts: false,
          syncRuntime: true,
          resetConfig: false,
        });
        await projectStore.repairConfig();
      });
    };
    showConfirmModal = true;
  }

  // Action 3: 1-Click Comprehensive Repair
  function requestOneClickFixAll() {
    confirmTitle = '1-Click Comprehensive Project Repair';
    
    let items = [];
    if (diffSummary) {
      for (const cat of diffSummary.categories) {
        if (cat.status !== 'clean') {
          if (cat.category === 'stock_chars') items.push('chars/kfm/ & chars/kfm_zss/ (repaired to modern syntax)');
          if (cat.category === 'stock_stages') items.push('stages/ (default arenas and training stages)');
          if (cat.category === 'fonts') items.push('font/ (default engine fonts)');
          if (cat.category === 'screenpack') items.push('data/ (system scripts & screenpack)');
          if (cat.category === 'sound') items.push('sound/ (stock engine music tracks)');
          if (cat.category === 'runtime') items.push('external/ & lib/ (runtime Lua VM)');
        }
      }
    }
    
    if (configInspection && !configInspection.isValid) {
      items.push('save/config.ini (Normalized unsupported parameters)');
    }
    
    if (items.length === 0) {
      confirmDescription = 'Your project is already healthy and synchronized. Proceeding will force-refresh all core engine files just to be absolutely sure. Custom assets will remain untouched.';
      confirmAffectedItems = ['All baseline engine assets (force refresh)'];
    } else {
      confirmDescription = 'This will safely update and repair all detected discrepancies against the engine baseline. Custom fighters and rosters are untouched.';
      items.push('YOUR CUSTOM FIGHTERS & ROSTERS REMAIN 100% UNTOUCHED');
      confirmAffectedItems = items;
    }
    confirmActionFn = async () => {
      await runWithAnimatedPhases(async () => {
        if (!$projectStore.current) return;
        lastSyncReport = await projectStore.syncAssets({
          projectDir: $projectStore.current.path,
          syncStockChars: true,
          syncStockStages: true,
          syncScreenpack: true,
          syncSound: true,
          syncFonts: true,
          syncRuntime: true,
          resetConfig: true,
        });
        await projectStore.repairConfig();
      });
    };
    showConfirmModal = true;
  }

  // Action 4: Granular Custom Sync
  function requestCustomSync() {
    const affected: string[] = [];
    if (syncStockChars) affected.push('chars/kfm & chars/kfm_zss (Stock Fighters)');
    if (syncScreenpack) affected.push('data/ screenpack and battle scripts (training.zss, fight.def)');
    if (syncSound) affected.push('sound/ stock audio tracks');
    if (syncRuntime) affected.push('external/ & lib/ core Lua runtime');
    if (syncStockStages) affected.push('stages/ default arenas');
    if (syncFonts) affected.push('font/ system fonts');
    if (syncResetConfig) affected.push('save/config.ini reset to engine defaults');

    if (affected.length === 0) return;

    confirmTitle = 'Confirm Asset Synchronization';
    confirmDescription =
      'You are about to synchronize selected engine components. Please verify the affected items below:';
    confirmAffectedItems = affected;
    confirmActionFn = async () => {
      await runWithAnimatedPhases(async () => {
        if (!$projectStore.current) return;
        lastSyncReport = await projectStore.syncAssets({
          projectDir: $projectStore.current.path,
          syncStockChars,
          syncStockStages,
          syncScreenpack,
          syncSound,
          syncFonts,
          syncRuntime,
          resetConfig: syncResetConfig,
        });
      });
    };
    showConfirmModal = true;
  }

  // Action 5: Reset Config Defaults
  function requestResetConfig() {
    confirmTitle = 'Reset Configuration to Engine Defaults';
    confirmDescription =
      'This will reset save/config.ini to clean engine defaults. Any customized resolution, volume, or keybindings in config.ini will be reset to default.';
    confirmAffectedItems = ['save/config.ini'];
    confirmActionFn = async () => {
      await runWithAnimatedPhases(async () => {
        await projectStore.resetConfig();
      });
    };
    showConfirmModal = true;
  }

  function handleRepairSingleCategory(cat: any) {
    if (!$projectStore.current) return;

    if (cat.category === 'stock_chars') {
      confirmTitle = 'Repair Default Fighters';
      confirmDescription =
        'Update stock Kung Fu Man characters (chars/kfm & chars/kfm_zss) to modern ZSS syntax from your active engine. Custom fighters are not affected.';
      confirmAffectedItems = ['chars/kfm/', 'chars/kfm_zss/'];
      confirmActionFn = async () => {
        await runWithAnimatedPhases(async () => {
          if (!$projectStore.current) return;
          lastSyncReport = await projectStore.syncAssets({
            projectDir: $projectStore.current.path,
            syncStockChars: true,
          });
        });
      };
      showConfirmModal = true;
    } else if (cat.category === 'screenpack') {
      confirmTitle = 'Repair System Scripts & Screenpack';
      confirmDescription =
        'Synchronize core engine system scripts (data/training.zss, data/fight.def, data/common1.cns.zss). Custom roster (data/select.def) is protected.';
      confirmAffectedItems = ['data/training.zss', 'data/fight.def', 'data/common1.cns.zss', 'data/dgl.zss'];
      confirmActionFn = async () => {
        await runWithAnimatedPhases(async () => {
          if (!$projectStore.current) return;
          lastSyncReport = await projectStore.syncAssets({
            projectDir: $projectStore.current.path,
            syncScreenpack: true,
          });
        });
      };
      showConfirmModal = true;
    } else if (cat.category === 'sound') {
      confirmTitle = 'Restore Engine Audio & BGM';
      confirmDescription =
        'Restore stock background music files (Title.mp3, Select.mp3, Versus.mp3) into the sound/ folder.';
      confirmAffectedItems = ['sound/ (official engine audio tracks)'];
      confirmActionFn = async () => {
        await runWithAnimatedPhases(async () => {
          if (!$projectStore.current) return;
          lastSyncReport = await projectStore.syncAssets({
            projectDir: $projectStore.current.path,
            syncSound: true,
          });
        });
      };
      showConfirmModal = true;
    } else if (cat.category === 'stock_stages') {
      confirmTitle = 'Restore Default Stages';
      confirmDescription = 'Restore official engine demo and training stages into stages/.';
      confirmAffectedItems = ['stages/stage0.def & default arenas'];
      confirmActionFn = async () => {
        await runWithAnimatedPhases(async () => {
          if (!$projectStore.current) return;
          lastSyncReport = await projectStore.syncAssets({
            projectDir: $projectStore.current.path,
            syncStockStages: true,
          });
        });
      };
      showConfirmModal = true;
    } else if (cat.category === 'fonts') {
      confirmTitle = 'Restore Engine Fonts';
      confirmDescription = 'Restore system typography and bitmap fonts in font/.';
      confirmAffectedItems = ['font/'];
      confirmActionFn = async () => {
        await runWithAnimatedPhases(async () => {
          if (!$projectStore.current) return;
          lastSyncReport = await projectStore.syncAssets({
            projectDir: $projectStore.current.path,
            syncFonts: true,
          });
        });
      };
      showConfirmModal = true;
    } else if (cat.category === 'runtime') {
      confirmTitle = 'Restore Engine Core Runtime';
      confirmDescription = 'Restore clean engine Lua VM, shaders, and system libraries.';
      confirmAffectedItems = ['external/script/', 'external/shaders/', 'lib/'];
      confirmActionFn = async () => {
        await runWithAnimatedPhases(async () => {
          if (!$projectStore.current) return;
          lastSyncReport = await projectStore.syncAssets({
            projectDir: $projectStore.current.path,
            syncRuntime: true,
          });
        });
      };
      showConfirmModal = true;
    } else if (cat.category === 'config') {
      handleAutoRepairConfigOnly();
    }
  }

  function handleRelaunch() {
    projectStore.dismissCrash();
    projectStore.launch();
  }

  async function executeConfirmedAction() {
    showConfirmModal = false;
    if (confirmActionFn) {
      const fn = confirmActionFn;
      confirmActionFn = null;
      await fn();
    }
  }

  // Helpers for summary statuses
  $: totalIssues = (diffSummary?.totalDiscrepancies || 0) + (configInspection && !configInspection.isValid ? 1 : 0);
</script>

<div class="p-8 max-w-6xl mx-auto space-y-6">
  <!-- Top Navigation & Title -->
  <div class="p-6 rounded-2xl bg-dark-800 border border-dark-600 shadow-xl relative overflow-hidden flex flex-col md:flex-row md:items-center justify-between gap-6">
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

        <div class="w-10 h-10 rounded-xl bg-amber-500/10 border border-amber-500/30 text-amber-400 flex items-center justify-center flex-shrink-0">
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
            Scan project structure against clean engine baseline, diagnose syntax errors, and repair missing assets safely
          </p>
        </div>
      </div>
    </div>

    <!-- Quick Action Launch Buttons -->
    <div class="flex items-center gap-2.5 flex-wrap">
      <button
        type="button"
        disabled={isScanning || isExecuting}
        class="px-3.5 py-2 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600 text-slate-300 text-xs font-semibold flex items-center gap-1.5 transition"
        on:click={performFullScan}
      >
        <RefreshCw class="w-3.5 h-3.5 {isScanning ? 'animate-spin text-amber-400' : ''}" />
        <span>{isScanning ? 'Scanning...' : 'Re-Scan'}</span>
      </button>

      <button
        type="button"
        disabled={isExecuting || isScanning}
        class="px-4 py-2.5 rounded-xl bg-gradient-to-r from-amber-600 to-orange-600 hover:from-amber-500 hover:to-orange-500 text-white text-xs font-bold shadow-md shadow-orange-950/40 flex items-center gap-2 transition"
        on:click={requestOneClickFixAll}
      >
        <ShieldCheck class="w-4 h-4" />
        <span>1-Click Safe Repair</span>
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

  <!-- Success Notification Banner -->
  {#if showSuccessBanner}
    <div class="p-4 rounded-2xl bg-emerald-950/40 border border-emerald-500/40 flex items-center justify-between gap-4 animate-in fade-in zoom-in-95 duration-200">
      <div class="flex items-center gap-3">
        <CheckCircle2 class="w-5 h-5 text-emerald-400 flex-shrink-0" />
        <div class="text-xs">
          <span class="font-bold text-slate-100">Repair operation finished successfully!</span>
          <p class="text-slate-300 mt-0.5">
            {lastSyncReport ? `${lastSyncReport.repairedCount} files verified & repaired.` : 'Configuration normalized.'} You can test your game now.
          </p>
        </div>
      </div>
      <button
        type="button"
        class="px-4 py-1.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-bold flex items-center gap-1.5 transition"
        on:click={handleRelaunch}
      >
        <Play class="w-3.5 h-3.5 fill-current" />
        <span>Launch Game</span>
      </button>
    </div>
  {/if}

  <!-- Navigation Tabs -->
  <div class="flex items-center gap-3 border-b border-dark-600 pb-3">
    <button
      type="button"
      class="pb-2 text-xs font-bold tracking-wide border-b-2 transition flex items-center gap-2 {
        activeTab === 'analysis'
          ? 'border-amber-500 text-amber-400'
          : 'border-transparent text-slate-400 hover:text-slate-200'
      }"
      on:click={() => (activeTab = 'analysis')}
    >
      <Activity class="w-4 h-4" />
      <span>Scan & Diagnosis</span>
      {#if totalIssues > 0}
        <span class="px-1.5 py-0.2 rounded-full bg-amber-500/20 text-amber-300 font-mono text-[10px]">
          {totalIssues} {totalIssues === 1 ? 'issue' : 'issues'}
        </span>
      {:else}
        <span class="px-1.5 py-0.2 rounded-full bg-emerald-500/20 text-emerald-400 font-mono text-[10px]">
          Healthy
        </span>
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
      <span>Asset Sync Wizard</span>
    </button>

    <button
      type="button"
      class="pb-2 text-xs font-bold tracking-wide border-b-2 transition flex items-center gap-2 {
        activeTab === 'config'
          ? 'border-cyan-500 text-cyan-400'
          : 'border-transparent text-slate-400 hover:text-slate-200'
      }"
      on:click={() => (activeTab = 'config')}
    >
      <Sliders class="w-4 h-4" />
      <span>Config Doctor</span>
      {#if configInspection && !configInspection.isValid}
        <span class="w-2 h-2 rounded-full bg-rose-500 animate-pulse"></span>
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
      <span>Audit & Logs</span>
    </button>
  </div>

  <!-- TAB CONTENT AREA -->
  <div class="relative min-h-[400px]">
    {#if isScanning || isExecuting}
      <div class="absolute inset-0 z-10 flex flex-col items-center justify-center bg-dark-900/60 backdrop-blur-sm rounded-2xl animate-in fade-in duration-200">
        <Loader2 class="w-10 h-10 text-amber-500 animate-spin mb-4 shadow-amber-500/50 drop-shadow-lg" />
        <div class="text-sm font-bold text-slate-100 drop-shadow-md">
          {isExecuting ? 'Repairing Project Assets...' : 'Scanning Project Assets...'}
        </div>
        <div class="text-xs text-amber-300/80 mt-1 drop-shadow font-mono">{executionStep}</div>
      </div>
    {/if}

    <!-- TAB 1: Scan & Diagnosis (Default) -->
    {#if activeTab === 'analysis'}
      <div class="space-y-6 animate-in fade-in duration-150 {(isScanning || isExecuting) ? 'opacity-30 pointer-events-none blur-sm' : ''}">
      <!-- Health Overview Card -->
      <div class="p-6 rounded-2xl bg-dark-800 border {totalIssues === 0 ? 'border-emerald-500/40 bg-emerald-950/10' : 'border-amber-500/40 bg-amber-950/10'} shadow-sm space-y-4">
        <div class="flex items-start justify-between gap-4 flex-wrap">
          <div class="flex items-start gap-3.5">
            {#if totalIssues === 0}
              <CheckCircle2 class="w-6 h-6 text-emerald-400 flex-shrink-0 mt-0.5" />
              <div>
                <h3 class="text-base font-bold text-slate-100">Project is Healthy & Synchronized</h3>
                <p class="text-xs text-slate-400 mt-1">
                  All baseline scripts, sound files, and configuration parameters match engine specifications.
                </p>
              </div>
            {:else}
              <AlertTriangle class="w-6 h-6 text-amber-400 flex-shrink-0 mt-0.5" />
              <div>
                <h3 class="text-base font-bold text-slate-100">{totalIssues} Discrepancies or Incompatibilities Found</h3>
                <p class="text-xs text-slate-400 mt-1">
                  Detected missing sound assets, legacy syntax, or config issues that could cause engine crashes.
                </p>
              </div>
            {/if}
          </div>

          <!-- Quick Solution Buttons -->
          <div class="flex items-center gap-2">
            {#if totalIssues > 0}
              <button
                type="button"
                disabled={isExecuting}
                class="px-4 py-2 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600 text-slate-200 text-xs font-semibold flex items-center gap-1.5 transition"
                on:click={requestRestoreMissingOnly}
              >
                <ShieldCheck class="w-3.5 h-3.5 text-emerald-400" />
                <span>Restore Missing Files Only</span>
              </button>

              <button
                type="button"
                disabled={isExecuting}
                class="px-4 py-2 rounded-xl bg-amber-600 hover:bg-amber-500 text-white text-xs font-bold shadow-md flex items-center gap-1.5 transition"
                on:click={requestOneClickFixAll}
              >
                <Wrench class="w-3.5 h-3.5" />
                <span>Fix All Issues</span>
              </button>
            {/if}
          </div>
        </div>

        <!-- Discrepancy Breakdown Grid -->
        {#if diffSummary}
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3 pt-2">
            {#each diffSummary.categories as cat}
              <div class="p-4 rounded-xl bg-dark-900 border border-dark-600 flex items-start justify-between gap-3">
                <div class="space-y-1 flex-1">
                  <div class="flex items-center gap-2">
                    <span class="text-xs font-bold text-slate-200">{cat.title}</span>
                    {#if cat.status === 'missing'}
                      <span class="px-2 py-0.5 rounded text-[10px] font-bold uppercase font-mono bg-rose-500/20 text-rose-300 border border-rose-500/30">
                        Missing Files
                      </span>
                    {:else if cat.status === 'outdated'}
                      <span class="px-2 py-0.5 rounded text-[10px] font-bold uppercase font-mono bg-amber-500/20 text-amber-300 border border-amber-500/30">
                        Outdated / Legacy
                      </span>
                    {:else}
                      <span class="px-2 py-0.5 rounded text-[10px] font-bold uppercase font-mono bg-emerald-500/20 text-emerald-400 border border-emerald-500/30">
                        Clean
                      </span>
                    {/if}
                  </div>
                  <p class="text-[11px] text-slate-400 leading-relaxed">{cat.description}</p>
                  
                  <!-- Preview of affected files -->
                  {#if cat.files && cat.files.length > 0}
                    <div class="pt-1 flex items-center gap-1.5 flex-wrap">
                      {#each cat.files.slice(0, 3) as f}
                        <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-dark-800 border border-dark-600 text-slate-400">
                          {f}
                        </span>
                      {/each}
                      {#if cat.files.length > 3}
                        <span class="text-[10px] text-slate-500">+{cat.files.length - 3} more</span>
                      {/if}
                    </div>
                  {/if}
                </div>

                <!-- Individual Repair Button for this Category -->
                {#if cat.status !== 'clean'}
                  <button
                    type="button"
                    disabled={isExecuting}
                    class="px-3 py-1.5 rounded-xl bg-amber-500/10 hover:bg-amber-500/20 text-amber-300 border border-amber-500/30 text-xs font-bold flex items-center gap-1.5 transition flex-shrink-0 self-center shadow-sm"
                    title="Repair {cat.title}"
                    on:click={() => handleRepairSingleCategory(cat)}
                  >
                    <Wrench class="w-3.5 h-3.5 text-amber-400" />
                    <span>Fix</span>
                  </button>
                {/if}
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>

  <!-- TAB 2: Asset Sync Wizard -->
  {:else if activeTab === 'sync'}
    <div class="space-y-6 animate-in fade-in duration-150">
      <div class="p-5 rounded-2xl bg-indigo-950/30 border border-indigo-500/30 flex items-start gap-3">
        <Sparkles class="w-5 h-5 text-indigo-400 flex-shrink-0 mt-0.5" />
        <div class="space-y-1 text-xs">
          <h3 class="font-bold text-slate-100">Selective Engine Asset Synchronizer</h3>
          <p class="text-slate-300 leading-relaxed">
            Selectively replace individual engine components from your clean engine version. All actions show an explicit confirmation preview before anything is modified.
          </p>
          <p class="text-emerald-400 font-semibold pt-1">
            ✓ Your custom fighters, stages, and roster (<code class="text-emerald-300">data/select.def</code>) are strictly protected.
          </p>
        </div>
      </div>

      <div class="p-6 rounded-2xl bg-dark-800 border border-dark-600 space-y-4 shadow-sm">
        <div class="flex items-center justify-between">
          <span class="text-xs font-bold uppercase tracking-wider text-slate-300">Select Components to Synchronize</span>
          <button
            type="button"
            class="px-5 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold shadow-md shadow-indigo-950/50 flex items-center gap-2 transition"
            disabled={isExecuting}
            on:click={requestCustomSync}
          >
            <Check class="w-3.5 h-3.5" />
            <span>Sync Selected Components...</span>
          </button>
        </div>

        <div class="space-y-3">
          <!-- 1. Screenpack & System Scripts -->
          <label class="p-4 rounded-xl border border-dark-600 flex items-start gap-3.5 cursor-pointer transition {syncScreenpack ? 'bg-dark-900 border-indigo-500/60 ring-1 ring-indigo-500/20' : 'bg-dark-900'}">
            <input type="checkbox" bind:checked={syncScreenpack} class="rounded border-dark-600 bg-dark-800 text-indigo-600 mt-1" />
            <div class="space-y-1 flex-1">
              <div class="flex items-center justify-between">
                <span class="text-xs font-bold text-slate-100">System Scripts & Screenpack (training.zss, fight.def, common1.cns.zss)</span>
                <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-300 border border-emerald-500/30">Recommended</span>
              </div>
              <p class="text-[11px] text-slate-400">Replaces core system scripts and training mode syntax while preserving your roster (select.def).</p>
            </div>
          </label>

          <!-- 2. Engine Sound & BGM -->
          <label class="p-4 rounded-xl border border-dark-600 flex items-start gap-3.5 cursor-pointer transition {syncSound ? 'bg-dark-900 border-indigo-500/60 ring-1 ring-indigo-500/20' : 'bg-dark-900'}">
            <input type="checkbox" bind:checked={syncSound} class="rounded border-dark-600 bg-dark-800 text-indigo-600 mt-1" />
            <div class="space-y-1 flex-1">
              <div class="flex items-center justify-between">
                <span class="text-xs font-bold text-slate-100">Engine Audio & BGM (sound/ Title, Select, Versus)</span>
                <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-purple-500/10 text-purple-300 border border-purple-500/30">Fixes BGM crash</span>
              </div>
              <p class="text-[11px] text-slate-400">Restores default engine background music files in the sound/ folder required by system.def.</p>
            </div>
          </label>

          <!-- 3. Stock Characters -->
          <label class="p-4 rounded-xl border border-dark-600 flex items-start gap-3.5 cursor-pointer transition {syncStockChars ? 'bg-dark-900 border-indigo-500/60 ring-1 ring-indigo-500/20' : 'bg-dark-900'}">
            <input type="checkbox" bind:checked={syncStockChars} class="rounded border-dark-600 bg-dark-800 text-indigo-600 mt-1" />
            <div class="space-y-1 flex-1">
              <div class="flex items-center justify-between">
                <span class="text-xs font-bold text-slate-100">Default Characters (KFM & KFM ZSS)</span>
                <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-purple-500/10 text-purple-300 border border-purple-500/30">Fixes kfm.zss crash</span>
              </div>
              <p class="text-[11px] text-slate-400">Replaces outdated stock Kung Fu Man characters with modern ZSS syntax versions from the engine.</p>
            </div>
          </label>

          <!-- 4. Engine Core Runtime -->
          <label class="p-4 rounded-xl border border-dark-600 flex items-start gap-3.5 cursor-pointer transition {syncRuntime ? 'bg-dark-900 border-indigo-500/60 ring-1 ring-indigo-500/20' : 'bg-dark-900'}">
            <input type="checkbox" bind:checked={syncRuntime} class="rounded border-dark-600 bg-dark-800 text-indigo-600 mt-1" />
            <div class="space-y-1 flex-1">
              <span class="text-xs font-bold text-slate-100">Engine Lua VM & Shaders (external/ & lib/)</span>
              <p class="text-[11px] text-slate-400">Restores clean Lua script execution environment and modern display shaders.</p>
            </div>
          </label>

          <!-- 5. Default Stages -->
          <label class="p-4 rounded-xl border border-dark-600 flex items-start gap-3.5 cursor-pointer transition {syncStockStages ? 'bg-dark-900 border-indigo-500/60 ring-1 ring-indigo-500/20' : 'bg-dark-900'}">
            <input type="checkbox" bind:checked={syncStockStages} class="rounded border-dark-600 bg-dark-800 text-indigo-600 mt-1" />
            <div class="space-y-1 flex-1">
              <span class="text-xs font-bold text-slate-100">Default Stages (stages/stage0.def)</span>
              <p class="text-[11px] text-slate-400">Ensures standard training and demo stages exist.</p>
            </div>
          </label>

          <!-- 6. System Fonts -->
          <label class="p-4 rounded-xl border border-dark-600 flex items-start gap-3.5 cursor-pointer transition {syncFonts ? 'bg-dark-900 border-indigo-500/60 ring-1 ring-indigo-500/20' : 'bg-dark-900'}">
            <input type="checkbox" bind:checked={syncFonts} class="rounded border-dark-600 bg-dark-800 text-indigo-600 mt-1" />
            <div class="space-y-1 flex-1">
              <span class="text-xs font-bold text-slate-100">Engine Fonts (font/)</span>
              <p class="text-[11px] text-slate-400">Restores engine UI and bitmap font assets.</p>
            </div>
          </label>

          <!-- 7. Reset Config -->
          <label class="p-4 rounded-xl border border-dark-600 flex items-start gap-3.5 cursor-pointer transition {syncResetConfig ? 'bg-dark-900 border-indigo-500/60 ring-1 ring-indigo-500/20' : 'bg-dark-900'}">
            <input type="checkbox" bind:checked={syncResetConfig} class="rounded border-dark-600 bg-dark-800 text-indigo-600 mt-1" />
            <div class="space-y-1 flex-1">
              <span class="text-xs font-bold text-slate-100">Reset save/config.ini to Clean Defaults (OpenGL 3.3)</span>
              <p class="text-[11px] text-slate-400">Eliminates legacy render mode syntax and corrupted display resolutions.</p>
            </div>
          </label>
        </div>
      </div>
    </div>

  <!-- TAB 3: Config Doctor -->
  {:else if activeTab === 'config'}
    <div class="space-y-6 animate-in fade-in duration-150">
      <div class="p-5 rounded-2xl bg-dark-800 border {configInspection?.isValid ? 'border-emerald-500/40 bg-emerald-950/10' : 'border-cyan-500/40 bg-cyan-950/10'} shadow-sm space-y-4">
        <div class="flex items-start justify-between gap-4">
          <div class="flex items-start gap-3">
            {#if configInspection?.isValid}
              <CheckCircle2 class="w-5 h-5 text-emerald-400 flex-shrink-0 mt-0.5" />
              <div>
                <h3 class="text-sm font-bold text-slate-100">Configuration is Valid & Compatible</h3>
                <p class="text-xs text-slate-400 mt-0.5">Your save/config.ini parameters match modern Ikemen GO engine requirements (RenderMode = OpenGL 3.3).</p>
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
              disabled={isExecuting}
              class="px-4 py-2 rounded-xl bg-cyan-600 hover:bg-cyan-500 disabled:opacity-50 text-white text-xs font-bold shadow-md flex items-center gap-1.5 transition"
              on:click={handleAutoRepairConfigOnly}
            >
              <Wrench class="w-3.5 h-3.5" />
              <span>Auto-Fix Config (OpenGL 3.3)</span>
            </button>

            <button
              type="button"
              disabled={isExecuting}
              class="px-4 py-2 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600 text-slate-200 text-xs font-semibold flex items-center gap-1.5 transition"
              on:click={requestResetConfig}
            >
              <RotateCcw class="w-3.5 h-3.5 text-rose-400" />
              <span>Reset to Defaults</span>
            </button>
          </div>
        </div>

        {#if configInspection && configInspection.issues.length > 0}
          <div class="pt-3 border-t border-dark-600 space-y-2">
            <div class="text-[11px] font-bold uppercase tracking-wider text-slate-400">Identified Parameter Issues:</div>
            <div class="space-y-1.5">
              {#each configInspection.issues as issue}
                <div class="p-3 rounded-xl bg-dark-900 border border-dark-600 flex items-start justify-between gap-3 text-xs">
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

  <!-- TAB 4: Logs & Diagnostics -->
  {:else if activeTab === 'logs'}
    <div class="space-y-6 animate-in fade-in duration-150">
      <div class="p-6 rounded-2xl bg-dark-800 border border-dark-600 space-y-4 shadow-sm">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <FileText class="w-4 h-4 text-purple-400" />
            <span class="text-xs font-bold uppercase tracking-wider text-slate-300">Engine & Diagnostic Output</span>
          </div>

          <div class="flex items-center gap-2">
            <button
              type="button"
              class="px-3 py-1.5 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600 text-slate-200 text-xs font-semibold flex items-center gap-1.5 transition"
              on:click={async () => {
                logContent = await projectStore.getLogs();
              }}
            >
              <RefreshCw class="w-3.5 h-3.5" />
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

        <div class="p-4 rounded-xl bg-dark-900 border border-dark-600 font-mono text-xs text-slate-300 whitespace-pre-wrap max-h-96 overflow-y-auto leading-relaxed shadow-inner">
          {logContent || 'No logs recorded yet.'}
        </div>
      </div>
    </div>
  {/if}
  </div>
</div>

<!-- Explicit Confirmation Modal for Sensitive Actions -->
{#if showConfirmModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-sm p-4 animate-in fade-in duration-150">
    <div class="bg-dark-800 border border-dark-600 rounded-2xl w-full max-w-lg shadow-2xl overflow-hidden animate-in zoom-in-95 duration-150">
      <div class="p-6 space-y-4">
        <div class="flex items-start justify-between gap-4">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-amber-500/10 border border-amber-500/30 text-amber-400 flex items-center justify-center flex-shrink-0">
              <ShieldAlert class="w-5 h-5" />
            </div>
            <div>
              <h3 class="text-base font-bold text-slate-100">{confirmTitle}</h3>
              <p class="text-xs text-slate-400 mt-0.5">Please review before proceeding</p>
            </div>
          </div>
          <button
            type="button"
            class="text-slate-400 hover:text-white p-1"
            on:click={() => (showConfirmModal = false)}
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <p class="text-xs text-slate-300 leading-relaxed bg-dark-900/60 p-3 rounded-xl border border-dark-600">
          {confirmDescription}
        </p>

        {#if confirmAffectedItems.length > 0}
          <div class="space-y-2">
            <div class="text-[11px] font-bold uppercase tracking-wider text-slate-400">Items to be updated:</div>
            <div class="max-h-48 overflow-y-auto space-y-1 p-2 rounded-xl bg-dark-900 border border-dark-600">
              {#each confirmAffectedItems as item}
                <div class="text-xs flex items-center gap-2 text-slate-300 font-mono">
                  <span class="w-1.5 h-1.5 rounded-full bg-amber-400 flex-shrink-0"></span>
                  <span>{item}</span>
                </div>
              {/each}
            </div>
          </div>
        {/if}

        <div class="flex items-center justify-end gap-3 pt-2">
          <button
            type="button"
            class="px-4 py-2 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600 text-slate-300 text-xs font-semibold transition"
            on:click={() => (showConfirmModal = false)}
          >
            Cancel
          </button>

          <button
            type="button"
            class="px-5 py-2 rounded-xl bg-gradient-to-r from-amber-600 to-orange-600 hover:from-amber-500 hover:to-orange-500 text-white text-xs font-bold shadow-md transition"
            on:click={executeConfirmedAction}
          >
            Confirm & Proceed
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}

