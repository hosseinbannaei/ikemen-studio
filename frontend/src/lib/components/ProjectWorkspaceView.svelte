<script lang="ts">
  import { onMount } from 'svelte';
  import { projectStore } from '../stores/projectStore';
  import { engineStore } from '../stores/engineStore';
  import { toastStore } from '../stores/toastStore';
  import type { VerificationReport } from '../types';
  import VerifyOptionsModal from './VerifyOptionsModal.svelte';
  import VerifyReportModal from './VerifyReportModal.svelte';
  import CustomLaunchModal from './CustomLaunchModal.svelte';
  import GameConfigModal from './GameConfigModal.svelte';
  import ConfirmModal from './ConfirmModal.svelte';
  import AddFromVaultModal from './AddFromVaultModal.svelte';
  import FileEditorModal from './FileEditorModal.svelte';

  import {
    Play,
    Square,
    Folder,
    ExternalLink,
    Users,
    Mountain,
    FileCode,
    Type,
    Music,
    HardDrive,
    Calendar,
    User,
    XCircle,
    Loader2,
    Activity,
    Wrench,
    FileText,
    ChevronDown,
    Sliders,
    Terminal,
    RotateCcw,
    Sparkles,
    Package,
    Plus,
    Grid,
    Copy,
    Check,
    ArrowRight,
    ShieldCheck,
    RefreshCw,
    Cpu,
    Layers,
    FolderOpen,
    Palette,
    Swords,
    Film,
    Search,
  } from 'lucide-svelte';

  export let onBackToProjects: () => void;
  export let onOpenRepairHub: () => void;
  export let onOpenRosterEditor: () => void;
  export let onOpenStages: () => void;
  export let onOpenMotifs: () => void;
  export let onOpenLifebars: () => void;
  export let onOpenSound: () => void;
  export let onOpenFonts: () => void;
  export let onOpenStoryboards: () => void;

  let showVerifyOptionsModal = false;
  let showVerifyReportModal = false;
  let showCustomLaunchModal = false;
  let showGameConfigModal = false;
  let showEngineConfirmModal = false;
  let showRollbackModal = false;
  let showAddFromVaultModal = false;
  let selectedEditFilePath: string | null = null;
  let vaultTargetCategory: 'fighters' | 'stages' | 'motifs' | 'lifebars' | 'sounds' | 'fonts' | 'storyboards' = 'fighters';

  let selectedBackupId = '';
  let pendingEngineVersion = '';
  let selectedEngineVersion = '';

  // Stats & Log Preview
  let fighterCount = 0;
  let stageCount = 0;
  let recentLogs = '';
  let isLoadingStats = false;
  let isRefreshingLogs = false;
  let copiedPath = false;
  let activeLaunchMode: 'normal' | 'training' | 'debug' = 'normal';

  $: if ($projectStore.current?.engine?.version && !showEngineConfirmModal) {
    selectedEngineVersion = $projectStore.current.engine.version;
  }

  let verificationReport: VerificationReport | null = null;
  let isVerifying = false;

  $: folderShortcuts = [
    {
      label: 'Characters',
      subpath: 'chars',
      icon: Users,
      desc: 'Fighter packages, sprites and .def files',
      color: 'from-blue-500/20 to-cyan-500/20 text-cyan-400 border-cyan-500/30',
      category: 'fighters' as const,
      canAddVault: true,
      onManage: onOpenRosterEditor,
      manageLabel: 'Roster',
    },
    {
      label: 'Stages',
      subpath: 'stages',
      icon: Mountain,
      desc: 'Background arenas, zoom defs & BGM',
      color: 'from-emerald-500/20 to-teal-500/20 text-emerald-400 border-emerald-500/30',
      category: 'stages' as const,
      canAddVault: true,
      onManage: onOpenStages,
      manageLabel: 'Manage',
    },
    {
      label: 'System & Motifs',
      subpath: 'data',
      icon: Palette,
      desc: 'Screenpacks, motifs, lifebars & cutscenes',
      color: 'from-indigo-500/20 to-purple-500/20 text-indigo-400 border-indigo-500/30',
      category: 'motifs' as const,
      canAddVault: true,
      onManage: onOpenMotifs,
      manageLabel: 'Motifs',
    },
    {
      label: 'Fonts',
      subpath: 'font',
      icon: Type,
      desc: 'Bitmap (.fnt) and TrueType (.ttf) fonts',
      color: 'from-purple-500/20 to-pink-500/20 text-purple-400 border-purple-500/30',
      category: 'fonts' as const,
      canAddVault: true,
      onManage: onOpenFonts,
      manageLabel: 'Manage',
    },
    {
      label: 'Sound & Music',
      subpath: 'sound',
      icon: Music,
      desc: 'BGM soundtracks, hits, announcer audio',
      color: 'from-rose-500/20 to-red-500/20 text-rose-400 border-rose-500/30',
      category: 'sounds' as const,
      canAddVault: true,
      onManage: onOpenSound,
      manageLabel: 'Manage',
    },
  ];

  onMount(async () => {
    await loadProjectStats();
    await fetchRecentLogs();
  });

  async function loadProjectStats() {
    if (!$projectStore.current?.path) return;
    isLoadingStats = true;
    try {
      const data = await projectStore.getFightersAndStages($projectStore.current.path);
      fighterCount = data?.characters?.length || 0;
      stageCount = data?.stages?.length || 0;
    } catch {
      fighterCount = 0;
      stageCount = 0;
    } finally {
      isLoadingStats = false;
    }
  }

  async function fetchRecentLogs() {
    if (!$projectStore.current?.path) return;
    isRefreshingLogs = true;
    try {
      recentLogs = await projectStore.getLogs($projectStore.current.path);
    } catch {
      recentLogs = '';
    } finally {
      isRefreshingLogs = false;
    }
  }

  function handleOpenVaultFor(cat: 'fighters' | 'stages' | 'motifs' | 'lifebars' | 'sounds' | 'fonts' | 'storyboards') {
    vaultTargetCategory = cat;
    showAddFromVaultModal = true;
  }

  function handleVerifyClick() {
    onOpenRepairHub();
  }

  function handleReportReceived(report: VerificationReport) {
    verificationReport = report;
    showVerifyReportModal = true;
  }

  async function handleLaunchMode(mode: 'normal' | 'training' | 'debug') {
    activeLaunchMode = mode;
    if (mode === 'normal') {
      await projectStore.launch();
    } else if (mode === 'training') {
      const data = await projectStore.getFightersAndStages();
      const char = data.characters.length > 0 ? data.characters[0] : 'kfm';
      const stage = data.stages.length > 0 ? data.stages[0] : '';
      const args = ['-p1', char, '-p2', char, '-p2.ai', '0', '-time', '-1'];
      if (stage) {
        args.push('-s', stage);
      }
      await projectStore.launchWithOptions(args);
    } else if (mode === 'debug') {
      await projectStore.launchWithOptions(['-debug', '-maxpowermode']);
    }
  }

  function handleEngineSelect(e: Event) {
    const val = (e.target as HTMLSelectElement).value;
    if (val && val !== $projectStore.current?.engine.version) {
      pendingEngineVersion = val;
      showEngineConfirmModal = true;
    }
  }

  function cancelEngineSwitch() {
    showEngineConfirmModal = false;
    pendingEngineVersion = '';
    if ($projectStore.current?.engine?.version) {
      selectedEngineVersion = $projectStore.current.engine.version;
    }
  }

  async function confirmEngineSwitch() {
    showEngineConfirmModal = false;
    if (pendingEngineVersion) {
      await projectStore.switchEngine(pendingEngineVersion);
      await loadProjectStats();
    }
  }

  async function confirmRollback() {
    showRollbackModal = false;
    if (selectedBackupId) {
      await projectStore.rollbackEngine(selectedBackupId);
      await loadProjectStats();
    }
  }

  async function copyPathToClipboard() {
    if (!$projectStore.current?.path) return;
    try {
      await navigator.clipboard.writeText($projectStore.current.path);
      copiedPath = true;
      toastStore.info('Path Copied', 'Project directory copied to clipboard');
      setTimeout(() => (copiedPath = false), 2000);
    } catch {
      // Fallback
    }
  }

  function formatDate(d: string): string {
    if (!d) return '-';
    try {
      return new Date(d).toLocaleDateString(undefined, {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
      });
    } catch {
      return d;
    }
  }

  $: isBusy = $projectStore.gameState !== 'idle';
