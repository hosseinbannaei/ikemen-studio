<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { projectStore } from '../stores/projectStore';
  import { vaultStore } from '../stores/vaultStore';
  import { toastStore } from '../stores/toastStore';
  import type { ProjectAudioInfo } from '../types';
  import AddFromVaultModal from './AddFromVaultModal.svelte';
  import ConfirmModal from './ConfirmModal.svelte';
  import {
    Music,
    ArrowLeft,
    Sparkles,
    Plus,
    FolderOpen,
    Play,
    Pause,
    Volume2,
    VolumeX,
    Trash2,
    Check,
    Loader2,
    Search,
    ListMusic,
    Disc,
    Tag,
  } from 'lucide-svelte';

  export let onBackToWorkspace: () => void;

  let tracks: ProjectAudioInfo[] = [];
  let loading = true;
  let searchQuery = '';

  // Audio Playback State
  let audioElement: HTMLAudioElement | null = null;
  let currentlyPlayingPath = '';
  let isPlaying = false;
  let currentTime = 0;
  let duration = 0;
  let volume = 0.8;
  let isMuted = false;

  let showAddFromVaultModal = false;
  let trackToDelete: ProjectAudioInfo | null = null;

  onMount(async () => {
    await loadAudioTracks();
    audioElement = new Audio();
    audioElement.volume = volume;

    audioElement.ontimeupdate = () => {
      if (audioElement) currentTime = audioElement.currentTime;
    };
    audioElement.onloadedmetadata = () => {
      if (audioElement) duration = audioElement.duration;
    };
    audioElement.onended = () => {
      isPlaying = false;
      currentTime = 0;
    };
  });

  onDestroy(() => {
    if (audioElement) {
      audioElement.pause();
      audioElement.src = '';
      audioElement = null;
    }
  });

  async function loadAudioTracks() {
    loading = true;
    tracks = await projectStore.loadAudio();
    loading = false;
  }

  $: filteredTracks = tracks.filter((t) => {
    const q = searchQuery.toLowerCase();
    return (
      t.file_name.toLowerCase().includes(q) ||
      t.relative_path.toLowerCase().includes(q) ||
      t.assigned_events.some((e) => e.toLowerCase().includes(q)) ||
      t.assigned_stages.some((s) => s.toLowerCase().includes(q))
    );
  });

  function togglePlayTrack(track: ProjectAudioInfo) {
    if (!audioElement || !$projectStore.current?.path) return;

    if (currentlyPlayingPath === track.relative_path && isPlaying) {
      audioElement.pause();
      isPlaying = false;
    } else {
      const fullPath = `${$projectStore.current.path}/${track.relative_path}`;
      // In Wails / WebKit, load local file or asset handler
      audioElement.src = `https://wails.localhost/asset?path=${encodeURIComponent(fullPath)}`;
      audioElement.play().catch(() => {
        // Fallback or direct local file URI
        if (audioElement) {
          audioElement.src = fullPath;
          audioElement.play().catch(() => {});
        }
      });
      currentlyPlayingPath = track.relative_path;
      isPlaying = true;
    }
  }

  function handleSeek(e: Event) {
    const val = parseFloat((e.target as HTMLInputElement).value);
    if (audioElement && !isNaN(val)) {
      audioElement.currentTime = val;
      currentTime = val;
    }
  }

  function handleVolume(e: Event) {
    const val = parseFloat((e.target as HTMLInputElement).value);
    volume = val;
    if (audioElement) {
      audioElement.volume = val;
      isMuted = val === 0;
    }
  }

  function toggleMute() {
    if (!audioElement) return;
    isMuted = !isMuted;
    audioElement.muted = isMuted;
  }

  async function handleAssignBGM(eventType: 'title' | 'select' | 'vs' | 'victory', track: ProjectAudioInfo) {
    const ok = await projectStore.setSystemBGM(eventType, track.relative_path);
    if (ok) {
      await loadAudioTracks();
    }
  }

  async function confirmDeleteTrack() {
    if (!trackToDelete) return;
    const target = trackToDelete;
    trackToDelete = null;
    const ok = await projectStore.deleteAudio(target.relative_path);
    if (ok) {
      if (currentlyPlayingPath === target.relative_path && audioElement) {
        audioElement.pause();
        isPlaying = false;
        currentlyPlayingPath = '';
      }
      await loadAudioTracks();
    }
  }

  async function handleBrowseAudio() {
    try {
      const filePath = await projectStore.selectAudioFileDialog();
      if (filePath) {
        // Ingest into vault and project
        await vaultStore.ingestAsset(filePath, 'sounds');
        await loadAudioTracks();
        toastStore.success('Audio Ingested', filePath);
      }
    } catch (err: any) {
      toastStore.error('Import Failed', err?.message || 'Could not import audio track');
    }
  }

  function formatTime(secs: number): string {
    if (isNaN(secs) || secs < 0) return '0:00';
    const m = Math.floor(secs / 60);
    const s = Math.floor(secs % 60);
    return `${m}:${s < 10 ? '0' : ''}${s}`;
  }

  function formatBytes(b: number): string {
    if (b < 1024) return `${b} B`;
    if (b < 1024 * 1024) return `${(b / 1024).toFixed(1)} KB`;
    return `${(b / (1024 * 1024)).toFixed(1)} MB`;
  }
