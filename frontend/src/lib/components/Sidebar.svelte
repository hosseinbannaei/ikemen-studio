<script lang="ts">
  import { onMount } from 'svelte';
  import { projectStore } from '../stores/projectStore';
  import { engineStore } from '../stores/engineStore';
  import { settingsStore } from '../stores/settingsStore';
  import {
    FolderKanban,
    Layers,
    Settings,
    Gamepad2,
    Play,
    Square,
    Loader2,
    ChevronRight,
    Sparkles,
    Sun,
    Moon,
    PanelLeftClose,
    PanelLeftOpen,
    LayoutDashboard,
    Grid,
    Mountain,
    Palette,
    Swords,
    Music,
    Type,
    Film,
    Wrench,
    ArrowLeft,
  } from 'lucide-svelte';

  export let activeTab: 'projects' | 'vault' | 'engines' | 'settings';
  export let projectsSubView: string = 'workspace';
  export let onSelectTab: (tab: 'projects' | 'vault' | 'engines' | 'settings') => void;
  export let onOpenProjectWorkspace: () => void;
  export let onNavigateSubView: (subView: string) => void = () => {};
  export let onBackToProjectsList: () => void = () => {};

  let collapsed = false;

  onMount(() => {
    const saved = localStorage.getItem('ikemen_sidebar_collapsed');
    if (saved === 'true') {
      collapsed = true;
    }
  });

  function toggleCollapse() {
    collapsed = !collapsed;
    localStorage.setItem('ikemen_sidebar_collapsed', String(collapsed));
  }

  $: isRunning = $projectStore.gameState === 'running';
  $: isStarting = $projectStore.gameState === 'starting';
  $: isStopping = $projectStore.gameState === 'stopping';

  const studioLinks = [
    { id: 'workspace', label: 'Studio Overview', icon: LayoutDashboard, color: 'text-indigo-400' },
    { id: 'roster', label: 'Roster & Select', icon: Grid, color: 'text-blue-400' },
    { id: 'stages', label: 'Stages Manager', icon: Mountain, color: 'text-emerald-400' },
    { id: 'motifs', label: 'Screenpacks & Motifs', icon: Palette, color: 'text-purple-400' },
    { id: 'lifebars', label: 'Lifebars & HUD', icon: Swords, color: 'text-rose-400' },
    { id: 'sound', label: 'Sound & Music', icon: Music, color: 'text-amber-400' },
    { id: 'fonts', label: 'Fonts & Typography', icon: Type, color: 'text-pink-400' },
    { id: 'storyboards', label: 'Storyboards', icon: Film, color: 'text-orange-400' },
    { id: 'repair', label: 'Repair & Diagnostics', icon: Wrench, color: 'text-cyan-400' },
  ];
</script>

<aside
  class="{collapsed ? 'w-16' : 'w-64'} bg-dark-850 border-r border-dark-600/60 flex flex-col justify-between select-none flex-shrink-0 z-20 transition-all duration-200"