</script>

{#if $projectStore.current}
  <div class="max-w-7xl mx-auto p-6 md:p-8 space-y-6 pb-12">
    <!-- Header: Project Overview & Core Utilities -->
    <header class="p-6 rounded-2xl bg-dark-800 border border-dark-600/70 shadow-lg relative overflow-hidden transition-all duration-300">
      <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-6 relative z-10">
        <!-- Left: Project Identity & Metadata -->
        <div class="space-y-3 min-w-0">
          <div class="flex items-center gap-3 flex-wrap">
            <h1 class="text-2xl md:text-3xl font-black text-slate-100 tracking-tight truncate">
              {$projectStore.current.name}
            </h1>

            <!-- Engine Version Selector Pill -->
            <div class="flex items-center gap-1.5 bg-purple-500/10 border border-purple-500/30 rounded-full px-2.5 py-1">
              <Cpu class="w-3.5 h-3.5 text-purple-400" />
              <select
                bind:value={selectedEngineVersion}
                on:change={handleEngineSelect}
                disabled={isBusy}
                class="text-xs font-mono bg-transparent text-purple-200 font-semibold focus:outline-none cursor-pointer disabled:opacity-50"
                title="Change project engine version (with automated safety backup)"
              >
                {#each $engineStore.installed as engine}
                  <option value={engine.version} class="bg-dark-800 text-slate-100">
                    {engine.version} ({engine.channel})
                  </option>
                {/each}
              </select>

              {#if $projectStore.backups && $projectStore.backups.length > 0}
                <button
                  type="button"
                  class="ml-1 p-0.5 rounded-full hover:bg-purple-500/20 text-purple-400 transition"
                  title="Rollback to previous engine backup ({$projectStore.backups[0].version})"
                  on:click={() => {
                    selectedBackupId = $projectStore.backups[0].id;
                    showRollbackModal = true;
                  }}
                >
                  <RotateCcw class="w-3.5 h-3.5" />
                </button>
              {/if}
            </div>

            <!-- Status Indicator -->
            {#if $projectStore.gameState === 'running'}
              <span class="flex items-center gap-1.5 text-xs px-3 py-1 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 font-semibold">
                <span class="w-2 h-2 rounded-full bg-emerald-400 animate-ping"></span>
                Engine Session Active
              </span>
            {:else if $projectStore.gameState === 'starting'}
              <span class="flex items-center gap-1.5 text-xs px-3 py-1 rounded-full bg-amber-500/10 text-amber-400 border border-amber-500/30 font-semibold">
                <Loader2 class="w-3 h-3 animate-spin" />
                Spawning Process...
              </span>
            {/if}
          </div>

          <!-- Metadata row -->
          <div class="flex items-center gap-4 text-xs text-slate-400 flex-wrap">
            {#if $projectStore.current.author}
              <div class="flex items-center gap-1.5">
                <User class="w-3.5 h-3.5 text-slate-500" />
                <span>Author: <strong class="text-slate-200">{$projectStore.current.author}</strong></span>
              </div>
            {/if}
            <div class="flex items-center gap-1.5">
              <Calendar class="w-3.5 h-3.5 text-slate-500" />
              <span>Created: {formatDate($projectStore.current.created_at)}</span>
            </div>
            <button
              type="button"
              class="flex items-center gap-1.5 text-slate-400 hover:text-slate-200 font-mono text-xs transition group"
              on:click={copyPathToClipboard}
              title="Click to copy path"
            >
              <HardDrive class="w-3.5 h-3.5 text-slate-500 group-hover:text-indigo-400 transition" />
              <span class="truncate max-w-xs md:max-w-md">{$projectStore.current.path}</span>
              {#if copiedPath}
                <Check class="w-3 h-3 text-emerald-400" />
              {:else}
                <Copy class="w-3 h-3 text-slate-600 group-hover:text-slate-400" />
              {/if}
            </button>
          </div>
        </div>

        <!-- Right: Utility Quick Actions -->
        <div class="flex items-center gap-2 flex-wrap self-start lg:self-center">
          <button
            type="button"
            class="px-3 py-2 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600/70 text-slate-200 text-xs font-semibold flex items-center gap-2 transition shadow-sm"
            on:click={() => projectStore.openFolder()}
            title="Open project folder in File Explorer"
          >
            <FolderOpen class="w-4 h-4 text-indigo-400" />
            <span>Explorer</span>
          </button>

          <button
            type="button"
            class="px-3 py-2 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600/70 text-slate-200 text-xs font-semibold flex items-center gap-2 transition shadow-sm"
            on:click={() => projectStore.openLogs()}
            title="Open project logs directory"
          >
            <FileText class="w-4 h-4 text-amber-400" />
            <span>Logs</span>
          </button>

          <button
            type="button"
            disabled={isBusy}
            class="px-3 py-2 rounded-xl bg-dark-700 hover:bg-rose-500/10 hover:text-rose-400 hover:border-rose-500/30 border border-dark-600/70 text-slate-400 text-xs font-semibold flex items-center gap-1.5 transition shadow-sm disabled:opacity-40"
            on:click={() => {
              projectStore.close();
              onBackToProjects();
            }}
            title="Close this project and return to hub"
          >
            <XCircle class="w-4 h-4" />
            <span>Close</span>
          </button>
        </div>
      </div>
    </header>

    <!-- Hero Launch & Live KPI Control Deck -->
    <div class="p-6 rounded-2xl bg-gradient-to-br from-dark-800 to-dark-850 border border-dark-600/70 shadow-lg space-y-6">
      <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-6">
        <!-- Launch Controls -->
        <div class="space-y-3">
          <div class="text-xs font-bold uppercase tracking-wider text-slate-400">Launch & Runtime Mode</div>
          <div class="flex flex-wrap items-center gap-2">
            {#if $projectStore.gameState === 'starting'}
              <button
                type="button"
                disabled
                class="px-6 py-3 rounded-xl font-bold text-sm bg-dark-700 border border-dark-600 text-slate-400 cursor-not-allowed flex items-center gap-2.5 transition shadow"
              >
                <Loader2 class="w-4 h-4 animate-spin text-amber-400" />
                <span>Starting Ikemen GO...</span>
              </button>
            {:else if $projectStore.gameState === 'stopping'}
              <button
                type="button"
                disabled
                class="px-6 py-3 rounded-xl font-bold text-sm bg-dark-700 border border-dark-600 text-slate-400 cursor-not-allowed flex items-center gap-2.5 transition shadow"
              >
                <Loader2 class="w-4 h-4 animate-spin text-rose-400" />
                <span>Stopping Engine...</span>
              </button>
            {:else if $projectStore.gameState === 'running'}
              <button
                type="button"
                disabled={!$projectStore.canStop}
                class="px-6 py-3 rounded-xl font-bold text-sm shadow-lg flex items-center gap-2.5 transition {
                  $projectStore.canStop
                    ? 'bg-rose-600 hover:bg-rose-500 text-white shadow-rose-950/50 cursor-pointer'
                    : 'bg-rose-900/60 text-rose-300 border border-rose-800/50 cursor-wait'
                }"
                on:click={() => projectStore.stop()}
              >
                <Square class="w-4 h-4 fill-current" />
                <span>{$projectStore.canStop ? 'Stop Game Session' : 'Game Running...'}</span>
              </button>
            {:else}
              <!-- Primary Launch Button -->
              <button
                type="button"
                class="px-6 py-3 rounded-xl font-bold text-sm shadow-xl bg-emerald-600 hover:bg-emerald-500 text-white shadow-emerald-950/60 flex items-center gap-2.5 transition hover:scale-[1.02]"
                on:click={() => handleLaunchMode('normal')}
              >
                <Play class="w-4 h-4 fill-current" />
                <span>Launch Arcade Game</span>
              </button>

              <!-- Quick Launch Presets -->
              <div class="flex items-center gap-1.5 bg-dark-900/80 p-1 rounded-xl border border-dark-600/60">
                <button
                  type="button"
                  class="px-3 py-2 rounded-lg text-xs font-semibold text-slate-300 hover:text-white hover:bg-dark-750 transition flex items-center gap-1.5"
                  on:click={() => handleLaunchMode('training')}
                  title="Direct 1v1 Sparring / Training Mode (-p1 ... -p2 ... -time -1)"
                >
                  <Sparkles class="w-3.5 h-3.5 text-indigo-400" />
                  <span>Sparring</span>
                </button>

                <button
                  type="button"
                  class="px-3 py-2 rounded-lg text-xs font-semibold text-slate-300 hover:text-white hover:bg-dark-750 transition flex items-center gap-1.5"
                  on:click={() => handleLaunchMode('debug')}
                  title="Developer Debug Mode (-debug -maxpowermode)"
                >
                  <Terminal class="w-3.5 h-3.5 text-amber-400" />
                  <span>Debug</span>
                </button>

                <button
                  type="button"
                  class="px-3 py-2 rounded-lg text-xs font-semibold text-purple-300 hover:text-purple-200 hover:bg-purple-500/10 transition flex items-center gap-1.5"
                  on:click={() => (showCustomLaunchModal = true)}
                  title="Custom Launch Flags and Argument History"
                >
                  <Sliders class="w-3.5 h-3.5 text-purple-400" />
                  <span>Custom Flags...</span>
                </button>
              </div>
            {/if}
          </div>
        </div>

        <!-- Project KPI Cards -->
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 lg:w-auto">
          <!-- Fighters Stat -->
          <div class="p-3.5 rounded-xl bg-dark-900/70 border border-dark-600/50 flex flex-col justify-between min-w-[110px]">
            <div class="flex items-center justify-between text-slate-400 mb-1">
              <span class="text-[11px] font-semibold">Fighters</span>
              <Users class="w-3.5 h-3.5 text-cyan-400" />
            </div>
            <div class="text-xl font-bold text-slate-100">
              {isLoadingStats ? '...' : fighterCount}
            </div>
            <div class="text-[10px] text-slate-500 font-mono mt-0.5">chars/</div>
          </div>

          <!-- Stages Stat -->
          <div class="p-3.5 rounded-xl bg-dark-900/70 border border-dark-600/50 flex flex-col justify-between min-w-[110px]">
            <div class="flex items-center justify-between text-slate-400 mb-1">
              <span class="text-[11px] font-semibold">Stages</span>
              <Mountain class="w-3.5 h-3.5 text-emerald-400" />
            </div>
            <div class="text-xl font-bold text-slate-100">
              {isLoadingStats ? '...' : stageCount}
            </div>
            <div class="text-[10px] text-slate-500 font-mono mt-0.5">stages/</div>
          </div>

          <!-- Engine Version -->
          <div class="p-3.5 rounded-xl bg-dark-900/70 border border-dark-600/50 flex flex-col justify-between min-w-[110px]">
            <div class="flex items-center justify-between text-slate-400 mb-1">
              <span class="text-[11px] font-semibold">Engine</span>
              <Cpu class="w-3.5 h-3.5 text-purple-400" />
            </div>
            <div class="text-xs font-bold text-purple-300 truncate" title={$projectStore.current.engine.version}>
              {$projectStore.current.engine.version}
            </div>
            <div class="text-[10px] text-slate-500 font-mono mt-0.5">{$projectStore.current.engine.channel}</div>
          </div>

          <!-- Backups Stat -->
          <div class="p-3.5 rounded-xl bg-dark-900/70 border border-dark-600/50 flex flex-col justify-between min-w-[110px]">
            <div class="flex items-center justify-between text-slate-400 mb-1">
              <span class="text-[11px] font-semibold">Backups</span>
              <RotateCcw class="w-3.5 h-3.5 text-amber-400" />
            </div>
            <div class="text-xl font-bold text-slate-100">
              {$projectStore.backups ? $projectStore.backups.length : 0}
            </div>
            <div class="text-[10px] text-slate-500 font-mono mt-0.5">snapshots</div>
          </div>
        </div>
      </div>

      <!-- Active Running Alert Notice -->
      {#if $projectStore.gameState === 'running'}
        <div class="p-4 rounded-xl bg-emerald-950/50 border border-emerald-500/40 flex items-center justify-between gap-4">
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 rounded-lg bg-emerald-500/20 text-emerald-400 flex items-center justify-center flex-shrink-0">
              <Activity class="w-4 h-4 animate-pulse" />
            </div>
            <div>
              <div class="text-xs font-bold text-emerald-200">Ikemen GO Game Process is Active</div>
              <div class="text-[11px] text-emerald-400/80">Play and test in the native engine window. Click Stop Game when done.</div>
            </div>
          </div>
          {#if $projectStore.canStop}
            <button
              type="button"
              class="px-3 py-1.5 bg-rose-600 hover:bg-rose-500 text-white text-xs font-semibold rounded-lg shadow transition flex items-center gap-1.5 flex-shrink-0"
              on:click={() => projectStore.stop()}
            >
              <Square class="w-3.5 h-3.5 fill-current" />
              Force Close
            </button>
          {/if}
        </div>
      {/if}
    </div>

    <!-- Core Studio Workbenches -->
    <div class="space-y-3">
      <div class="flex items-center justify-between px-1">
        <h2 class="text-xs font-bold uppercase tracking-wider text-slate-400">Core Studio Workbenches</h2>
        <span class="text-xs text-slate-500">Full visual management suites for all Ikemen asset types</span>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <!-- 1: Roster Editor -->
        <div class="p-5 rounded-2xl bg-dark-800 border border-dark-600/70 hover:border-indigo-500/50 transition-all duration-200 flex flex-col justify-between group shadow-sm hover:shadow-indigo-950/20">
          <div class="space-y-3">
            <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-indigo-500/20 to-purple-500/20 text-indigo-400 border border-indigo-500/30 flex items-center justify-center group-hover:scale-105 transition">
              <Grid class="w-5 h-5" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h3 class="text-sm font-bold text-slate-100 group-hover:text-indigo-300 transition">Roster Editor</h3>
                <span class="text-[10px] font-mono text-slate-500">select.def</span>
              </div>
              <p class="text-xs text-slate-400 mt-1 leading-relaxed">
                Visual character grid, slot assignment, order, and extra stages manager.
              </p>
            </div>
          </div>

          <div class="pt-4 mt-4 border-t border-dark-600/50">
            <button
              type="button"
              class="w-full py-2 px-3 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold flex items-center justify-center gap-1.5 transition shadow"
              on:click={onOpenRosterEditor}
            >
              <span>Open Roster Editor</span>
              <ArrowRight class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        <!-- 2: Stages Manager -->
        <div class="p-5 rounded-2xl bg-dark-800 border border-dark-600/70 hover:border-emerald-500/50 transition-all duration-200 flex flex-col justify-between group shadow-sm hover:shadow-emerald-950/20">
          <div class="space-y-3">
            <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500/20 to-teal-500/20 text-emerald-400 border border-emerald-500/30 flex items-center justify-center group-hover:scale-105 transition">
              <Mountain class="w-5 h-5" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h3 class="text-sm font-bold text-slate-100 group-hover:text-emerald-300 transition">Stages Manager</h3>
                <span class="text-[10px] font-mono text-slate-500">stages/</span>
              </div>
              <p class="text-xs text-slate-400 mt-1 leading-relaxed">
                Arenas, ExtraStages registration, BGM assignments, and sparring match test.
              </p>
            </div>
          </div>

          <div class="pt-4 mt-4 border-t border-dark-600/50">
            <button
              type="button"
              class="w-full py-2 px-3 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-bold flex items-center justify-center gap-1.5 transition shadow"
              on:click={onOpenStages}
            >
              <span>Open Stage Manager</span>
              <ArrowRight class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        <!-- 3: Screenpacks & Motifs -->
        <div class="p-5 rounded-2xl bg-dark-800 border border-dark-600/70 hover:border-purple-500/50 transition-all duration-200 flex flex-col justify-between group shadow-sm hover:shadow-purple-950/20">
          <div class="space-y-3">
            <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500/20 to-indigo-500/20 text-purple-400 border border-purple-500/30 flex items-center justify-center group-hover:scale-105 transition">
              <Palette class="w-5 h-5" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h3 class="text-sm font-bold text-slate-100 group-hover:text-purple-300 transition">Screenpacks & Motifs</h3>
                <span class="text-[10px] font-mono text-slate-500">system.def</span>
              </div>
              <p class="text-xs text-slate-400 mt-1 leading-relaxed">
                Theme previews, select screen grid capacity, and 1-click active motif switcher.
              </p>
            </div>
          </div>

          <div class="pt-4 mt-4 border-t border-dark-600/50">
            <button
              type="button"
              class="w-full py-2 px-3 rounded-xl bg-purple-600 hover:bg-purple-500 text-white text-xs font-bold flex items-center justify-center gap-1.5 transition shadow"
              on:click={onOpenMotifs}
            >
              <span>Manage Motifs</span>
              <ArrowRight class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        <!-- 4: Lifebars & Fight HUD -->
        <div class="p-5 rounded-2xl bg-dark-800 border border-dark-600/70 hover:border-rose-500/50 transition-all duration-200 flex flex-col justify-between group shadow-sm hover:shadow-rose-950/20">
          <div class="space-y-3">
            <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-rose-500/20 to-red-500/20 text-rose-400 border border-rose-500/30 flex items-center justify-center group-hover:scale-105 transition">
              <Swords class="w-5 h-5" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h3 class="text-sm font-bold text-slate-100 group-hover:text-rose-300 transition">Lifebars & Fight HUD</h3>
                <span class="text-[10px] font-mono text-slate-500">fight.def</span>
              </div>
              <p class="text-xs text-slate-400 mt-1 leading-relaxed">
                Health bars, super gauge HUDs, round announcers, and active fight HUD switching.
              </p>
            </div>
          </div>

          <div class="pt-4 mt-4 border-t border-dark-600/50">
            <button
              type="button"
              class="w-full py-2 px-3 rounded-xl bg-rose-600 hover:bg-rose-500 text-white text-xs font-bold flex items-center justify-center gap-1.5 transition shadow"
              on:click={onOpenLifebars}
            >
              <span>Manage Lifebars</span>
              <ArrowRight class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        <!-- 5: Sound & Music Library -->
        <div class="p-5 rounded-2xl bg-dark-800 border border-dark-600/70 hover:border-rose-500/50 transition-all duration-200 flex flex-col justify-between group shadow-sm hover:shadow-rose-950/20">
          <div class="space-y-3">
            <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-rose-500/20 to-orange-500/20 text-rose-400 border border-rose-500/30 flex items-center justify-center group-hover:scale-105 transition">
              <Music class="w-5 h-5" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h3 class="text-sm font-bold text-slate-100 group-hover:text-rose-300 transition">Sound & Music</h3>
                <span class="text-[10px] font-mono text-slate-500">sound/</span>
              </div>
              <p class="text-xs text-slate-400 mt-1 leading-relaxed">
                Built-in BGM audio player and 1-click Title/Select/VS system music mapping.
              </p>
            </div>
          </div>

          <div class="pt-4 mt-4 border-t border-dark-600/50">
            <button
              type="button"
              class="w-full py-2 px-3 rounded-xl bg-dark-700 hover:bg-dark-650 hover:border-rose-500/40 border border-dark-600 text-rose-300 text-xs font-bold flex items-center justify-center gap-1.5 transition shadow-sm"
              on:click={onOpenSound}
            >
              <span>Open Sound Library</span>
              <ArrowRight class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        <!-- 6: Fonts & Typography -->
        <div class="p-5 rounded-2xl bg-dark-800 border border-dark-600/70 hover:border-purple-500/50 transition-all duration-200 flex flex-col justify-between group shadow-sm hover:shadow-purple-950/20">
          <div class="space-y-3">
            <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500/20 to-pink-500/20 text-purple-400 border border-purple-500/30 flex items-center justify-center group-hover:scale-105 transition">
              <Type class="w-5 h-5" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h3 class="text-sm font-bold text-slate-100 group-hover:text-purple-300 transition">Fonts & Typography</h3>
                <span class="text-[10px] font-mono text-slate-500">font/</span>
              </div>
              <p class="text-xs text-slate-400 mt-1 leading-relaxed">
                Specimen visualizer and font slot assignment for menu and fight displays.
              </p>
            </div>
          </div>

          <div class="pt-4 mt-4 border-t border-dark-600/50">
            <button
              type="button"
              class="w-full py-2 px-3 rounded-xl bg-dark-700 hover:bg-dark-650 hover:border-purple-500/40 border border-dark-600 text-purple-300 text-xs font-bold flex items-center justify-center gap-1.5 transition shadow-sm"
              on:click={onOpenFonts}
            >
              <span>Manage Fonts</span>
              <ArrowRight class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        <!-- 7: Storyboards & Cutscenes -->
        <div class="p-5 rounded-2xl bg-dark-800 border border-dark-600/70 hover:border-amber-500/50 transition-all duration-200 flex flex-col justify-between group shadow-sm hover:shadow-amber-950/20">
          <div class="space-y-3">
            <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-amber-500/20 to-yellow-500/20 text-amber-400 border border-amber-500/30 flex items-center justify-center group-hover:scale-105 transition">
              <Film class="w-5 h-5" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h3 class="text-sm font-bold text-slate-100 group-hover:text-amber-300 transition">Storyboards</h3>
                <span class="text-[10px] font-mono text-slate-500">data/*.def</span>
              </div>
              <p class="text-xs text-slate-400 mt-1 leading-relaxed">
                Cinematic cutscenes, opening intros, ending sequences, and credits rolls.
              </p>
            </div>
          </div>

          <div class="pt-4 mt-4 border-t border-dark-600/50">
            <button
              type="button"
              class="w-full py-2 px-3 rounded-xl bg-dark-700 hover:bg-dark-650 hover:border-amber-500/40 border border-dark-600 text-amber-300 text-xs font-bold flex items-center justify-center gap-1.5 transition shadow-sm"
              on:click={onOpenStoryboards}
            >
              <span>Manage Storyboards</span>
              <ArrowRight class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        <!-- 8: Maintenance & Repair Hub -->
        <div class="p-5 rounded-2xl bg-dark-800 border border-dark-600/70 hover:border-cyan-500/50 transition-all duration-200 flex flex-col justify-between group shadow-sm hover:shadow-cyan-950/20">
          <div class="space-y-3">
            <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-cyan-500/20 to-blue-500/20 text-cyan-400 border border-cyan-500/30 flex items-center justify-center group-hover:scale-105 transition">
              <Wrench class="w-5 h-5" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h3 class="text-sm font-bold text-slate-100 group-hover:text-cyan-300 transition">Repair Hub</h3>
                <span class="text-[10px] font-mono text-slate-500">Diagnostics</span>
              </div>
              <p class="text-xs text-slate-400 mt-1 leading-relaxed">
                Fix OpenGL 3.3 configs, sync stock files, verify integrity, and inspect diffs.
              </p>
            </div>
          </div>

          <div class="pt-4 mt-4 border-t border-dark-600/50">
            <button
              type="button"
              class="w-full py-2 px-3 rounded-xl bg-dark-700 hover:bg-dark-650 hover:border-cyan-500/40 border border-dark-600 text-cyan-300 text-xs font-bold flex items-center justify-center gap-1.5 transition shadow-sm"
              on:click={handleVerifyClick}
            >
              <ShieldCheck class="w-3.5 h-3.5" />
              <span>Open Repair Hub</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Project Asset Directories -->
    <div class="space-y-3">
      <div class="flex items-center justify-between px-1">
        <h2 class="text-xs font-bold uppercase tracking-wider text-slate-400">Project Asset Directories</h2>
        <span class="text-xs text-slate-500">Open on disk, manage directly, or link from Vault</span>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-3">
        {#each folderShortcuts as sc}
          <div
            class="p-4 rounded-xl bg-dark-800/90 border border-dark-600/60 hover:border-dark-600 flex flex-col justify-between gap-3 text-left transition group shadow-sm"
          >
            <div class="space-y-3">
              <div class="flex items-center justify-between">
                <div class="w-9 h-9 rounded-xl bg-gradient-to-br {sc.color} border flex items-center justify-center flex-shrink-0 group-hover:scale-105 transition">
                  <svelte:component this={sc.icon} class="w-4 h-4" />
                </div>
                <span class="font-mono text-[10px] text-slate-500 px-2 py-0.5 bg-dark-900 rounded-md border border-dark-700/50">
                  {sc.subpath}/
                </span>
              </div>

              <div>
                <h4 class="text-sm font-semibold text-slate-200 group-hover:text-white transition">
                  {sc.label}
                </h4>
                <p class="text-[11px] text-slate-400 line-clamp-2 mt-0.5 leading-snug">
                  {sc.desc}
                </p>
              </div>
            </div>

            <!-- Card Actions -->
            <div class="flex items-center gap-1.5 pt-2 border-t border-dark-600/40">
              <button
                type="button"
                class="py-1.5 px-2 bg-dark-700 hover:bg-dark-600 text-slate-300 text-[11px] font-medium rounded-lg transition flex items-center justify-center gap-1"
                on:click={() => projectStore.openFolder(sc.subpath)}
                title="Open {sc.subpath}/ directory in File Explorer"
              >
                <FolderOpen class="w-3 h-3 text-indigo-400" />
                <span>Open</span>
              </button>

              {#if sc.onManage}
                <button
                  type="button"
                  class="flex-1 py-1.5 px-2 bg-dark-700 hover:bg-dark-600 text-slate-200 text-[11px] font-semibold rounded-lg transition flex items-center justify-center gap-1"
                  on:click={sc.onManage}
                >
                  <span>{sc.manageLabel || 'Manage'}</span>
                </button>
              {/if}

              {#if sc.canAddVault && sc.category}
                <button
                  type="button"
                  class="py-1.5 px-2 bg-brand-500/10 hover:bg-brand-500/20 text-brand-300 text-[11px] font-medium rounded-lg border border-brand-500/20 transition flex items-center justify-center gap-1"
                  on:click={() => {
                    if (sc.category) handleOpenVaultFor(sc.category);
                  }}
                  title="Add {sc.label} from Vault library"
                >
                  <Plus class="w-3 h-3" />
                  <span>Vault</span>
                </button>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    </div>

    <!-- Direct File Inspector & Code Editor Quick Access -->
    <div class="p-5 rounded-2xl bg-dark-800 border border-dark-600/70 shadow-lg space-y-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2.5">
          <div class="w-8 h-8 rounded-xl bg-brand-500/20 text-brand-400 border border-brand-500/30 flex items-center justify-center">
            <FileCode class="w-4 h-4" />
          </div>
          <div>
            <h3 class="text-xs font-bold uppercase tracking-wider text-slate-200">Universal File Inspector & Code Editor</h3>
            <p class="text-[11px] text-slate-400">Directly inspect, analyze structural metrics, and live-edit any project file</p>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <button
            type="button"
            class="px-3 py-1.5 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600 text-slate-200 text-xs font-semibold flex items-center gap-1.5 transition"
            on:click={() => {
              const custom = prompt('Enter relative project path (e.g. data/select.def, chars/kfm/kfm.def, save/config.ini):', 'data/select.def');
              if (custom && custom.trim()) {
                selectedEditFilePath = custom.trim();
              }
            }}
          >
            <Search class="w-3.5 h-3.5 text-brand-400" />
            <span>Open Custom Path...</span>
          </button>
        </div>
      </div>

      <!-- Quick File Action Chips -->
      <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-2.5 pt-1">
        <button
          type="button"
          class="p-2.5 rounded-xl bg-dark-900/80 hover:bg-dark-700/80 border border-dark-700/80 hover:border-brand-500/40 text-left transition group"
          on:click={() => selectedEditFilePath = 'data/select.def'}
        >
          <div class="text-[10px] font-bold text-slate-400 group-hover:text-brand-300 transition">ROSTER FILE</div>
          <div class="text-xs font-mono font-bold text-white truncate">select.def</div>
        </button>

        <button
          type="button"
          class="p-2.5 rounded-xl bg-dark-900/80 hover:bg-dark-700/80 border border-dark-700/80 hover:border-purple-500/40 text-left transition group"
          on:click={() => selectedEditFilePath = 'data/system.def'}
        >
          <div class="text-[10px] font-bold text-slate-400 group-hover:text-purple-300 transition">SYSTEM MOTIF</div>
          <div class="text-xs font-mono font-bold text-white truncate">system.def</div>
        </button>

        <button
          type="button"
          class="p-2.5 rounded-xl bg-dark-900/80 hover:bg-dark-700/80 border border-dark-700/80 hover:border-rose-500/40 text-left transition group"
          on:click={() => selectedEditFilePath = 'data/fight.def'}
        >
          <div class="text-[10px] font-bold text-slate-400 group-hover:text-rose-300 transition">FIGHT HUD</div>
          <div class="text-xs font-mono font-bold text-white truncate">fight.def</div>
        </button>

        <button
          type="button"
          class="p-2.5 rounded-xl bg-dark-900/80 hover:bg-dark-700/80 border border-dark-700/80 hover:border-amber-500/40 text-left transition group"
          on:click={() => selectedEditFilePath = 'save/config.ini'}
        >
          <div class="text-[10px] font-bold text-slate-400 group-hover:text-amber-300 transition">CONFIG FILE</div>
          <div class="text-xs font-mono font-bold text-white truncate">config.ini</div>
        </button>

        <button
          type="button"
          class="p-2.5 rounded-xl bg-dark-900/80 hover:bg-dark-700/80 border border-dark-700/80 hover:border-emerald-500/40 text-left transition group"
          on:click={() => selectedEditFilePath = 'data/gofx/gofx.def'}
        >
          <div class="text-[10px] font-bold text-slate-400 group-hover:text-emerald-300 transition">GLOBAL FX</div>
          <div class="text-xs font-mono font-bold text-white truncate">gofx.def</div>
        </button>

        <button
          type="button"
          class="p-2.5 rounded-xl bg-dark-900/80 hover:bg-dark-700/80 border border-dark-700/80 hover:border-cyan-500/40 text-left transition group"
          on:click={() => selectedEditFilePath = 'external/gamecontrollerdb.txt'}
        >
          <div class="text-[10px] font-bold text-slate-400 group-hover:text-cyan-300 transition">CONTROLLERS</div>
          <div class="text-xs font-mono font-bold text-white truncate">gamecontrollerdb</div>
        </button>
      </div>
    </div>


    <!-- Engine Console & Log Preview Console -->
    <div class="p-5 rounded-2xl bg-dark-800 border border-dark-600/70 shadow-lg space-y-3">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <Terminal class="w-4 h-4 text-amber-400" />
          <h3 class="text-xs font-bold uppercase tracking-wider text-slate-300">Engine Runtime Output & Logs</h3>
          <span class="font-mono text-[10px] text-slate-500">save/ikemen.log</span>
        </div>

        <div class="flex items-center gap-2">
          <button
            type="button"
            class="px-2.5 py-1 rounded-lg bg-dark-700 hover:bg-dark-600 text-slate-300 text-[11px] font-semibold flex items-center gap-1.5 transition"
            on:click={fetchRecentLogs}
            disabled={isRefreshingLogs}
            title="Reload recent logs"
          >
            <RefreshCw class="w-3 h-3 text-slate-400 {isRefreshingLogs ? 'animate-spin' : ''}" />
            <span>Refresh</span>
          </button>

          <button
            type="button"
            class="px-2.5 py-1 rounded-lg bg-dark-700 hover:bg-dark-600 text-slate-300 text-[11px] font-semibold flex items-center gap-1.5 transition"
            on:click={() => projectStore.openLogs()}
            title="Open log files in directory"
          >
            <ExternalLink class="w-3 h-3 text-indigo-400" />
            <span>Open Folder</span>
          </button>
        </div>
      </div>

      <!-- Log Terminal Box -->
      <div class="p-3.5 rounded-xl bg-dark-900 border border-dark-700/80 font-mono text-[11px] text-slate-300 h-32 overflow-y-auto leading-relaxed select-text">
        {#if recentLogs && recentLogs.trim().length > 0}
          <pre class="whitespace-pre-wrap">{recentLogs}</pre>
        {:else}
          <div class="h-full flex items-center justify-center text-slate-500 italic text-xs">
            No runtime logs recorded yet. Launch the game or inspect save/ folder.
          </div>
        {/if}
      </div>
    </div>

    <!-- Bottom Navigation Footer -->
    <div class="pt-2 flex items-center justify-between border-t border-dark-600/40">
      <button
        type="button"
        class="text-xs font-semibold text-indigo-400 hover:text-indigo-300 transition flex items-center gap-1"
        on:click={onBackToProjects}
      >
        &larr; Back to all projects
      </button>

      <button
        type="button"
        disabled={isBusy}
        class="flex items-center gap-1.5 px-3 py-1.5 text-xs text-slate-500 hover:text-slate-300 hover:bg-dark-800 disabled:opacity-40 disabled:cursor-not-allowed rounded-lg transition"
        on:click={() => {
          projectStore.close();
          onBackToProjects();
        }}
      >
        <XCircle class="w-4 h-4" />
        Close Active Project
      </button>
    </div>
  </div>

  <!-- Pre-Verification Options & Confirmation Modal -->
  {#if showVerifyOptionsModal}
    <VerifyOptionsModal
      onClose={() => (showVerifyOptionsModal = false)}
      onReport={handleReportReceived}
    />
  {/if}

  <!-- Verification Report Modal -->
  {#if showVerifyReportModal}
    <VerifyReportModal
      report={verificationReport}
      isLoading={false}
      onClose={() => (showVerifyReportModal = false)}
    />
  {/if}

  <!-- Custom Launch Modal -->
  {#if showCustomLaunchModal}
    <CustomLaunchModal onClose={() => (showCustomLaunchModal = false)} />
  {/if}

  <!-- Game Config (config.ini) Modal -->
  {#if showGameConfigModal}
    <GameConfigModal onClose={() => (showGameConfigModal = false)} />
  {/if}

  <!-- Engine Switch Confirmation Modal -->
  {#if showEngineConfirmModal}
    <ConfirmModal
      title="Switch Project Engine Version?"
      message="Studio will create a safety backup of your runtime files in save/backups/ and migrate your project to {pendingEngineVersion}. Your characters and stages will not be affected."
      confirmLabel="Switch Engine"
      confirmVariant="primary"
      onConfirm={confirmEngineSwitch}
      onCancel={cancelEngineSwitch}
    />
  {/if}

  <!-- Rollback Engine Confirmation Modal -->
  {#if showRollbackModal}
    <ConfirmModal
      title="Rollback Engine Version?"
      message="This will restore the previous engine runtime snapshot from save/backups/."
      confirmLabel="Rollback"
      confirmVariant="primary"
      onConfirm={confirmRollback}
      onCancel={() => (showRollbackModal = false)}
    />
  {/if}

  <!-- Add From Vault Modal -->
  {#if showAddFromVaultModal}
    <AddFromVaultModal
      isOpen={showAddFromVaultModal}
      projectDir={$projectStore.current.path}
      targetCategory={vaultTargetCategory}
      onClose={() => (showAddFromVaultModal = false)}
    />
  {/if}

  <!-- Universal File Editor Modal -->
  {#if selectedEditFilePath}
    <FileEditorModal
      projectDir={$projectStore.current.path}
      filePath={selectedEditFilePath}
      onClose={() => (selectedEditFilePath = null)}
    />
  {/if}
{/if}

