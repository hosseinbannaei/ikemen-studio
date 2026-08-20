<script lang="ts">
  import { engineStore } from '../stores/engineStore';
  import { projectStore } from '../stores/projectStore';
  import type { ExistingGameInspection } from '../types';
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
    ArrowRight,
  } from 'lucide-svelte';

  export let inspection: ExistingGameInspection;
  export let onClose: () => void;
  export let onSuccess: () => void;

  let projectName = inspection.detectedName || 'My Imported Game';
  let targetDir = `${inspection.sourcePath}_studio`;
  let selectedEngine = $engineStore.installed[0]?.version || 'nightly';
  let author = '';
  let importing = false;

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
    const ok = await projectStore.importExisting(
      inspection.sourcePath,
      targetDir.trim(),
      projectName.trim(),
      selectedEngine,
      author.trim()
    );
    importing = false;

    if (ok) {
      onSuccess();
    }
  }
</script>

<div class="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
  <div class="bg-dark-800 border border-dark-600/80 rounded-2xl w-full max-w-xl shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-150 flex flex-col max-h-[90vh]">
    <!-- Header -->
    <div class="p-5 border-b border-dark-600/60 flex items-center justify-between bg-dark-850">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 flex items-center justify-center flex-shrink-0">
          <FolderDown class="w-5 h-5" />
        </div>
        <div>
          <h2 class="text-base font-bold text-slate-100">Import Existing MUGEN / Ikemen Game</h2>
          <p class="text-xs text-slate-400">Adopt legacy folder into a managed Ikemen Studio workspace</p>
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

    <!-- Content -->
    <div class="p-6 space-y-5 overflow-y-auto flex-1">
      <!-- Safe Copy Notice -->
      <div class="p-4 rounded-xl bg-indigo-950/40 border border-indigo-500/30 flex items-start gap-3">
        <ShieldCheck class="w-5 h-5 text-indigo-400 flex-shrink-0 mt-0.5" />
        <div class="space-y-1">
          <div class="text-xs font-bold text-indigo-200">Safe, Non-Destructive Import</div>
          <div class="text-xs text-indigo-300/80 leading-relaxed">
            Studio will copy your characters, stages, and roster into a new dedicated destination. Your original folder will remain <strong>100% untouched</strong>.
          </div>
        </div>
      </div>

      <!-- Discovered Game Stats -->
      <div class="space-y-2">
        <label class="text-[11px] font-bold uppercase tracking-wider text-slate-400">Detected Content</label>
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
      </div>

      <!-- Form Inputs -->
      <div class="space-y-4">
        <div>
          <label class="block text-xs font-semibold text-slate-300 mb-1.5">Project Name</label>
          <input
            type="text"
            bind:value={projectName}
            placeholder="My Fighter Game"
            class="w-full bg-dark-900 border border-dark-600 rounded-xl px-3.5 py-2.5 text-xs text-slate-100 focus:outline-none focus:border-indigo-500"
          />
        </div>

        <div>
          <label class="block text-xs font-semibold text-slate-300 mb-1.5">New Project Location</label>
          <div class="flex gap-2">
            <input
              type="text"
              bind:value={targetDir}
              class="flex-1 bg-dark-900 border border-dark-600 rounded-xl px-3.5 py-2.5 text-xs text-slate-100 font-mono focus:outline-none focus:border-indigo-500"
            />
            <button
              type="button"
              class="px-3.5 py-2.5 bg-dark-700 hover:bg-dark-600 border border-dark-600 rounded-xl text-xs font-semibold text-slate-200 flex items-center gap-1.5 transition"
              on:click={handleBrowseTarget}
            >
              <Folder class="w-4 h-4 text-indigo-400" />
              <span>Browse</span>
            </button>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-xs font-semibold text-slate-300 mb-1.5">Target Engine Version</label>
            <select
              bind:value={selectedEngine}
              class="w-full bg-dark-900 border border-dark-600 rounded-xl px-3.5 py-2.5 text-xs text-slate-200 focus:outline-none focus:border-indigo-500"
            >
              {#each $engineStore.installed as engine}
                <option value={engine.version}>{engine.version} ({engine.channel})</option>
              {/each}
              {#if $engineStore.installed.length === 0}
                <option value="nightly">nightly (will require download)</option>
              {/if}
            </select>
          </div>

          <div>
            <label class="block text-xs font-semibold text-slate-300 mb-1.5">Author (Optional)</label>
            <input
              type="text"
              bind:value={author}
              placeholder="Your Name / Studio"
              class="w-full bg-dark-900 border border-dark-600 rounded-xl px-3.5 py-2.5 text-xs text-slate-100 focus:outline-none focus:border-indigo-500"
            />
          </div>
        </div>
      </div>
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
        disabled={importing || !projectName.trim() || !targetDir.trim()}
        class="px-6 py-2.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 text-white text-xs font-bold shadow-md shadow-emerald-950/50 flex items-center gap-2 transition"
        on:click={handleImport}
      >
        {#if importing}
          <Loader2 class="w-3.5 h-3.5 animate-spin" />
          <span>Copying Assets & Importing...</span>
        {:else}
          <FolderDown class="w-4 h-4" />
          <span>Import & Adopt Project</span>
        {/if}
      </button>
    </div>
  </div>
</div>
