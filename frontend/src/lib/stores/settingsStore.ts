import { writable } from 'svelte/store';
import type { Settings, ThemeId, RadiusStyle, ThemePreset } from '../types';
import { GetSettings, UpdateSettings, SelectDirectoryDialog } from '../../../wailsjs/go/main/App';
import { toastStore } from './toastStore';

export const THEME_PRESETS: ThemePreset[] = [
  {
    id: 'mkx',
    name: 'NetherRealm MKX',
    subtitle: 'Cold Carbon & Crimson (Default)',
    description: 'Deep cold charcoal and visceral blood crimson accents, inspired by Mortal Kombat X.',
    accentHex: '#e11d48',
    bgHex: '#0c0d11',
    cardHex: '#151820',
    borderHex: '#262b3a',
    tag: 'FIGHTING DARK',
  },
  {
    id: 'obsidian',
    name: 'Arcade Obsidian',
    subtitle: 'Pure OLED Minimal Black',
    description: 'Ultra-minimal pure pitch black with sharp high-contrast neon cyan & sky accents.',
    accentHex: '#0ea5e9',
    bgHex: '#050505',
    cardHex: '#101010',
    borderHex: '#262626',
    tag: 'MINIMAL OLED',
  },
  {
    id: 'cyber',
    name: 'Cyber Strike',
    subtitle: 'Tekken Electric Night',
    description: 'Deep night-blue chassis with high-voltage electric blue accents and neon glow.',
    accentHex: '#3b82f6',
    bgHex: '#080b16',
    cardHex: '#10162a',
    borderHex: '#202c50',
    tag: 'CYBERPUNK',
  },
  {
    id: 'capcom',
    name: 'Capcom Classic',
    subtitle: 'Street Fighter Navy & Gold',
    description: 'Classic arcade tournament dark navy with championship amber gold highlights.',
    accentHex: '#f59e0b',
    bgHex: '#0d1222',
    cardHex: '#18213c',
    borderHex: '#2c3c66',
    tag: 'TOURNAMENT',
  },
  {
    id: 'light',
    name: 'Clean Studio',
    subtitle: 'Modern Light Mode',
    description: 'High-contrast clean workspace with crisp indigo accents for bright environments.',
    accentHex: '#4f46e5',
    bgHex: '#f1f5f9',
    cardHex: '#ffffff',
    borderHex: '#cbd5e1',
    tag: 'LIGHT MODE',
  },
];

function applyThemeAttributes(theme: string, radius: RadiusStyle = 'sharp') {
  if (typeof document !== 'undefined') {
    const root = document.documentElement;
    root.setAttribute('data-theme', theme);
    root.setAttribute('data-radius', radius);

    if (theme === 'light') {
      root.classList.add('light');
      root.classList.remove('dark');
    } else {
      root.classList.add('dark');
      root.classList.remove('light');
    }
  }
}

function createSettingsStore() {
  const { subscribe, set, update } = writable<Settings>({
    enginesDir: '',
    theme: 'mkx',
    radiusStyle: 'sharp',
    recentProjects: [],
    defaultChannel: 'stable',
  });

  async function load() {
    try {
      const s = await GetSettings();
      const loadedTheme = s.theme || 'mkx';
      const loadedRadius = (s.radiusStyle as RadiusStyle) || 'sharp';
      const mergedSettings: Settings = {
        ...s,
        theme: loadedTheme,
        radiusStyle: loadedRadius,
      };
      set(mergedSettings);
      applyThemeAttributes(loadedTheme, loadedRadius);
    } catch (err: any) {
      console.error('Failed to load settings:', err);
      applyThemeAttributes('mkx', 'sharp');
    }
  }

  async function save(newSettings: Settings) {
    try {
      await UpdateSettings(newSettings as any);
      set(newSettings);
      applyThemeAttributes(newSettings.theme, (newSettings.radiusStyle as RadiusStyle) || 'sharp');
      toastStore.success('Settings Saved');
    } catch (err: any) {
      console.error('Failed to save settings:', err);
      toastStore.error('Save Failed', err?.message || 'Failed to update settings');
    }
  }

  async function setTheme(themeId: ThemeId) {
    update((s) => {
      const updated: Settings = { ...s, theme: themeId };
      applyThemeAttributes(themeId, (s.radiusStyle as RadiusStyle) || 'sharp');
      save(updated);
      return updated;
    });
  }

  async function setRadius(radiusStyle: RadiusStyle) {
    update((s) => {
      const updated: Settings = { ...s, radiusStyle };
      applyThemeAttributes(s.theme || 'mkx', radiusStyle);
      save(updated);
      return updated;
    });
  }

  async function toggleTheme() {
    update((s) => {
      const nextTheme = s.theme === 'light' ? 'mkx' : 'light';
      const updated = { ...s, theme: nextTheme };
      applyThemeAttributes(nextTheme, (s.radiusStyle as RadiusStyle) || 'sharp');
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
    setTheme,
    setRadius,
    toggleTheme,
    chooseEnginesDir,
  };
}

export const settingsStore = createSettingsStore();
