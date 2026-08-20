<script lang="ts">
  import { projectStore } from '../stores/projectStore';
  import {
    AlertTriangle,
    Wrench,
    FolderOpen,
    X,
    CheckCircle2,
    Play,
    Loader2,
    FileText,
    Sliders,
    Sparkles,
  } from 'lucide-svelte';

  export let onOpenRepairHub: () => void;

  let fixingConfig = false;
  let configFixed = false;

  async function handleQuickFixConfig() {
    fixingConfig = true;
    const ok = await projectStore.repairConfig();
    fixingConfig = false;
    if (ok) {
      configFixed = true;
    }
  }

  function handleOpenHub() {
    projectStore.dismissCrash();
    onOpenRepairHub();
  }

  function handleRelaunch() {
    projectStore.dismissCrash();
    projectStore.launch();
  }
</script>

{#if $projectStore.activeCrash}
  <div class="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
    <div class="bg-dark-800 border border-amber-500/50 rounded-2xl w-full max-w-xl shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-200">
      <!-- Header -->
      <div class="p-5 border-b border-dark-600/60 bg-amber-500/10 flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-amber-500/20 text-amber-400 border border-amber-500/30 flex items-center justify-center flex-shrink-0 shadow-sm">
            <AlertTriangle class="w-5 h-5" />
          </div>
          <div>
            <h2 class="text-base font-bold text-slate-100">Game Crash / Abnormal Exit Detected</h2>
            <p class="text-xs text-amber-300/80">Ikemen GO encountered an unexpected runtime error</p>
          </div>
        </div>
        <button
          type="button"
          class="p-2 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-dark-700/60 transition"
          on:click={() => projectStore.dismissCrash()}
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Content -->
      <div class="p-6 space-y-5">
        <!-- Error Output Box -->
        {#if $projectStore.activeCrash.errorSummary}
          <div class="space-y-1.5">
            <div class="text-[11px] font-bold uppercase tracking-wider text-slate-400 flex items-center gap-1.5">
              <FileText class="w-3.5 h-3.5 text-rose-400" />
              <span>Crash Diagnostic Summary</span>
            </div>
            <div class="p-3.5 rounded-xl bg-dark-900 border border-dark-600/60 text-xs font-mono text-rose-300 whitespace-pre-wrap leading-relaxed max-h-36 overflow-y-auto shadow-inner">
              {$projectStore.activeCrash.errorSummary}
            </div>
          </div>
        {/if}

        <!-- Diagnostic Hint & Action Options -->
        <div class="p-4 rounded-xl bg-dark-900/60 border border-dark-600/60 space-y-3">
          <div class="flex items-start gap-2.5">
            <Wrench class="w-4 h-4 text-indigo-400 mt-0.5 flex-shrink-0" />
            <div class="text-xs text-slate-300 leading-relaxed">
              <strong class="text-slate-100 font-semibold">Recommended Troubleshooting:</strong> Use the
              <span class="text-indigo-300 font-bold">Maintenance & Repair Hub</span> to fix invalid render modes, update legacy ZSS syntax in stock characters (e.g. <code class="text-purple-300">kfm_zss</code>), or reset configuration.
            </div>
          </div>

          {#if configFixed}
            <div class="p-3 rounded-lg bg-emerald-950/40 border border-emerald-500/30 flex items-center gap-2 text-xs text-emerald-300 font-medium">
              <CheckCircle2 class="w-4 h-4 text-emerald-400" />
              <span>Configuration repaired! Legacy RenderMode and display keys normalized.</span>
            </div>
          {/if}
        </div>

        <!-- Logs Directory Note -->
        <div class="text-[11px] text-slate-400 flex items-center justify-between">
          <span class="truncate max-w-sm font-mono text-[10px]">Log: save/logs/ikemen-latest.log</span>
          <button
            type="button"
            class="text-indigo-400 hover:text-indigo-300 font-semibold inline-flex items-center gap-1 transition"
            on:click={() => projectStore.openLogs()}
          >
            <FolderOpen class="w-3.5 h-3.5" />
            <span>Open Logs Folder</span>
          </button>
        </div>
      </div>

      <!-- Footer Buttons -->
      <div class="p-5 border-t border-dark-600/60 bg-dark-850 flex items-center justify-between gap-3">
        <button
          type="button"
          class="px-4 py-2 rounded-xl text-slate-400 hover:text-slate-200 text-xs font-medium transition"
          on:click={() => projectStore.dismissCrash()}
        >
          Dismiss
        </button>

        <div class="flex items-center gap-2.5">
          <button
            type="button"
            disabled={fixingConfig}
            class="flex items-center gap-1.5 px-3.5 py-2.5 rounded-xl bg-amber-600/90 hover:bg-amber-500 text-white text-xs font-semibold shadow-sm transition"
            on:click={handleQuickFixConfig}
            title="Auto-fix RenderMode and invalid config parameters"
          >
            {#if fixingConfig}
              <Loader2 class="w-3.5 h-3.5 animate-spin" />
              <span>Fixing...</span>
            {:else}
              <Sliders class="w-3.5 h-3.5" />
              <span>Fix Config</span>
            {/if}
          </button>

          <button
            type="button"
            class="flex items-center gap-2 px-5 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold shadow-md shadow-indigo-600/30 transition"
            on:click={handleOpenHub}
          >
            <Wrench class="w-3.5 h-3.5" />
            <span>Open Repair Hub</span>
          </button>

          <button
            type="button"
            class="flex items-center gap-2 px-5 py-2.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-bold shadow-md shadow-emerald-600/30 transition"
            on:click={handleRelaunch}
          >
            <Play class="w-3.5 h-3.5 fill-current" />
            <span>Re-Launch</span>
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
