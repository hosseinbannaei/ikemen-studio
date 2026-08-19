import { writable } from 'svelte/store';
import type { ReleaseInfo, InstalledEngine, DownloadProgress } from '../types';
import { FetchAvailableEngines, GetInstalledEngines, DownloadEngine, DeleteEngine } from '../../../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime';
import { toastStore } from './toastStore';

interface EngineState {
  available: ReleaseInfo[];
  installed: InstalledEngine[];
  downloads: Record<string, DownloadProgress>;
  loadingAvailable: boolean;
  loadingInstalled: boolean;
}

let eventsInitialized = false;

function createEngineStore() {
  const { subscribe, update } = writable<EngineState>({
    available: [],
    installed: [],
    downloads: {},
    loadingAvailable: false,
    loadingInstalled: false,
  });

  function initEvents() {
    if (eventsInitialized) return;
    try {
      if (typeof window !== 'undefined' && (window as any).runtime) {
        try {
          EventsOff('engine-download-progress');
        } catch {}

        EventsOn('engine-download-progress', (progress: DownloadProgress) => {
          update((state) => {
            const downloads = { ...state.downloads, [progress.version]: progress };

            if (progress.status === 'completed') {
              toastStore.success('Download Completed', `Engine ${progress.version} is ready to use`);
              setTimeout(() => {
                loadInstalled();
              }, 300);
            } else if (progress.status === 'error') {
              toastStore.error('Download Failed', `Error downloading ${progress.version}: ${progress.error || 'Unknown error'}`);
            }

            return { ...state, downloads };
          });
        });
        eventsInitialized = true;
      }
    } catch (e) {
      console.warn('Wails EventsOn init error:', e);
    }
  }

  async function loadAvailable() {
    initEvents();
    update((s) => ({ ...s, loadingAvailable: true }));
    try {
      const releases = await FetchAvailableEngines();
      update((s) => ({ ...s, available: releases as any[], loadingAvailable: false }));
    } catch (err: any) {
      console.error('Failed to fetch engine releases:', err);
      toastStore.error('GitHub API Error', err?.message || 'Could not fetch Ikemen GO releases');
      update((s) => ({ ...s, loadingAvailable: false }));
    }
  }

  async function loadInstalled() {
    initEvents();
    update((s) => ({ ...s, loadingInstalled: true }));
    try {
      const installed = await GetInstalledEngines();
      update((s) => ({ ...s, installed: installed as any[], loadingInstalled: false }));
    } catch (err: any) {
      console.error('Failed to load installed engines:', err);
      update((s) => ({ ...s, loadingInstalled: false }));
    }
  }

  async function startDownload(tag: string) {
    initEvents();
    try {
      update((s) => ({
        ...s,
        downloads: {
          ...s.downloads,
          [tag]: {
            version: tag,
            percent: 0,
            downloadedBytes: 0,
            totalBytes: 0,
            status: 'downloading',
          },
        },
      }));

      await DownloadEngine(tag);
      toastStore.info('Download Started', `Downloading Ikemen GO ${tag}...`);
    } catch (err: any) {
      toastStore.error('Download Error', err?.message || 'Failed to start download');
      update((s) => {
        const downloads = { ...s.downloads };
        delete downloads[tag];
        return { ...s, downloads };
      });
    }
  }

  async function removeEngine(version: string) {
    try {
      await DeleteEngine(version);
      toastStore.success('Engine Removed', `Version ${version} deleted successfully`);
      await loadInstalled();
    } catch (err: any) {
      toastStore.error('Delete Failed', err?.message || 'Failed to remove engine version');
    }
  }

  return {
    subscribe,
    initEvents,
    loadAvailable,
    loadInstalled,
    startDownload,
    removeEngine,
  };
}

export const engineStore = createEngineStore();
