<script lang="ts">
  import { onMount } from 'svelte';
  import { projectStore } from '../stores/projectStore';
  import {
    Sliders,
    Monitor,
    Volume2,
    Cpu,
    X,
    Check,
    Loader2,
  } from 'lucide-svelte';

  export let onClose: () => void;

  let loading = true;
  let saving = false;

  let width = '1280';
  let height = '720';
  let fullscreen = '0';
  let volumeMaster = '80';
  let volumeBgm = '80';
  let volumeSfx = '80';
  let vsync = '1';
  let renderer = '0';

  const resPresets = [
    { label: '720p HD (1280x720 - 16:9)', w: '1280', h: '720' },
    { label: '1080p Full HD (1920x1080 - 16:9)', w: '1920', h: '1080' },
    { label: '1440p 2K (2560x1440 - 16:9)', w: '2560', h: '1440' },
    { label: '4K Ultra HD (3840x2160 - 16:9)', w: '3840', h: '2160' },
    { label: 'Classic MUGEN (640x480 - 4:3)', w: '640', h: '480' },
    { label: 'Classic Hi-Res (1024x768 - 4:3)', w: '1024', h: '768' },
  ];

  let selectedPreset = '1280x720';

  onMount(async () => {
    loading = true;
    const cfg = await projectStore.getGameConfig();
    if (cfg) {
      width = cfg['Width'] || cfg['width'] || '1280';
      height = cfg['Height'] || cfg['height'] || '720';
      fullscreen = cfg['Fullscreen'] || cfg['fullscreen'] || '0';
      volumeMaster = cfg['VolumeMaster'] || cfg['volumemaster'] || '80';
      volumeBgm = cfg['VolumeBgm'] || cfg['volumebgm'] || '80';
      volumeSfx = cfg['VolumeSfx'] || cfg['volumesfx'] || '80';
      vsync = cfg['Vsync'] || cfg['vsync'] || '1';
      renderer = cfg['Renderer'] || cfg['renderer'] || '0';

      selectedPreset = `${width}x${height}`;
    }
    loading = false;
  });

  function handlePresetChange(e: Event) {
    const val = (e.target as HTMLSelectElement).value;
    selectedPreset = val;
    const preset = resPresets.find((p) => `${p.w}x${p.h}` === val);
    if (preset) {
      width = preset.w;
      height = preset.h;
    }
  }

  async function handleSave() {
    saving = true;
    const updates: Record<string, string> = {
      Width: width,
      Height: height,
      Fullscreen: fullscreen,
      VolumeMaster: volumeMaster,
      VolumeBgm: volumeBgm,
      VolumeSfx: volumeSfx,
      Vsync: vsync,
      Renderer: renderer,
    };

    await projectStore.saveGameConfig(updates);
    saving = false;
    onClose();
  }
</script>

