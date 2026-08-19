<script lang="ts">
  import { toastStore } from '../stores/toastStore';
  import { CheckCircle2, AlertCircle, Info, AlertTriangle, X } from 'lucide-svelte';

  const icons = {
    success: CheckCircle2,
    error: AlertCircle,
    warning: AlertTriangle,
    info: Info,
  };

  const colorStyles = {
    success: 'bg-emerald-950/90 border-emerald-500/50 text-emerald-200',
    error: 'bg-rose-950/90 border-rose-500/50 text-rose-200',
    warning: 'bg-amber-950/90 border-amber-500/50 text-amber-200',
    info: 'bg-sky-950/90 border-sky-500/50 text-sky-200',
  };

  const iconColors = {
    success: 'text-emerald-400',
    error: 'text-rose-400',
    warning: 'text-amber-400',
    info: 'text-sky-400',
  };
</script>

<div class="fixed bottom-4 right-4 z-50 flex flex-col gap-2 max-w-md w-full pointer-events-none">
  {#each $toastStore as toast (toast.id)}
    <div
      class="pointer-events-auto flex items-start gap-3 p-3.5 rounded-lg border shadow-xl backdrop-blur-md transition-all duration-200 {colorStyles[toast.type]}"
    >
      <svelte:component this={icons[toast.type]} class="w-5 h-5 flex-shrink-0 mt-0.5 {iconColors[toast.type]}" />
      <div class="flex-1 text-sm">
        <div class="font-semibold">{toast.title}</div>
        {#if toast.message}
          <div class="text-xs opacity-90 mt-0.5 leading-relaxed">{toast.message}</div>
        {/if}
      </div>
      <button
        type="button"
        class="opacity-60 hover:opacity-100 p-0.5 rounded transition"
        on:click={() => toastStore.dismiss(toast.id)}
      >
        <X class="w-4 h-4" />
      </button>
    </div>
  {/each}
</div>
