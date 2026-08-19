<script lang="ts">
  import { onMount } from 'svelte';
  import { projectStore } from './lib/stores/projectStore';
  import { engineStore } from './lib/stores/engineStore';
  import { settingsStore } from './lib/stores/settingsStore';

  import Header from './lib/components/Header.svelte';
  import ProjectDashboard from './lib/components/ProjectDashboard.svelte';
  import EngineManager from './lib/components/EngineManager.svelte';
  import NewProjectModal from './lib/components/NewProjectModal.svelte';
  import OpenProjectModal from './lib/components/OpenProjectModal.svelte';
  import SettingsModal from './lib/components/SettingsModal.svelte';
  import Toast from './lib/components/Toast.svelte';

  import { Plus, FolderOpen, Layers, Sparkles, Gamepad2, Clock, ChevronRight, HardDrive } from 'lucide-svelte';

  let showEnginesModal = false;
  let showSettingsModal = false;
  let showNewProjectModal = false;
  let showOpenProjectModal = false;

  onMount(async () => {
    await settingsStore.load();
    await engineStore.loadInstalled();
    await projectStore.loadRecent();
  });
</script>

<div class="flex flex-col h-screen w-screen bg-dark-900 text-slate-100 overflow-hidden select-none font-sans">
  <!-- Top Navigation Header -->
  <Header
    onOpenEngines={() => (showEnginesModal = true)}
    onOpenSettings={() => (showSettingsModal = true)}
    onOpenNewProject={() => (showNewProjectModal = true)}
    onOpenExistingProject={() => (showOpenProjectModal = true)}
  />

  <!-- Main View Area -->
  <main class="flex-1 overflow-y-auto">
    {#if $projectStore.current}
      <ProjectDashboard />
    {:else}
      <!-- Empty State / Welcome Screen -->
      <div class="max-w-4xl mx-auto py-12 px-6 space-y-10">
        <!-- Hero Header -->
        <div class="text-center space-y-3">
          <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-indigo-500 to-purple-600 shadow-xl shadow-indigo-500/10 mb-2">
            <Gamepad2 class="w-9 h-9 text-white" />
          </div>
          <h1 class="text-3xl font-extrabold tracking-tight text-white">
            Welcome to Ikemen GO Studio
          </h1>
          <p class="text-sm text-slate-400 max-w-lg mx-auto leading-relaxed">
            The cross-platform developer environment and engine manager for building modern 2D fighting games.
          </p>
        </div>

        <!-- Primary Action Cards -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <!-- New Project -->
          <button
            type="button"
            class="p-6 rounded-2xl bg-dark-800/90 hover:bg-dark-700/80 border border-dark-600/60 hover:border-indigo-500/50 flex flex-col items-start gap-4 text-left transition group shadow-lg shadow-black/20"
            on:click={() => (showNewProjectModal = true)}
          >
            <div class="w-12 h-12 rounded-xl bg-indigo-500/10 text-indigo-400 group-hover:bg-indigo-600 group-hover:text-white flex items-center justify-center transition">
              <Plus class="w-6 h-6" />
            </div>
            <div>
              <h3 class="text-base font-bold text-slate-100 group-hover:text-white transition">New Project</h3>
              <p class="text-xs text-slate-400 mt-1 leading-relaxed">
                Scaffold a fresh Ikemen GO project with standard folder layout and base data.
              </p>
            </div>
          </button>

          <!-- Open Project -->
          <button
            type="button"
            class="p-6 rounded-2xl bg-dark-800/90 hover:bg-dark-700/80 border border-dark-600/60 hover:border-indigo-500/50 flex flex-col items-start gap-4 text-left transition group shadow-lg shadow-black/20"
            on:click={() => (showOpenProjectModal = true)}
          >
            <div class="w-12 h-12 rounded-xl bg-purple-500/10 text-purple-400 group-hover:bg-purple-600 group-hover:text-white flex items-center justify-center transition">
              <FolderOpen class="w-6 h-6" />
            </div>
            <div>
              <h3 class="text-base font-bold text-slate-100 group-hover:text-white transition">Open Project</h3>
              <p class="text-xs text-slate-400 mt-1 leading-relaxed">
                Open an existing Ikemen GO project directory or select from recent workspaces.
              </p>
            </div>
          </button>

          <!-- Manage Engines -->
          <button
            type="button"
            class="p-6 rounded-2xl bg-dark-800/90 hover:bg-dark-700/80 border border-dark-600/60 hover:border-indigo-500/50 flex flex-col items-start gap-4 text-left transition group shadow-lg shadow-black/20"
            on:click={() => (showEnginesModal = true)}
          >
            <div class="w-12 h-12 rounded-xl bg-cyan-500/10 text-cyan-400 group-hover:bg-cyan-600 group-hover:text-white flex items-center justify-center transition">
              <Layers class="w-6 h-6" />
            </div>
            <div>
              <div class="flex items-center justify-between">
                <h3 class="text-base font-bold text-slate-100 group-hover:text-white transition">Engines</h3>
                <span class="text-[10px] font-mono px-2 py-0.5 rounded-full bg-dark-700 text-slate-300">
                  {$engineStore.installed.length} installed
                </span>
              </div>
              <p class="text-xs text-slate-400 mt-1 leading-relaxed">
                Discover, download, and manage official GitHub Ikemen GO releases.
              </p>
            </div>
          </button>
        </div>

        <!-- Recent Projects Section -->
        {#if $projectStore.recent.length > 0}
          <div class="space-y-3">
            <div class="flex items-center gap-2 text-xs font-bold text-slate-400 uppercase tracking-wider">
              <Clock class="w-3.5 h-3.5" />
              <span>Recent Projects</span>
            </div>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
              {#each $projectStore.recent.slice(0, 4) as path}
                <button
                  type="button"
                  class="p-3.5 rounded-xl bg-dark-800/60 hover:bg-dark-800 border border-dark-600/40 hover:border-dark-600 flex items-center justify-between text-left transition group"
                  on:click={() => projectStore.open(path)}
                >
                  <div class="flex items-center gap-3 min-w-0">
                    <div class="w-8 h-8 rounded-lg bg-dark-700 text-slate-400 group-hover:text-indigo-400 flex items-center justify-center flex-shrink-0 transition">
                      <HardDrive class="w-4 h-4" />
                    </div>
                    <div class="min-w-0">
                      <div class="text-xs font-semibold text-slate-200 group-hover:text-white truncate">
                        {path.split(/[/\\]/).pop()}
                      </div>
                      <div class="text-[11px] font-mono text-slate-500 truncate">{path}</div>
                    </div>
                  </div>
                  <ChevronRight class="w-4 h-4 text-slate-600 group-hover:text-slate-300 transition" />
                </button>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    {/if}
  </main>

  <!-- Modals -->
  {#if showEnginesModal}
    <EngineManager onClose={() => (showEnginesModal = false)} />
  {/if}

  {#if showNewProjectModal}
    <NewProjectModal
      onClose={() => (showNewProjectModal = false)}
      onOpenEngines={() => (showEnginesModal = true)}
    />
  {/if}

  {#if showOpenProjectModal}
    <OpenProjectModal onClose={() => (showOpenProjectModal = false)} />
  {/if}

  {#if showSettingsModal}
    <SettingsModal onClose={() => (showSettingsModal = false)} />
  {/if}

  <!-- Toast Notification Overlay -->
  <Toast />
</div>
