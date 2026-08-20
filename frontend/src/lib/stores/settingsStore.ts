import { writable } from 'svelte/store';
import type { Settings } from '../types';
import { GetSettings, UpdateSettings, SelectDirectoryDialog } from '../../../wailsjs/go/main/App';
import { toastStore } from './toastStore';

function applyTheme(theme: string) {
  if (typeof document !== 'undefined') {
    if (theme === 'light') {
      document.documentElement.classList.add('light');
      document.documentElement.classList.remove('dark');
    } else {
      document.documentElement.classList.add('dark');
      document.documentElement.classList.remove('light');
    }
  }
}

function createSettingsStore() {
  const { subscribe, set, update } = writable<Settings>({
    enginesDir: '',
    theme: 'dark',
    recentProjects: [],
    defaultChannel: 'stable',
  });

  async function load() {
    try {
      const s = await GetSettings();
      set(s);
      applyTheme(s.theme || 'dark');
    } catch (err: any) {
      console.error('Failed to load settings:', err);
      toastStore.error('Settings Error', err?.message || 'Failed to load settings');
    }
  }

  async function save(newSettings: Settings) {
    try {
      await UpdateSettings(newSettings as any);
      set(newSettings);
      applyTheme(newSettings.theme);
      toastStore.success('Settings Saved');
    } catch (err: any) {
      console.error('Failed to save settings:', err);
      toastStore.error('Save Failed', err?.message || 'Failed to update settings');
    }
  }

  async function toggleTheme() {
    update((s) => {
      const nextTheme = s.theme === 'light' ? 'dark' : 'light';
      const updated = { ...s, theme: nextTheme };
      applyTheme(nextTheme);
      save(updated);
      return updated;
    });
  }

  async function chooseEnginesDir(): Promise<string | null> {
    try {
      const selected = await SelectDirectoryDialog('Select Engines Directory');
      if (selected) {
        update((s) => ({ ...s, enginesDir: selected }));
        return selected;
      }
      return null;
    } catch (err: any) {
      toastStore.error('Directory Picker Error', err?.message || 'Failed to select directory');
      return null;
    }
  }

  return {
    subscribe,
    load,
    save,
    toggleTheme,
    chooseEnginesDir,
  };
}

export const settingsStore = createSettingsStore();