<div class="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
  <div class="bg-dark-800 border border-dark-600/80 rounded-2xl w-full max-w-lg shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-150 flex flex-col max-h-[90vh]">
    <!-- Header -->
    <div class="p-5 border-b border-dark-600/60 flex items-center justify-between bg-dark-850">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-cyan-500/10 border border-cyan-500/20 text-cyan-400 flex items-center justify-center flex-shrink-0">
          <Sliders class="w-5 h-5" />
        </div>
        <div>
          <h2 class="text-base font-bold text-slate-100">Game Preferences & Config</h2>
          <p class="text-xs text-slate-400">Configure resolution, graphics, and audio for save/config.ini</p>
        </div>
      </div>
      <button
        type="button"
        class="p-2 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-dark-700/60 transition"
        on:click={onClose}
      >
        <X class="w-5 h-5" />
      </button>
    </div>

    <!-- Body -->
    <div class="p-6 space-y-6 overflow-y-auto flex-1">
      {#if loading}
        <div class="py-12 flex flex-col items-center justify-center gap-2 text-slate-400">
          <Loader2 class="w-6 h-6 animate-spin text-indigo-400" />
          <span class="text-xs">Loading game preferences...</span>
        </div>
      {:else}
        <!-- Display & Resolution -->
        <div class="space-y-3">
          <div class="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-300">
            <Monitor class="w-4 h-4 text-indigo-400" />
            <span>Display & Resolution</span>
          </div>

          <div class="space-y-2">
            <label class="text-[11px] font-semibold text-slate-400">Resolution Preset</label>
            <select
              value={selectedPreset}
              on:change={handlePresetChange}
              class="w-full bg-dark-900 border border-dark-600 rounded-xl px-3.5 py-2.5 text-xs text-slate-200 focus:outline-none focus:border-indigo-500"
            >
              {#each resPresets as p}
                <option value="{p.w}x{p.h}">{p.label}</option>
              {/each}
              <option value="custom">Custom Resolution</option>
            </select>
          </div>

          {#if selectedPreset === 'custom'}
            <div class="grid grid-cols-2 gap-3 pt-1">
              <div>
                <label class="text-[11px] text-slate-400">Width</label>
                <input
                  type="number"
                  bind:value={width}
                  class="w-full bg-dark-900 border border-dark-600 rounded-xl px-3 py-2 text-xs text-slate-100 font-mono"
                />
              </div>
              <div>
                <label class="text-[11px] text-slate-400">Height</label>
                <input
                  type="number"
                  bind:value={height}
                  class="w-full bg-dark-900 border border-dark-600 rounded-xl px-3 py-2 text-xs text-slate-100 font-mono"
                />
              </div>
            </div>
          {/if}

          <!-- Display Mode & VSync -->
          <div class="grid grid-cols-2 gap-3 pt-1">
            <label class="flex items-center gap-2.5 p-3 rounded-xl bg-dark-900 border border-dark-600/60 cursor-pointer">
              <input
                type="checkbox"
                checked={fullscreen === '1'}
                on:change={(e) => (fullscreen = e.currentTarget.checked ? '1' : '0')}
                class="rounded border-dark-600 bg-dark-800 text-indigo-600 focus:ring-0 w-4 h-4"
              />
              <span class="text-xs text-slate-200 font-semibold">Fullscreen Mode</span>
            </label>

            <label class="flex items-center gap-2.5 p-3 rounded-xl bg-dark-900 border border-dark-600/60 cursor-pointer">
              <input
                type="checkbox"
                checked={vsync === '1'}
                on:change={(e) => (vsync = e.currentTarget.checked ? '1' : '0')}
                class="rounded border-dark-600 bg-dark-800 text-indigo-600 focus:ring-0 w-4 h-4"
              />
              <span class="text-xs text-slate-200 font-semibold">V-Sync (60 FPS)</span>
            </label>
          </div>
        </div>

        <!-- Audio Volumes -->
        <div class="space-y-3 pt-2 border-t border-dark-600/40">
          <div class="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-300">
            <Volume2 class="w-4 h-4 text-emerald-400" />
            <span>Audio Levels</span>
          </div>

          <div class="space-y-3">
            <div>
              <div class="flex items-center justify-between text-xs text-slate-400 mb-1">
                <span>Master Volume</span>
                <span class="font-mono text-slate-200 font-bold">{volumeMaster}%</span>
              </div>
              <input
                type="range"
                min="0"
                max="100"
                bind:value={volumeMaster}
                class="w-full h-1.5 bg-dark-900 rounded-lg appearance-none cursor-pointer accent-indigo-500"
              />
            </div>

            <div>
              <div class="flex items-center justify-between text-xs text-slate-400 mb-1">
                <span>BGM (Music) Volume</span>
                <span class="font-mono text-slate-200 font-bold">{volumeBgm}%</span>
              </div>
              <input
                type="range"
                min="0"
                max="100"
                bind:value={volumeBgm}
                class="w-full h-1.5 bg-dark-900 rounded-lg appearance-none cursor-pointer accent-emerald-500"
              />
            </div>

            <div>
              <div class="flex items-center justify-between text-xs text-slate-400 mb-1">
                <span>SFX (Sound Effects) Volume</span>
                <span class="font-mono text-slate-200 font-bold">{volumeSfx}%</span>
              </div>
              <input
                type="range"
                min="0"
                max="100"
                bind:value={volumeSfx}
                class="w-full h-1.5 bg-dark-900 rounded-lg appearance-none cursor-pointer accent-amber-500"
              />
            </div>
          </div>
        </div>

        <!-- Renderer -->
        <div class="space-y-2 pt-2 border-t border-dark-600/40">
          <div class="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-300">
            <Cpu class="w-4 h-4 text-purple-400" />
            <span>Graphics Renderer</span>
          </div>
          <select
            bind:value={renderer}
            class="w-full bg-dark-900 border border-dark-600 rounded-xl px-3.5 py-2.5 text-xs text-slate-200 focus:outline-none focus:border-indigo-500"
          >
            <option value="0">OpenGL 3.3 Core (Default Hardware Accelerated)</option>
            <option value="1">Vulkan 1.x (High Performance / Next-Gen)</option>
            <option value="2">DirectX / Software (Compatibility Fallback)</option>
          </select>
        </div>
      {/if}
    </div>

    <!-- Footer -->
    <div class="p-4 border-t border-dark-600/60 bg-dark-850 flex items-center justify-between">
      <button
        type="button"
        class="px-4 py-2 rounded-xl text-slate-400 hover:text-slate-200 text-xs font-semibold hover:bg-dark-700/60 transition"
        on:click={onClose}
      >
        Cancel
      </button>

      <button
        type="button"
        disabled={saving || loading}
        class="px-6 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white text-xs font-bold shadow-md shadow-indigo-950/50 flex items-center gap-2 transition"
        on:click={handleSave}
      >
        {#if saving}
          <Loader2 class="w-3.5 h-3.5 animate-spin" />
          <span>Saving...</span>
        {:else}
          <Check class="w-3.5 h-3.5" />
          <span>Save Game Config</span>
        {/if}
      </button>
    </div>
  </div>
</div>
