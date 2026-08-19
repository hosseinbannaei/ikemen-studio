import { writable } from 'svelte/store';
import type { ToastMessage } from '../types';

let lastToastKey = '';
let lastToastTime = 0;

function createToastStore() {
  const { subscribe, update } = writable<ToastMessage[]>([]);

  function show(toast: Omit<ToastMessage, 'id'>) {
    const key = `${toast.type}:${toast.title}:${toast.message || ''}`;
    const now = Date.now();
    if (key === lastToastKey && now - lastToastTime < 1500) {
      // Deduplicate rapid identical toasts
      return;
    }
    lastToastKey = key;
    lastToastTime = now;

    const id = Math.random().toString(36).substring(2, 9);
    const duration = toast.duration ?? 4000;
    const newToast: ToastMessage = { ...toast, id, duration };

    update((toasts) => [...toasts, newToast]);

    if (duration > 0) {
      setTimeout(() => {
        dismiss(id);
      }, duration);
    }
  }

  function dismiss(id: string) {
    update((toasts) => toasts.filter((t) => t.id !== id));
  }

  return {
    subscribe,
    show,
    success: (title: string, message?: string) => show({ type: 'success', title, message }),
    error: (title: string, message?: string) => show({ type: 'error', title, message, duration: 6000 }),
    info: (title: string, message?: string) => show({ type: 'info', title, message }),
    warning: (title: string, message?: string) => show({ type: 'warning', title, message }),
    dismiss,
  };
}

export const toastStore = createToastStore();
