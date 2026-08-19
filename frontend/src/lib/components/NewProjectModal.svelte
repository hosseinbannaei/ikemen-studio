<script lang="ts">
  import { onMount } from 'svelte';
  import { projectStore } from '../stores/projectStore';
  import { engineStore } from '../stores/engineStore';
  import { toastStore } from '../stores/toastStore';
  import { SelectDirectoryDialog } from '../../../wailsjs/go/main/App';
  import { X, Folder, Plus, Layers, AlertCircle, Loader2 } from 'lucide-svelte';

  export let onClose: () => void;
  export let onOpenEngines: () => void;

  let projectName = '';
  let author = '';
  let targetDir = '';
  let selectedEngine = '';
  let isSubmitting = false;

  onMount(async () => {
    await engineStore.loadInstalled();
    if ($engineStore.installed.length > 0) {
      selectedEngine = $engineStore.installed[0].version;
    }
  });

  async function handleBrowse() {
    try {
      const dir = await SelectDirectoryDialog('Select Project Location');
      if (dir) {
        targetDir = dir;
      }
    } catch (err: any) {
      toastStore.error('Folder Picker Error', err?.message);
    }
  }

  async function handleSubmit() {
    if (!projectName.trim()) {
      toastStore.warning('Project Name Required', 'Please enter a name for your project');
      return;
    }

    if (!targetDir.trim()) {
      toastStore.warning('Location Required', 'Please select a directory for your project');
      return;
    }

    if (!selectedEngine) {
      toastStore.warning('Engine Required', 'Please select an installed Ikemen GO engine version');
      return;
    }

    // Build the final target directory (append project name folder if desired)
    const normalizedDir = targetDir.replace(/\\/g, '/').replace(/\/+$/, '');
    const projectPath = normalizedDir.endsWith(projectName.trim().toLowerCase().replace(/\s+/g, '-'))
      ? normalizedDir
      : `${normalizedDir}/${projectName.trim().toLowerCase().replace(/\s+/g, '-')}`;

    isSubmitting = true;
    const success = await projectStore.create(projectName.trim(), projectPath, selectedEngine, author.trim());
    isSubmitting = false;

    if (success) {
      onClose();
    }
  }
</script>

<div class="fixed inset-0 z-40 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
  <div class="bg-dark-800 border border-dark-600/80 rounded-2xl w-full max-w-lg shadow-2xl overflow-hidden">
    <!-- Header -->
    <div class="p-5 border-b border-dark-600/60 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <div class="w-9 h-9 rounded-xl bg-indigo-500/10 border border-indigo-500/30 flex items-center justify-center text-indigo-400">
          <Plus class="w-5 h-5" />
        </div>
        <div>
          <h2 class="text-base font-bold text-slate-100">Create New Project</h2>
          <p class="text-xs text-slate-400">Scaffold a new Ikemen GO game workspace</p>
        </div>
      </div>
      <button
        type="button"
        class="p-2 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-dark-700 transition"
        on:click={onClose}
      >
        <X class="w-5 h-5" />
      </button>
    </div>

    <!-- Form -->
    <form on:submit|preventDefault={handleSubmit} class="p-5 space-y-4">
      <!-- Project Name -->
      <div>
        <label for="projectName" class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
          Project Name *
        </label>
        <input
          id="projectName"
          type="text"
          bind:value={projectName}
          placeholder="e.g. Marvel vs Capcom Rebirth"
          class="w-full px-3.5 py-2.5 rounded-xl bg-dark-900 border border-dark-600 text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition"
          required
        />
      </div>

      <!-- Author -->
      <div>
        <label for="projectAuthor" class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
          Author / Creator
        </label>
        <input
          id="projectAuthor"
          type="text"
          bind:value={author}
          placeholder="e.g. WrenchWorks"
          class="w-full px-3.5 py-2.5 rounded-xl bg-dark-900 border border-dark-600 text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition"
        />
      </div>

      <!-- Target Directory -->
      <div>
        <label for="projectLocation" class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
          Location *
        </label>
        <div class="flex gap-2">
          <input
            id="projectLocation"
            type="text"
            bind:value={targetDir}
            placeholder="Select parent folder..."
            class="flex-1 px-3.5 py-2.5 rounded-xl bg-dark-900 border border-dark-600 text-xs font-mono text-slate-100 placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition"
            required
          />
          <button
            type="button"
            class="px-3.5 py-2.5 rounded-xl bg-dark-700 hover:bg-dark-600 border border-dark-600 text-slate-200 text-xs font-medium flex items-center gap-1.5 transition"
            on:click={handleBrowse}
          >
            <Folder class="w-4 h-4 text-indigo-400" />
            Browse
          </button>
        </div>
      </div>

      <!-- Engine Version -->
      <div>
        <div class="flex items-center justify-between mb-1.5">
          <label for="engineVersion" class="block text-xs font-semibold text-slate-300 uppercase tracking-wider">
            Engine Version *
          </label>
          <button
            type="button"
            class="text-[11px] text-indigo-400 hover:underline inline-flex items-center gap-1"
            on:click={() => {
              onClose();
              onOpenEngines();
            }}
          >
            <Layers class="w-3 h-3" /> Manage Engines
          </button>
        </div>

        {#if $engineStore.installed.length === 0}
          <div class="p-3.5 rounded-xl bg-amber-950/40 border border-amber-800/50 flex items-start gap-3">
            <AlertCircle class="w-5 h-5 text-amber-400 flex-shrink-0 mt-0.5" />
            <div class="text-xs text-amber-200">
              <span class="font-semibold">No Ikemen GO engine installed.</span>
              <p class="mt-0.5">Please download an engine version first so project assets can be copied.</p>
              <button
                type="button"
                class="mt-2 px-3 py-1 bg-amber-600 hover:bg-amber-500 text-white font-medium rounded-lg transition"
                on:click={() => {
                  onClose();
                  onOpenEngines();
                }}
              >
                Download Engine
              </button>
            </div>
          </div>
        {:else}
          <select
            id="engineVersion"
            bind:value={selectedEngine}
            class="w-full px-3.5 py-2.5 rounded-xl bg-dark-900 border border-dark-600 text-sm text-slate-100 focus:outline-none focus:border-indigo-500 transition"
          >
            {#each $engineStore.installed as eng}
              <option value={eng.version}>
                {eng.version} ({eng.channel})
              </option>
            {/each}
          </select>
        {/if}
      </div>

      <!-- Footer Buttons -->
      <div class="pt-4 border-t border-dark-600/60 flex items-center justify-end gap-3">
        <button
          type="button"
          class="px-4 py-2 rounded-xl text-slate-400 hover:text-slate-200 text-xs font-medium transition"
          on:click={onClose}
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={isSubmitting || $engineStore.installed.length === 0}
          class="px-5 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white text-xs font-semibold shadow-md flex items-center gap-2 transition"
        >
          {#if isSubmitting}
            <Loader2 class="w-3.5 h-3.5 animate-spin" />
            Scaffolding...
          {:else}
            <Plus class="w-3.5 h-3.5" />
            Create Project
          {/if}
        </button>
      </div>
    </form>
  </div>
</div>
