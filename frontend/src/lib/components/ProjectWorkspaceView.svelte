<script lang="ts">
  import { onMount } from 'svelte';
  import { projectStore } from '../stores/projectStore';
  import { engineStore } from '../stores/engineStore';
  import type { VerificationReport } from '../types';
  import VerifyOptionsModal from './VerifyOptionsModal.svelte';
  import VerifyReportModal from './VerifyReportModal.svelte';
  import CustomLaunchModal from './CustomLaunchModal.svelte';
  import GameConfigModal from './GameConfigModal.svelte';
  import ConfirmModal from './ConfirmModal.svelte';
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
  } from 'lucide-svelte';

  export let onBackToProjects: () => void;
  export let onOpenRepairHub: () => void;

  let showVerifyOptionsModal = false;
  let showVerifyReportModal = false;
  let showCustomLaunchModal = false;
  let showGameConfigModal = false;
  let showEngineConfirmModal = false;
  let showRollbackModal = false;
  let selectedBackupId = '';
  let pendingEngineVersion = '';
  let selectedEngineVersion = '';

  $: if ($projectStore.current?.engine?.version && !showEngineConfirmModal) {
    selectedEngineVersion = $projectStore.current.engine.version;
  }

  let verificationReport: VerificationReport | null = null;
  let isVerifying = false;
  let showPlayDropdown = false;

  const folderShortcuts = [
    { label: 'Characters', subpath: 'chars', icon: Users, desc: 'Fighter packages and .def files', color: 'from-blue-500/20 to-cyan-500/20 text-cyan-400 border-cyan-500/30' },
    { label: 'Stages', subpath: 'stages', icon: Mountain, desc: 'Background arenas and music defs', color: 'from-emerald-500/20 to-teal-500/20 text-emerald-400 border-emerald-500/30' },
    { label: 'System Data', subpath: 'data', icon: FileCode, desc: 'select.def, system.def, fonts', color: 'from-amber-500/20 to-orange-500/20 text-amber-400 border-amber-500/30' },
    { label: 'Fonts', subpath: 'font', icon: Type, desc: 'Bitmap and TrueType font assets', color: 'from-purple-500/20 to-pink-500/20 text-purple-400 border-purple-500/30' },
    { label: 'Sound & Music', subpath: 'sound', icon: Music, desc: 'BGM tracks, hits, and announcer voices', color: 'from-rose-500/20 to-red-500/20 text-rose-400 border-rose-500/30' },
  ];

  function handleVerifyClick() {
    onOpenRepairHub();
  }

  function handleReportReceived(report: VerificationReport) {
    verificationReport = report;
    showVerifyReportModal = true;
  }

  async function handleLaunchMode(mode: 'normal' | 'training' | 'debug') {
    showPlayDropdown = false;
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
    }
  }

  async function confirmRollback() {
    showRollbackModal = false;
    if (selectedBackupId) {
      await projectStore.rollbackEngine(selectedBackupId);
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

<svelte:window on:click={() => (showPlayDropdown = false)} />

{#if $projectStore.current}
  <div class="max-w-6xl mx-auto p-8 space-y-6">
    <!-- Top Hero Banner / Project Details -->
    <div class="p-6 rounded-2xl bg-dark-800 border {$projectStore.gameState === 'running' ? 'border-emerald-500/50 shadow-emerald-950/30' : 'border-dark-600/70'} shadow-lg relative overflow-visible z-20 transition-all duration-300">
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-6 relative z-10">
        <div class="space-y-2">
          <div class="flex items-center gap-3 flex-wrap">
            <h1 class="text-2xl font-black text-slate-100">{$projectStore.current.name}</h1>
            
            <!-- Engine Version Selector with Switcher -->
            <div class="flex items-center gap-1.5">
              <select
                bind:value={selectedEngineVersion}
                on:change={handleEngineSelect}
                disabled={isBusy}
                class="text-xs font-mono px-2.5 py-1 rounded-full bg-purple-500/10 text-purple-300 border border-purple-500/30 font-semibold focus:outline-none focus:border-purple-400 cursor-pointer disabled:opacity-50"
                title="Change project engine version (with automated backup)"
              >
                {#each $engineStore.installed as engine}
                  <option value={engine.version}>{engine.version} ({engine.channel})</option>
                {/each}
              </select>

              <!-- Rollback Button if backups exist -->
              {#if $projectStore.backups && $projectStore.backups.length > 0}
                <button
                  type="button"
                  class="p-1 rounded-full bg-dark-700 hover:bg-dark-600 text-purple-400 border border-dark-600/60 transition"
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

            {#if $projectStore.gameState === 'running'}
              <span class="flex items-center gap-1.5 text-xs px-2.5 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 font-semibold">
                <span class="w-2 h-2 rounded-full bg-emerald-400 animate-ping"></span>
                Game Active
              </span>
            {:else if $projectStore.gameState === 'starting'}
              <span class="flex items-center gap-1.5 text-xs px-2.5 py-0.5 rounded-full bg-amber-500/10 text-amber-400 border border-amber-500/30 font-semibold">
                <Loader2 class="w-3 h-3 animate-spin" />
                Spawning Process...
              </span>
            {/if}
          </div>

          <div class="flex items-center gap-4 text-xs text-slate-400 flex-wrap">
            {#if $projectStore.current.author}
              <div class="flex items-center gap-1.5">
                <User class="w-3.5 h-3.5 text-slate-500" />
                <span>Author: <strong class="text-slate-300">{$projectStore.current.author}</strong></span>
              </div>
            {/if}
            <div class="flex items-center gap-1.5">
              <Calendar class="w-3.5 h-3.5 text-slate-500" />
              <span>Created: {formatDate($projectStore.current.created_at)}</span>
            </div>
            <div class="flex items-center gap-1.5">
              <HardDrive class="w-3.5 h-3.5 text-slate-500" />
              <span class="font-mono truncate max-w-sm">{$projectStore.current.path}</span>
            </div>
          </div>
        </div>

        <!-- Action Controls -->
        <div class="flex items-center gap-2 flex-wrap">
          <!-- Game Settings (config.ini) -->
          <button
            type="button"
            class="px-3 py-2.5 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600/70 text-slate-200 text-xs font-semibold flex items-center gap-1.5 transition shadow-sm"
            on:click={() => (showGameConfigModal = true)}
            title="Edit Game Settings & save/config.ini"
          >
            <Sliders class="w-3.5 h-3.5 text-cyan-400" />
            <span>Config</span>
          </button>

          <!-- Explorer -->
          <button
            type="button"
            class="px-3 py-2.5 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600/70 text-slate-200 text-xs font-semibold flex items-center gap-1.5 transition shadow-sm"
            on:click={() => projectStore.openFolder()}
            title="Open project directory in File Explorer"
          >
            <Folder class="w-3.5 h-3.5 text-indigo-400" />
            <span>Explorer</span>
          </button>

          <!-- Logs -->
          <button
            type="button"
            class="px-3 py-2.5 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600/70 text-slate-200 text-xs font-semibold flex items-center gap-1.5 transition shadow-sm"
            on:click={() => projectStore.openLogs()}
            title="Open project logs folder"
          >
            <FileText class="w-3.5 h-3.5 text-amber-400" />
            <span>Logs</span>
          </button>

          <!-- Repair Hub -->
          <button
            type="button"
            disabled={isVerifying || isBusy}
            class="px-3 py-2.5 rounded-xl bg-dark-700 hover:bg-dark-600 disabled:opacity-50 border border-dark-600/70 text-slate-200 text-xs font-semibold flex items-center gap-1.5 transition shadow-sm"
            on:click={handleVerifyClick}
            title="Open Maintenance & Repair Hub"
          >
            {#if isVerifying}
              <Loader2 class="w-3.5 h-3.5 animate-spin text-indigo-400" />
              <span>Checking...</span>
            {:else}
              <Wrench class="w-3.5 h-3.5 text-purple-400" />
              <span>Repair Hub</span>
            {/if}
          </button>

          <!-- Split Play / Launch Button -->
          {#if $projectStore.gameState === 'starting'}
            <button
              type="button"
              disabled
              class="px-5 py-2.5 rounded-xl font-bold text-sm bg-dark-700 border border-dark-600 text-slate-400 cursor-not-allowed flex items-center gap-2 transition"
            >
              <Loader2 class="w-4 h-4 animate-spin text-amber-400" />
              <span>Starting...</span>
            </button>
          {:else if $projectStore.gameState === 'stopping'}
            <button
              type="button"
              disabled
              class="px-5 py-2.5 rounded-xl font-bold text-sm bg-dark-700 border border-dark-600 text-slate-400 cursor-not-allowed flex items-center gap-2 transition"
            >
              <Loader2 class="w-4 h-4 animate-spin text-rose-400" />
              <span>Stopping...</span>
            </button>
          {:else if $projectStore.gameState === 'running'}
            <button
              type="button"
              disabled={!$projectStore.canStop}
              class="px-5 py-2.5 rounded-xl font-bold text-sm shadow-md flex items-center gap-2 transition {
                $projectStore.canStop
                  ? 'bg-rose-600 hover:bg-rose-500 text-white shadow-rose-950/50'
                  : 'bg-rose-900/60 text-rose-300 border border-rose-800/50 cursor-wait'
              }"
              on:click={() => projectStore.stop()}
            >
              <Square class="w-4 h-4 fill-current" />
              <span>{$projectStore.canStop ? 'Stop Game' : 'Game Active...'}</span>
            </button>
          {:else}
            <!-- Play Split Group -->
            <div class="relative flex items-center" on:click|stopPropagation>
              <button
                type="button"
                class="px-5 py-2.5 rounded-l-xl font-bold text-sm shadow-md bg-emerald-600 hover:bg-emerald-500 text-white shadow-emerald-950/50 flex items-center gap-2 transition"
                on:click={() => handleLaunchMode('normal')}
              >
                <Play class="w-4 h-4 fill-current" />
                <span>Launch Game</span>
              </button>

              <button
                type="button"
                class="px-2.5 py-2.5 rounded-r-xl bg-emerald-700 hover:bg-emerald-600 border-l border-emerald-500/40 text-white transition"
                on:click={() => (showPlayDropdown = !showPlayDropdown)}
                title="Launch Options & Presets"
              >
                <ChevronDown class="w-4 h-4" />
              </button>

              <!-- Launch Dropdown Menu -->
              {#if showPlayDropdown}
                <div class="absolute right-0 top-full mt-2 w-64 rounded-2xl bg-dark-800 border border-dark-600/80 shadow-2xl p-1.5 z-50 animate-in fade-in zoom-in-95 duration-100">
                  <button
                    type="button"
                    class="w-full px-3 py-2.5 rounded-xl text-left hover:bg-dark-700/80 text-xs font-semibold text-slate-200 flex items-center gap-2.5 transition"
                    on:click={() => handleLaunchMode('normal')}
                  >
                    <Play class="w-3.5 h-3.5 text-emerald-400 fill-current" />
                    <div>
                      <div>Normal Game</div>
                      <div class="text-[10px] text-slate-500">Standard arcade launcher</div>
                    </div>
                  </button>

                  <button
                    type="button"
                    class="w-full px-3 py-2.5 rounded-xl text-left hover:bg-dark-700/80 text-xs font-semibold text-slate-200 flex items-center gap-2.5 transition"
                    on:click={() => handleLaunchMode('training')}
                  >
                    <Sparkles class="w-3.5 h-3.5 text-indigo-400" />
                    <div>
                      <div>Direct Training / Sparring</div>
                      <div class="text-[10px] text-slate-500 font-mono">-p1 ... -p2 ... -time -1</div>
                    </div>
                  </button>

                  <button
                    type="button"
                    class="w-full px-3 py-2.5 rounded-xl text-left hover:bg-dark-700/80 text-xs font-semibold text-slate-200 flex items-center gap-2.5 transition"
                    on:click={() => handleLaunchMode('debug')}
                  >
                    <Terminal class="w-3.5 h-3.5 text-amber-400" />
                    <div>
                      <div>Developer / Debug Mode</div>
                      <div class="text-[10px] text-slate-500 font-mono">-debug -maxpowermode</div>
                    </div>
                  </button>

                  <div class="h-px bg-dark-600/50 my-1"></div>

                  <button
                    type="button"
                    class="w-full px-3 py-2.5 rounded-xl text-left hover:bg-dark-700/80 text-xs font-semibold text-purple-300 flex items-center gap-2.5 transition"
                    on:click={() => {
                      showPlayDropdown = false;
                      showCustomLaunchModal = true;
                    }}
                  >
                    <Sliders class="w-3.5 h-3.5 text-purple-400" />
                    <div>
                      <div>Custom Launch Options...</div>
                      <div class="text-[10px] text-slate-500">Flags & argument history</div>
                    </div>
                  </button>
                </div>
              {/if}
            </div>
          {/if}
        </div>
      </div>
    </div>

    <!-- Active Session Notification Bar -->
    {#if $projectStore.gameState === 'running'}
      <div class="p-4 rounded-xl bg-emerald-950/40 border border-emerald-600/40 flex items-center justify-between gap-4">
        <div class="flex items-center gap-3">
          <div class="w-8 h-8 rounded-lg bg-emerald-500/20 text-emerald-400 flex items-center justify-center">
            <Activity class="w-4 h-4 animate-pulse" />
          </div>
          <div>
            <div class="text-xs font-bold text-emerald-200">Ikemen GO is actively running</div>
            <div class="text-[11px] text-emerald-400/80">Play and test in the engine window. Use Stop Game when done.</div>
          </div>
        </div>
        {#if $projectStore.canStop}
          <button
            type="button"
            class="px-3 py-1.5 bg-rose-600 hover:bg-rose-500 text-white text-xs font-semibold rounded-lg shadow transition flex items-center gap-1.5"
            on:click={() => projectStore.stop()}
          >
            <Square class="w-3.5 h-3.5 fill-current" />
            Force Close
          </button>
        {/if}
      </div>
    {/if}

    <!-- Quick Folder Access Grid -->
    <div>
      <h2 class="text-xs font-bold uppercase tracking-wider text-slate-400 mb-3 px-1">Project Asset Directories</h2>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
        {#each folderShortcuts as sc}
          <button
            type="button"
            class="p-4 rounded-xl bg-dark-800/80 hover:bg-dark-700/80 border border-dark-600/60 hover:border-dark-600 flex items-start gap-3.5 text-left transition group shadow-sm"
            on:click={() => projectStore.openFolder(sc.subpath)}
          >
            <div class="w-10 h-10 rounded-xl bg-gradient-to-br {sc.color} border flex items-center justify-center flex-shrink-0 group-hover:scale-105 transition">
              {#if sc.subpath === 'chars'}
                <Users class="w-5 h-5" />
              {:else if sc.subpath === 'stages'}
                <Mountain class="w-5 h-5" />
              {:else if sc.subpath === 'data'}
                <FileCode class="w-5 h-5" />
              {:else if sc.subpath === 'font'}
                <Type class="w-5 h-5" />
              {:else}
                <Music class="w-5 h-5" />
              {/if}
            </div>
            <div class="space-y-0.5 min-w-0 flex-1">
              <div class="flex items-center justify-between">
                <span class="text-sm font-semibold text-slate-200 group-hover:text-white transition">{sc.label}</span>
                <span class="font-mono text-[10px] text-slate-500">{sc.subpath}/</span>
              </div>
              <p class="text-xs text-slate-400 line-clamp-1">{sc.desc}</p>
            </div>
          </button>
        {/each}

        <!-- select.def roster quick view -->
        <button
          type="button"
          class="p-4 rounded-xl bg-dark-800/80 hover:bg-dark-700/80 border border-dark-600/60 hover:border-dark-600 flex items-start gap-3.5 text-left transition group shadow-sm"
          on:click={() => projectStore.openFolder('data')}
        >
          <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-indigo-500/20 to-purple-500/20 text-indigo-400 border border-indigo-500/30 flex items-center justify-center flex-shrink-0 group-hover:scale-105 transition">
            <FileCode class="w-5 h-5" />
          </div>
          <div class="space-y-0.5 min-w-0 flex-1">
            <div class="flex items-center justify-between">
              <span class="text-sm font-semibold text-slate-200 group-hover:text-white transition">Roster (select.def)</span>
              <span class="font-mono text-[10px] text-slate-500">data/</span>
            </div>
            <p class="text-xs text-slate-400 line-clamp-1">Character & Stage roster configuration</p>
          </div>
        </button>
      </div>
    </div>

    <!-- Bottom Actions -->
    <div class="pt-4 flex items-center justify-between border-t border-dark-600/40">
      <button
        type="button"
        class="text-xs font-semibold text-indigo-400 hover:text-indigo-300 transition"
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
{/if}
