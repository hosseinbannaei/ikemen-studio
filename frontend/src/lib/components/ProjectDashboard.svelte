<script lang="ts">
  import { projectStore } from '../stores/projectStore';
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
  } from 'lucide-svelte';

  const folderShortcuts = [
    { label: 'Characters', subpath: 'chars', icon: Users, desc: 'Fighter packages and .def files', color: 'from-blue-500/20 to-cyan-500/20 text-cyan-400 border-cyan-500/30' },
    { label: 'Stages', subpath: 'stages', icon: Mountain, desc: 'Background arenas and music defs', color: 'from-emerald-500/20 to-teal-500/20 text-emerald-400 border-emerald-500/30' },
    { label: 'System Data', subpath: 'data', icon: FileCode, desc: 'select.def, system.def, fonts', color: 'from-amber-500/20 to-orange-500/20 text-amber-400 border-amber-500/30' },
    { label: 'Fonts', subpath: 'font', icon: Type, desc: 'Bitmap and TrueType font assets', color: 'from-purple-500/20 to-pink-500/20 text-purple-400 border-purple-500/30' },
    { label: 'Sound & Music', subpath: 'sound', icon: Music, desc: 'BGM tracks, hits, and announcer voices', color: 'from-rose-500/20 to-red-500/20 text-rose-400 border-rose-500/30' },
  ];

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
</script>

{#if $projectStore.current}
  <div class="max-w-6xl mx-auto p-6 space-y-6">
    <!-- Top Hero Banner / Project Details -->
    <div class="p-6 rounded-2xl bg-dark-800 border border-dark-600/70 shadow-lg relative overflow-hidden">
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-6 relative z-10">
        <div class="space-y-2">
          <div class="flex items-center gap-3">
            <h1 class="text-2xl font-bold text-slate-100">{$projectStore.current.name}</h1>
            <span class="text-xs font-mono px-2.5 py-0.5 rounded-full bg-purple-500/10 text-purple-300 border border-purple-500/20">
              {$projectStore.current.engine.version} ({$projectStore.current.engine.channel})
            </span>
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

        <!-- Launch / Stop Actions -->
        <div class="flex items-center gap-3">
          <button
            type="button"
            class="px-3.5 py-2 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600/70 text-slate-200 text-xs font-medium flex items-center gap-2 transition"
            on:click={() => projectStore.openFolder()}
          >
            <Folder class="w-4 h-4 text-indigo-400" />
            <span>Open in Explorer</span>
          </button>

          <button
            type="button"
            class="px-6 py-2.5 rounded-xl font-bold text-sm shadow-md flex items-center gap-2.5 transition {
              $projectStore.isRunning
                ? 'bg-rose-600 hover:bg-rose-500 text-white shadow-rose-950/50'
                : 'bg-emerald-600 hover:bg-emerald-500 text-white shadow-emerald-950/50'
            }"
            on:click={() => ($projectStore.isRunning ? projectStore.stop() : projectStore.launch())}
          >
            {#if $projectStore.isRunning}
              <Square class="w-4 h-4 fill-current" />
              <span>Stop Game</span>
            {:else}
              <Play class="w-4 h-4 fill-current" />
              <span>Launch Game</span>
            {/if}
          </button>
        </div>
      </div>
    </div>

    <!-- Quick Folder Access Grid -->
    <div>
      <h2 class="text-sm font-bold uppercase tracking-wider text-slate-400 mb-3">Project Asset Directories</h2>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
        {#each folderShortcuts as sc}
          <button
            type="button"
            class="p-4 rounded-xl bg-dark-800/80 hover:bg-dark-700/80 border border-dark-600/60 hover:border-dark-600 flex items-start gap-3.5 text-left transition group"
            on:click={() => projectStore.openFolder(sc.subpath)}
          >
            <div class="w-10 h-10 rounded-xl bg-gradient-to-br {sc.color} border flex items-center justify-center flex-shrink-0 group-hover:scale-105 transition">
              <svelte:component this={sc.icon} class="w-5 h-5" />
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

        <!-- select.def quick view -->
        <button
          type="button"
          class="p-4 rounded-xl bg-dark-800/80 hover:bg-dark-700/80 border border-dark-600/60 hover:border-dark-600 flex items-start gap-3.5 text-left transition group"
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

    <!-- Close Project Footer -->
    <div class="pt-4 flex justify-end">
      <button
        type="button"
        class="flex items-center gap-1.5 px-3 py-1.5 text-xs text-slate-500 hover:text-slate-300 hover:bg-dark-800 rounded-lg transition"
        on:click={() => projectStore.close()}
      >
        <XCircle class="w-4 h-4" />
        Close Active Project
      </button>
    </div>
  </div>
{/if}