</script>

<div class="h-full flex flex-col min-w-0 bg-dark-900 select-none overflow-hidden">
  <!-- Top Bar -->
  <header class="p-6 border-b border-dark-600/60 bg-dark-850/60 flex flex-col md:flex-row md:items-center justify-between gap-4 flex-shrink-0">
    <div class="flex items-center gap-3">
      <button
        type="button"
        class="p-2 rounded-xl bg-dark-700 hover:bg-dark-600 text-slate-300 hover:text-white transition shadow-sm"
        on:click={onBackToWorkspace}
        title="Back to Workspace"
      >
        <ArrowLeft class="w-4 h-4" />
      </button>

      <div>
        <div class="flex items-center gap-2">
          <h1 class="text-xl font-black text-slate-100 flex items-center gap-2">
            <Music class="w-5 h-5 text-rose-400" />
            <span>Sound & Music Library</span>
          </h1>
          <span class="text-xs px-2 py-0.5 rounded-full bg-rose-500/10 text-rose-300 border border-rose-500/20 font-mono">
            {tracks.length} Tracks
          </span>
        </div>
        <p class="text-xs text-slate-400 mt-0.5">Manage BGM soundtracks, preview audio, and map title, select, and VS screen music (system.def)</p>
      </div>
    </div>

    <!-- Actions -->
    <div class="flex items-center gap-2.5 flex-wrap">
      <button
        type="button"
        class="px-3.5 py-2 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600/70 text-slate-200 text-xs font-semibold flex items-center gap-2 transition shadow-sm"
        on:click={() => projectStore.openFolder('sound')}
        title="Open sound/ folder"
      >
        <FolderOpen class="w-4 h-4 text-indigo-400" />
        <span>Open sound/</span>
      </button>

      <button
        type="button"
        class="px-3.5 py-2 rounded-xl bg-brand-600/20 hover:bg-brand-600/30 border border-brand-500/40 text-brand-300 text-xs font-semibold flex items-center gap-2 transition shadow-sm"
        on:click={() => (showAddFromVaultModal = true)}
        title="Link audio from Asset Vault"
      >
        <Sparkles class="w-4 h-4 text-brand-400" />
        <span>+ From Vault</span>
      </button>

      <button
        type="button"
        class="px-4 py-2 rounded-xl bg-rose-600 hover:bg-rose-500 text-white text-xs font-bold flex items-center gap-2 transition shadow-md shadow-rose-950/40"
        on:click={handleBrowseAudio}
        title="Import audio track (.mp3, .ogg, .wav)"
      >
        <Plus class="w-4 h-4" />
        <span>Add Audio Track</span>
      </button>
    </div>
  </header>

  <!-- Global Audio Player Header Deck -->
  <div class="p-4 px-6 border-b border-dark-600/50 bg-dark-850/90 flex flex-col md:flex-row md:items-center justify-between gap-4 flex-shrink-0 shadow-inner">
    <div class="flex items-center gap-3 min-w-0 flex-1">
      <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-rose-500/20 to-purple-500/20 text-rose-400 border border-rose-500/30 flex items-center justify-center flex-shrink-0">
        <Disc class="w-5 h-5 {isPlaying ? 'animate-spin' : ''}" />
      </div>

      <div class="min-w-0 flex-1">
        <div class="text-xs font-bold text-slate-100 truncate">
          {currentlyPlayingPath ? currentlyPlayingPath : 'No Track Selected'}
        </div>
        <div class="flex items-center gap-2 text-[11px] font-mono text-slate-400 mt-0.5">
          <span>{formatTime(currentTime)} / {formatTime(duration)}</span>
        </div>
      </div>
    </div>

    <!-- Player Controls & Seekbar -->
    <div class="flex items-center gap-4 flex-1 max-w-xl">
      <input
        type="range"
        min="0"
        max={duration || 100}
        value={currentTime}
        on:input={handleSeek}
        class="w-full h-1.5 bg-dark-700 rounded-lg appearance-none cursor-pointer accent-rose-500"
      />

      <!-- Volume -->
      <div class="flex items-center gap-2 text-slate-400 flex-shrink-0">
        <button type="button" on:click={toggleMute} class="hover:text-slate-200">
          {#if isMuted || volume === 0}
            <VolumeX class="w-4 h-4 text-rose-400" />
          {:else}
            <Volume2 class="w-4 h-4" />
          {/if}
        </button>
        <input
          type="range"
          min="0"
          max="1"
          step="0.05"
          value={isMuted ? 0 : volume}
          on:input={handleVolume}
          class="w-20 h-1 bg-dark-700 rounded-lg appearance-none cursor-pointer accent-rose-500"
        />
      </div>
    </div>
  </div>

  <!-- Search Bar -->
  <div class="p-4 px-6 border-b border-dark-600/40 bg-dark-900 flex items-center justify-between gap-4 flex-shrink-0">
    <div class="relative flex-1 max-w-md">
      <Search class="w-4 h-4 text-slate-500 absolute left-3 top-1/2 -translate-y-1/2" />
      <input
        type="text"
        bind:value={searchQuery}
        placeholder="Search audio tracks or events..."
        class="w-full bg-dark-800 border border-dark-600/70 rounded-xl pl-9 pr-3 py-1.5 text-xs text-slate-200 placeholder:text-slate-500 focus:outline-none focus:border-rose-500 transition"
      />
    </div>
  </div>

  <!-- Main Content Area -->
  <main class="flex-1 overflow-y-auto p-6">
    {#if loading}
      <div class="h-full flex flex-col items-center justify-center gap-3 text-slate-400">
        <Loader2 class="w-8 h-8 animate-spin text-rose-400" />
        <span class="text-xs">Scanning sound tracks and music definitions...</span>
      </div>
    {:else if filteredTracks.length === 0}
      <div class="h-full flex flex-col items-center justify-center p-12 text-center text-slate-500 border-2 border-dashed border-dark-700/60 rounded-3xl">
        <Music class="w-16 h-16 stroke-1 opacity-40 mb-3 text-rose-400" />
        <h3 class="text-base font-bold text-slate-300">No Audio Tracks Found</h3>
        <p class="text-xs text-slate-500 max-w-sm mt-1.5">
          Add your background music tracks (.mp3, .ogg, .wav) to sound/ or link them from Vault.
        </p>
      </div>
    {:else}
      <div class="space-y-2">
        {#each filteredTracks as track (track.relative_path)}
          {@const isThisPlaying = currentlyPlayingPath === track.relative_path && isPlaying}
          <div
            class="p-4 rounded-xl bg-dark-800 border transition flex flex-col sm:flex-row sm:items-center justify-between gap-4 group {isThisPlaying ? 'border-rose-500/80 bg-dark-800/90 shadow-md shadow-rose-950/20' : 'border-dark-600/70 hover:border-dark-600'}"
          >
            <!-- Left Info -->
            <div class="flex items-center gap-3.5 min-w-0 flex-1">
              <button
                type="button"
                class="w-9 h-9 rounded-xl flex items-center justify-center transition flex-shrink-0 {isThisPlaying ? 'bg-rose-500 text-white shadow-md shadow-rose-950/50' : 'bg-dark-700 group-hover:bg-rose-600 text-slate-300 group-hover:text-white'}"
                on:click={() => togglePlayTrack(track)}
                title={isThisPlaying ? 'Pause' : 'Play Track'}
              >
                {#if isThisPlaying}
                  <Pause class="w-4 h-4 fill-current" />
                {:else}
                  <Play class="w-4 h-4 fill-current ml-0.5" />
                {/if}
              </button>

              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2 flex-wrap">
                  <span class="text-xs font-bold text-slate-100 group-hover:text-rose-300 transition truncate">
                    {track.file_name}
                  </span>
                  <span class="text-[10px] font-mono uppercase px-1.5 py-0.2 bg-dark-900 text-slate-400 rounded border border-dark-700">
                    {track.format}
                  </span>
                  <span class="text-[10px] font-mono text-slate-500">
                    {formatBytes(track.size_bytes)}
                  </span>
                </div>

                <!-- Assigned Tags -->
                <div class="flex items-center gap-1.5 flex-wrap mt-1">
                  {#each track.assigned_events as event}
                    <span class="text-[10px] font-bold px-2 py-0.5 rounded-full bg-rose-500/20 text-rose-300 border border-rose-500/30 flex items-center gap-1">
                      <Tag class="w-2.5 h-2.5" />
                      {event}
                    </span>
                  {/each}

                  {#each track.assigned_stages as stage}
                    <span class="text-[10px] font-medium px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-300 border border-emerald-500/20">
                      Stage: {stage}
                    </span>
                  {/each}
                </div>
              </div>
            </div>

            <!-- Right Quick Assignment Controls -->
            <div class="flex items-center gap-2 flex-wrap">
              <div class="flex items-center bg-dark-900 p-1 rounded-xl border border-dark-700 text-xs">
                <button
                  type="button"
                  class="px-2.5 py-1 text-[11px] font-semibold text-slate-300 hover:text-white hover:bg-dark-750 rounded-lg transition"
                  on:click={() => handleAssignBGM('title', track)}
                  title="Assign as Title Screen BGM"
                >
                  Set Title
                </button>
                <button
                  type="button"
                  class="px-2.5 py-1 text-[11px] font-semibold text-slate-300 hover:text-white hover:bg-dark-750 rounded-lg transition"
                  on:click={() => handleAssignBGM('select', track)}
                  title="Assign as Character Select BGM"
                >
                  Set Select
                </button>
                <button
                  type="button"
                  class="px-2.5 py-1 text-[11px] font-semibold text-slate-300 hover:text-white hover:bg-dark-750 rounded-lg transition"
                  on:click={() => handleAssignBGM('vs', track)}
                  title="Assign as VS Screen BGM"
                >
                  Set VS
                </button>
                <button
                  type="button"
                  class="px-2.5 py-1 text-[11px] font-semibold text-slate-300 hover:text-white hover:bg-dark-750 rounded-lg transition"
                  on:click={() => handleAssignBGM('victory', track)}
                  title="Assign as Victory Screen BGM"
                >
                  Set Victory
                </button>
              </div>

              <button
                type="button"
                class="p-2 text-slate-500 hover:text-rose-400 hover:bg-rose-500/10 rounded-xl transition"
                on:click={() => (trackToDelete = track)}
                title="Delete Audio Track"
              >
                <Trash2 class="w-4 h-4" />
              </button>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </main>
</div>

<!-- Confirm Delete Modal -->
{#if trackToDelete}
  <ConfirmModal
    title="Delete Audio File?"
    message="Are you sure you want to remove '{trackToDelete.file_name}' from the sound/ folder?"
    confirmLabel="Delete Track"
    confirmVariant="danger"
    onConfirm={confirmDeleteTrack}
    onCancel={() => (trackToDelete = null)}
  />
{/if}

<!-- Add From Vault Modal -->
{#if showAddFromVaultModal}
  <AddFromVaultModal
    isOpen={showAddFromVaultModal}
    projectDir={$projectStore.current?.path || ''}
    targetCategory="sounds"
    onClose={() => {
      showAddFromVaultModal = false;
      loadAudioTracks();
    }}
  />
{/if}
