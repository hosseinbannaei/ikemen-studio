<script lang="ts">
  import { AlertTriangle, X } from 'lucide-svelte';

  export let title = 'Confirm Action';
  export let message = 'Are you sure you want to proceed?';
  export let confirmLabel = 'Confirm';
  export let confirmVariant: 'danger' | 'primary' = 'danger';
  export let onConfirm: () => void;
  export let onCancel: () => void;
</script>

<div class="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
  <div class="bg-dark-800 border border-dark-600/80 rounded-2xl w-full max-w-md shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-150">
    <!-- Header -->
    <div class="p-5 border-b border-dark-600/60 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <div class="w-9 h-9 rounded-xl {confirmVariant === 'danger' ? 'bg-rose-500/10 text-rose-400 border-rose-500/20' : 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20'} border flex items-center justify-center">
          <AlertTriangle class="w-5 h-5" />
        </div>
        <h2 class="text-sm font-bold text-slate-100">{title}</h2>
      </div>
      <button
        type="button"
        class="p-2 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-dark-700/60 transition"
        on:click={onCancel}
      >
        <X class="w-4 h-4" />
      </button>
    </div>

    <!-- Body -->
    <div class="p-5 text-xs text-slate-300 leading-relaxed">
      {message}
    </div>

    <!-- Footer -->
    <div class="p-4 border-t border-dark-600/60 bg-dark-850 flex items-center justify-end gap-2.5">
      <button
        type="button"
        class="px-4 py-2 rounded-xl text-slate-400 hover:text-slate-200 text-xs font-semibold hover:bg-dark-700/60 transition"
        on:click={onCancel}
      >
        Cancel
      </button>
      <button
        type="button"
        class="px-5 py-2 rounded-xl text-xs font-bold shadow-md transition {
          confirmVariant === 'danger'
            ? 'bg-rose-600 hover:bg-rose-500 text-white shadow-rose-950/40'
            : 'bg-indigo-600 hover:bg-indigo-500 text-white shadow-indigo-950/40'
        }"
        on:click={onConfirm}
      >
        {confirmLabel}
      </button>
    </div>
  </div>
</div>
