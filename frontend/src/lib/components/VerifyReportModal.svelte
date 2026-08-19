<script lang="ts">
  import { projectStore } from '../stores/projectStore';
  import type { VerificationReport } from '../types';
  import {
    ShieldCheck,
    Wrench,
    CheckCircle2,
    AlertCircle,
    FolderOpen,
    X,
    FileText,
    Loader2,
  } from 'lucide-svelte';

  export let report: VerificationReport | null = null;
  export let isLoading = false;
  export let onClose: () => void;
</script>

<div class="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
  <div class="bg-dark-800 border border-dark-600/80 rounded-2xl w-full max-w-lg shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-200 flex flex-col">
    <!-- Header -->
    <div class="p-5 border-b border-dark-600/60 flex items-center justify-between bg-dark-850">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-indigo-500/10 border border-indigo-500/30 flex items-center justify-center text-indigo-400 flex-shrink-0 shadow-sm">
          <ShieldCheck class="w-5 h-5" />
        </div>
        <div>
          <h2 class="text-base font-bold text-slate-100">Game Files Verification & Repair</h2>
          <p class="text-xs text-slate-400">Verifies project assets against engine repository</p>
        </div>
      </div>
      {#if !isLoading}
        <button
          type="button"
          class="p-2 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-dark-700/60 transition"
          on:click={onClose}
        >
          <X class="w-5 h-5" />
        </button>
      {/if}
    </div>

    <!-- Content Area -->
    <div class="p-6 space-y-5 flex-1 overflow-y-auto">
      {#if isLoading}
        <div class="py-12 text-center flex flex-col items-center justify-center gap-3">
          <Loader2 class="w-8 h-8 animate-spin text-indigo-400" />
          <div class="space-y-1">
            <div class="text-sm font-bold text-slate-200">Scanning Project Files...</div>
            <div class="text-xs text-slate-400">Comparing scripts, DLLs, shaders, and system folders against engine cache.</div>
          </div>
        </div>
      {:else if report}
        <!-- Result Status Banner -->
        {#if report.repairedCount > 0}
          <div class="p-4 rounded-xl bg-emerald-950/40 border border-emerald-500/50 flex items-start gap-3">
            <CheckCircle2 class="w-5 h-5 text-emerald-400 flex-shrink-0 mt-0.5" />
            <div class="space-y-0.5">
              <div class="text-sm font-bold text-emerald-200">Verification & Repair Complete</div>
              <div class="text-xs text-emerald-300/90">
                Successfully restored <strong class="text-white">{report.repairedCount} missing file(s)</strong> to your project.
              </div>
            </div>
          </div>
        {:else if report.success}
          <div class="p-4 rounded-xl bg-indigo-950/40 border border-indigo-500/40 flex items-start gap-3">
            <CheckCircle2 class="w-5 h-5 text-indigo-400 flex-shrink-0 mt-0.5" />
            <div class="space-y-0.5">
              <div class="text-sm font-bold text-indigo-200">All Files Validated (100% Intact)</div>
              <div class="text-xs text-indigo-300/80">
                All core engine scripts, libraries, shaders, and system folders are present and ready.
              </div>
            </div>
          </div>
        {:else}
          <div class="p-4 rounded-xl bg-rose-950/40 border border-rose-500/50 flex items-start gap-3">
            <AlertCircle class="w-5 h-5 text-rose-400 flex-shrink-0 mt-0.5" />
            <div class="space-y-0.5">
              <div class="text-sm font-bold text-rose-200">Verification Error</div>
              <div class="text-xs text-rose-300/90">{report.errorMessage || 'Could not complete file verification'}</div>
            </div>
          </div>
        {/if}

        <!-- Statistics Grid -->
        <div class="grid grid-cols-3 gap-3 text-center">
          <div class="p-3 rounded-xl bg-dark-900 border border-dark-600/60">
            <div class="text-[10px] uppercase font-bold text-slate-400">Total Checked</div>
            <div class="text-base font-bold font-mono text-slate-100 mt-0.5">{report.totalChecked}</div>
          </div>
          <div class="p-3 rounded-xl bg-dark-900 border border-dark-600/60">
            <div class="text-[10px] uppercase font-bold text-slate-400">Missing / Damaged</div>
            <div class="text-base font-bold font-mono {report.missingCount > 0 ? 'text-amber-400' : 'text-slate-100'} mt-0.5">
              {report.missingCount}
            </div>
          </div>
          <div class="p-3 rounded-xl bg-dark-900 border border-dark-600/60">
            <div class="text-[10px] uppercase font-bold text-slate-400">Restored</div>
            <div class="text-base font-bold font-mono {report.repairedCount > 0 ? 'text-emerald-400' : 'text-slate-100'} mt-0.5">
              {report.repairedCount}
            </div>
          </div>
        </div>

        <!-- Repaired Files List (if any) -->
        {#if report.repairedFiles && report.repairedFiles.length > 0}
          <div class="space-y-1.5">
            <div class="text-[11px] font-bold uppercase tracking-wider text-slate-400 flex items-center gap-1.5">
              <Wrench class="w-3.5 h-3.5 text-indigo-400" />
              <span>Restored Files ({report.repairedFiles.length})</span>
            </div>
            <div class="p-3 rounded-xl bg-dark-900 border border-dark-600/60 text-xs font-mono text-slate-300 max-h-36 overflow-y-auto space-y-1 shadow-inner">
              {#each report.repairedFiles as file}
                <div class="flex items-center gap-2 text-emerald-400">
                  <span class="text-slate-500">&bull;</span>
                  <span class="truncate">{file}</span>
                </div>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Log File Note & Button -->
        <div class="p-3.5 rounded-xl bg-dark-900/60 border border-dark-600/40 flex items-center justify-between gap-3 text-xs">
          <div class="flex items-center gap-2 text-slate-400 truncate min-w-0">
            <FileText class="w-4 h-4 text-slate-500 flex-shrink-0" />
            <span class="truncate font-mono text-[11px]">save/logs/verify_report.log</span>
          </div>
          <button
            type="button"
            class="text-indigo-400 hover:text-indigo-300 font-semibold flex items-center gap-1 flex-shrink-0 transition"
            on:click={() => projectStore.openLogs()}
          >
            <FolderOpen class="w-3.5 h-3.5" />
            <span>Open Logs</span>
          </button>
        </div>
      {/if}
    </div>

    <!-- Footer -->
    <div class="p-5 border-t border-dark-600/60 bg-dark-850 flex items-center justify-end gap-3">
      {#if !isLoading}
        <button
          type="button"
          class="px-6 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold shadow-md shadow-indigo-600/30 transition"
          on:click={onClose}
        >
          OK, Done
        </button>
      {/if}
    </div>
  </div>
</div>