>
  <!-- Brand / Hub Header & Collapse Toggle -->
  <div class="p-3 border-b border-dark-600/40 flex items-center {collapsed ? 'justify-center flex-col gap-2' : 'justify-between'} flex-shrink-0">
    <div class="flex items-center gap-2.5 min-w-0">
      <div class="w-9 h-9 rounded-xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center shadow-lg shadow-indigo-500/20 text-white flex-shrink-0">
        <Gamepad2 class="w-5 h-5" />
      </div>
      {#if !collapsed}
        <div class="min-w-0">
          <div class="text-sm font-bold text-slate-100 truncate">Ikemen Studio</div>
          <div class="text-[10px] font-mono text-slate-400">Hub &bull; Studio Suite</div>
        </div>
      {/if}
    </div>

    <!-- Collapse Button -->
    <button
      type="button"
      class="p-1.5 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-dark-700 transition flex-shrink-0"
      title={collapsed ? 'Expand Sidebar' : 'Collapse Sidebar'}
      on:click={toggleCollapse}
    >
      {#if collapsed}
        <PanelLeftOpen class="w-4 h-4" />
      {:else}
        <PanelLeftClose class="w-4 h-4" />
      {/if}
    </button>
  </div>

  <!-- Navigation Links -->
  <div class="p-2 space-y-3 flex-1 overflow-y-auto custom-scrollbar">
    <!-- Main Hub Links -->
    <div class="space-y-0.5">
      <!-- Projects Hub Tab -->
      <button
        type="button"
        class="w-full flex items-center {collapsed ? 'justify-center p-2.5' : 'justify-between px-3 py-2'} rounded-xl text-xs font-semibold transition-all relative group {
          activeTab === 'projects' && projectsSubView === 'list'
            ? 'bg-brand-600 text-white shadow-md shadow-brand-950/40'
            : 'text-slate-400 hover:text-slate-200 hover:bg-dark-700/60'
        }"
        title={collapsed ? 'Projects Hub' : ''}
        on:click={() => {
          onSelectTab('projects');
          onBackToProjectsList();
        }}
      >
        <div class="flex items-center gap-2.5">
          <FolderKanban class="w-4 h-4 {activeTab === 'projects' && projectsSubView === 'list' ? 'text-white' : 'text-brand-400'}" />
          {#if !collapsed}
            <span>Projects Hub</span>
          {/if}
        </div>
        {#if !collapsed}
          <span class="text-[10px] font-mono opacity-80 px-1.5 py-0.2 rounded bg-black/20">
            {$projectStore.recent.length}
          </span>
        {/if}
      </button>

      <!-- Asset Vault Tab -->
      <button
        type="button"
        class="w-full flex items-center {collapsed ? 'justify-center p-2.5' : 'justify-between px-3 py-2'} rounded-xl text-xs font-semibold transition-all group {
          activeTab === 'vault'
            ? 'bg-brand-600 text-white shadow-md shadow-brand-950/40'
            : 'text-slate-400 hover:text-slate-200 hover:bg-dark-700/60'
        }"
        title={collapsed ? 'Asset Vault' : ''}
        on:click={() => onSelectTab('vault')}
      >
        <div class="flex items-center gap-2.5">
          <Sparkles class="w-4 h-4 {activeTab === 'vault' ? 'text-white' : 'text-brand-400'}" />
          {#if !collapsed}
            <span>Asset Vault</span>
          {/if}
        </div>
      </button>

      <!-- Engines Tab -->
      <button
        type="button"
        class="w-full flex items-center {collapsed ? 'justify-center p-2.5' : 'justify-between px-3 py-2'} rounded-xl text-xs font-semibold transition-all group {
          activeTab === 'engines'
            ? 'bg-brand-600 text-white shadow-md shadow-brand-950/40'
            : 'text-slate-400 hover:text-slate-200 hover:bg-dark-700/60'
        }"
        title={collapsed ? 'Engines' : ''}
        on:click={() => onSelectTab('engines')}
      >
        <div class="flex items-center gap-2.5">
          <Layers class="w-4 h-4 {activeTab === 'engines' ? 'text-white' : 'text-purple-400'}" />
          {#if !collapsed}
            <span>Engines</span>
          {/if}
        </div>
        {#if !collapsed}
          <span class="text-[10px] font-mono px-1.5 py-0.5 rounded-full {
            activeTab === 'engines' ? 'bg-brand-700 text-white' : 'bg-dark-700 text-slate-400'
          }">
            {$engineStore.installed.length}
          </span>
        {/if}
      </button>

      <!-- Settings Tab -->
      <button
        type="button"
        class="w-full flex items-center {collapsed ? 'justify-center p-2.5' : 'justify-between px-3 py-2'} rounded-xl text-xs font-semibold transition-all group {
          activeTab === 'settings'
            ? 'bg-brand-600 text-white shadow-md shadow-brand-950/40'
            : 'text-slate-400 hover:text-slate-200 hover:bg-dark-700/60'
        }"
        title={collapsed ? 'Settings' : ''}
        on:click={() => onSelectTab('settings')}
      >
        <div class="flex items-center gap-2.5">
          <Settings class="w-4 h-4 {activeTab === 'settings' ? 'text-white' : 'text-cyan-400'}" />
          {#if !collapsed}
            <span>Settings</span>
          {/if}
        </div>
      </button>
    </div>

    <!-- Active Studio Project Workbenches (Visible whenever a project is loaded) -->
    {#if $projectStore.current}
      <div class="pt-2 border-t border-dark-700/60 space-y-1">
        {#if !collapsed}
          <div class="px-3 py-1 flex items-center justify-between">
            <span class="text-[10px] uppercase font-black tracking-wider text-slate-400">
              Studio Suite
            </span>
            <span class="text-[10px] font-bold text-emerald-400 font-mono truncate max-w-[90px]">
              {$projectStore.current.name}
            </span>
          </div>
        {/if}

        {#each studioLinks as link}
          {@const isActiveSub = activeTab === 'projects' && projectsSubView === link.id}
          <button
            type="button"
            class="w-full flex items-center {collapsed ? 'justify-center p-2.5' : 'justify-between px-3 py-1.5'} rounded-xl text-xs font-semibold transition-all group {
              isActiveSub
                ? 'bg-indigo-600/30 text-indigo-200 border border-indigo-500/40 shadow-sm'
                : 'text-slate-400 hover:text-slate-200 hover:bg-dark-700/60'
            }"
            title={collapsed ? link.label : ''}
            on:click={() => onNavigateSubView(link.id)}
          >
            <div class="flex items-center gap-2.5 truncate">
              <svelte:component this={link.icon} class="w-4 h-4 flex-shrink-0 {isActiveSub ? 'text-indigo-300' : link.color}" />
              {#if !collapsed}
                <span class="truncate">{link.label}</span>
              {/if}
            </div>
            {#if !collapsed && isActiveSub}
              <span class="w-1.5 h-1.5 rounded-full bg-indigo-400"></span>
            {/if}
          </button>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Bottom: Active Project Widget & Theme Switcher -->
  <div class="p-2 border-t border-dark-600/40 space-y-1.5 bg-dark-900/40 flex-shrink-0">
    {#if $projectStore.current}
      <button
        type="button"
        class="w-full {collapsed ? 'p-2 flex items-center justify-center' : 'p-2.5 flex flex-col gap-1.5 text-left'} rounded-xl bg-dark-800 border transition group shadow-sm {
          activeTab === 'projects' && projectsSubView === 'workspace'
            ? 'border-indigo-500/80 bg-indigo-950/20'
            : 'border-dark-600/60 hover:border-brand-500/50'
        }"
        title={collapsed ? `Current Project: ${$projectStore.current.name}` : ''}
        on:click={onOpenProjectWorkspace}
      >
        {#if !collapsed}
          <div class="flex items-center justify-between w-full">
            <span class="text-[10px] uppercase font-bold tracking-wider text-slate-400">Current Project</span>
            {#if isRunning}
              <span class="flex items-center gap-1 text-[10px] text-emerald-400 font-semibold">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-ping"></span>
                Live
              </span>
            {:else if isStarting}
              <span class="flex items-center gap-1 text-[10px] text-amber-400 font-semibold">
                <Loader2 class="w-2.5 h-2.5 animate-spin" />
                Starting
              </span>
            {/if}
          </div>
          <div class="flex items-center justify-between w-full">
            <span class="text-xs font-bold text-slate-100 truncate group-hover:text-brand-300 transition">
              {$projectStore.current.name}
            </span>
            <ChevronRight class="w-3.5 h-3.5 text-slate-500 group-hover:text-slate-300 transition flex-shrink-0" />
          </div>
          <span class="text-[10px] font-mono text-purple-300/80 truncate">
            {$projectStore.current.engine.version}
          </span>
        {:else}
          <div class="relative">
            <Gamepad2 class="w-5 h-5 text-brand-400" />
            {#if isRunning}
              <span class="absolute -top-1 -right-1 w-2 h-2 rounded-full bg-emerald-400"></span>
            {/if}
          </div>
        {/if}
      </button>
    {/if}

    <!-- Theme Switcher & Version -->
    <div class="p-1 rounded-xl bg-dark-800/80 border border-dark-600/50 flex items-center {collapsed ? 'justify-center' : 'justify-between'} text-xs">
      {#if !collapsed}
        <div class="flex items-center gap-1.5 px-1 font-mono text-[10px] text-slate-400">
          <span>v0.1.0</span>
        </div>
      {/if}

      <button
        type="button"
        class="flex items-center gap-1.5 {collapsed ? 'p-1.5' : 'px-2.5 py-1'} rounded-lg bg-dark-700 hover:bg-dark-600 border border-dark-600/70 text-slate-200 text-xs font-medium transition shadow-sm"
        on:click={() => settingsStore.toggleTheme()}
        title="Toggle {$settingsStore.theme === 'light' ? 'Dark' : 'Light'} Mode"
      >
        {#if $settingsStore.theme === 'light'}
          <Sun class="w-3.5 h-3.5 text-amber-500" />
          {#if !collapsed}<span class="text-[11px] font-semibold text-slate-700">Light</span>{/if}
        {:else}
          <Moon class="w-3.5 h-3.5 text-brand-400" />
          {#if !collapsed}<span class="text-[11px] font-semibold text-slate-300">Dark</span>{/if}
        {/if}
      </button>
    </div>
  </div>
</aside>
