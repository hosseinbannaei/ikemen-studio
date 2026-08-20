<script lang="ts">
  import { projectStore } from '../stores/projectStore';
  import type { VerificationReport } from '../types';
  import {
    ShieldCheck,
    Wrench,
    Sparkles,
    CheckCircle2,
    AlertCircle,
    X,
    Loader2,
    FileCode,
    Check,
  } from 'lucide-svelte';

  export let onClose: () => void;
  export let onReport: (report: VerificationReport) => void;

  let updateCoreSystem = false;
  let running = false;

  async function handleStartVerification() {
    running = true;
    const report = await projectStore.verifyAndRepairWithMode(updateCoreSystem);
    running = false;
    onClose();
    if (report) {
      onReport(report);
    }
  }
</script>

<div class="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
  <div class="bg-dark-800 border border-dark-600/80 rounded-2xl w-full max-w-lg shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-150 flex flex-col max-h-[90vh]">
    <!-- Header -->
    <div class="p-5 border-b border-dark-600/60 flex items-center justify-between bg-dark-850">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 flex items-center justify-center flex-shrink-0">
          <ShieldCheck class="w-5 h-5" />
        </div>
        <div>
          <h2 class="text-base font-bold text-slate-100">Verify & Repair Project Files</h2>
          <p class="text-xs text-slate-400">Validate game assets, restore missing files, or update core scripts</p>
        </div>
      </div>
      <button
        type="button"
        disabled={running}
        class="p-2 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-dark-700/60 transition disabled:opacity-50"
        on:click={onClose}
      >
        <X class="w-5 h-5" />
      </button>
    </div>

    <!-- Body -->
    <div class="p-6 space-y-5 overflow-y-auto flex-1 text-xs">
      <!-- Mode Selection -->
      <div class="space-y-3">
        <span class="block text-[11px] font-bold uppercase tracking-wider text-slate-400">Choose Verification Scope</span>

        <!-- Standard Safe Integrity Option -->
        <button
          type="button"
          class="w-full p-4 rounded-xl border text-left flex items-start gap-3.5 transition {
            !updateCoreSystem
              ? 'border-indigo-500 bg-indigo-950/20 ring-1 ring-indigo-500/30'
              : 'border-dark-600/80 bg-dark-900/60 text-slate-400 hover:text-slate-200'
          }"
          on:click={() => (updateCoreSystem = false)}
        >
          <div class="mt-0.5 w-4 h-4 rounded-full border flex items-center justify-center flex-shrink-0 {
            !updateCoreSystem ? 'border-indigo-400 bg-indigo-500' : 'border-dark-500 bg-dark-800'
          }">
            {#if !updateCoreSystem}
              <div class="w-1.5 h-1.5 rounded-full bg-white"></div>
            {/if}
          </div>

          <div class="space-y-1.5">
            <div class="font-bold text-slate-100 flex items-center gap-1.5">
              <span>Standard Asset Integrity (Safe Restore)</span>
              <span class="text-[10px] px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400 font-mono">Recommended</span>
            </div>
            <p class="text-[11px] text-slate-400 leading-relaxed">
              Scans for missing runtime DLLs, default shaders, and missing assets. Restores only files that are absent.
            </p>
            <div class="pt-1 flex flex-wrap gap-2 text-[10px] font-mono text-slate-300">
              <span class="px-2 py-0.5 rounded bg-dark-800 border border-dark-600 text-emerald-300">✓ Keeps all custom chars & stages</span>
              <span class="px-2 py-0.5 rounded bg-dark-800 border border-dark-600 text-emerald-300">✓ Keeps roster (select.def)</span>
            </div>
          </div>
        </button>

        <!-- Core System Update Option -->
        <button
          type="button"
          class="w-full p-4 rounded-xl border text-left flex items-start gap-3.5 transition {
            updateCoreSystem
              ? 'border-purple-500 bg-purple-950/20 ring-1 ring-purple-500/30'
              : 'border-dark-600/80 bg-dark-900/60 text-slate-400 hover:text-slate-200'
          }"
          on:click={() => (updateCoreSystem = true)}
        >
          <div class="mt-0.5 w-4 h-4 rounded-full border flex items-center justify-center flex-shrink-0 {
            updateCoreSystem ? 'border-purple-400 bg-purple-500' : 'border-dark-500 bg-dark-800'
          }">
            {#if updateCoreSystem}
              <div class="w-1.5 h-1.5 rounded-full bg-white"></div>
            {/if}
          </div>

          <div class="space-y-1.5">
            <div class="font-bold text-slate-100 flex items-center gap-1.5">
              <span>Core Engine Data Update (Fix common1.cns.zss & Battle Crashes)</span>
            </div>
            <p class="text-[11px] text-slate-400 leading-relaxed">
              Replaces legacy engine system scripts (<code class="text-purple-300">data/common1.cns.zss</code>, <code class="text-purple-300">data/dgl.zss</code>, <code class="text-purple-300">external/</code>) with clean files from your current engine version.
            </p>
            <div class="pt-1 flex flex-wrap gap-2 text-[10px] font-mono text-slate-300">
              <span class="px-2 py-0.5 rounded bg-dark-800 border border-dark-600 text-amber-300">Fixes 'Missing [' syntax errors</span>
              <span class="px-2 py-0.5 rounded bg-dark-800 border border-dark-600 text-emerald-300">✓ 100% preserves your chars & roster</span>
            </div>
          </div>
        </button>
      </div>

      <!-- Preserved Content Guarantee Banner -->
      <div class="p-3.5 rounded-xl bg-dark-900/70 border border-dark-600/60 flex items-center gap-3">
        <CheckCircle2 class="w-4 h-4 text-emerald-400 flex-shrink-0" />
        <div class="text-[11px] text-slate-300 leading-tight">
          <strong class="text-slate-100">Safe Operation Guarantee:</strong> Your characters (<code class="text-indigo-300">chars/</code>), stages (<code class="text-indigo-300">stages/</code>), music (<code class="text-indigo-300">sound/</code>), and roster (<code class="text-indigo-300">data/select.def</code>) are strictly protected and will never be deleted.
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="p-4 border-t border-dark-600/60 bg-dark-850 flex items-center justify-between">
      <button
        type="button"
        disabled={running}
        class="px-4 py-2 rounded-xl text-slate-400 hover:text-slate-200 text-xs font-semibold hover:bg-dark-700/60 transition disabled:opacity-50"
        on:click={onClose}
      >
        Cancel
      </button>

      <button
        type="button"
        disabled={running}
        class="px-6 py-2.5 rounded-xl {
          updateCoreSystem ? 'bg-purple-600 hover:bg-purple-500 shadow-purple-950/50' : 'bg-indigo-600 hover:bg-indigo-500 shadow-indigo-950/50'
        } disabled:opacity-50 text-white text-xs font-bold shadow-md flex items-center gap-2 transition"
        on:click={handleStartVerification}
      >
        {#if running}
          <Loader2 class="w-3.5 h-3.5 animate-spin" />
          <span>Verifying & Repairing...</span>
        {:else}
          <Wrench class="w-3.5 h-3.5" />
          <span>{updateCoreSystem ? 'Update Core Engine & Verify' : 'Start Verification'}</span>
        {/if}
      </button>
    </div>
  </div>
</div>
