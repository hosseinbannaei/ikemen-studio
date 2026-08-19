import { writable } from 'svelte/store';
import type { ProjectManifest } from '../types';
import {
  CreateProject,
  OpenProject,
  GetRecentProjects,
  SelectProjectDirectoryDialog,
  LaunchProject,
  StopProject,
  IsProjectRunning,
  OpenFolderInExplorer,
} from '../../../wailsjs/go/main/App';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import { toastStore } from './toastStore';

interface ProjectState {
  current: ProjectManifest | null;
  recent: string[];
  isRunning: boolean;
  loading: boolean;
}

let eventsInitialized = false;

function createProjectStore() {
  const { subscribe, update, set } = writable<ProjectState>({
    current: null,
    recent: [],
    isRunning: false,
    loading: false,
  });

  function initEvents() {
    if (eventsInitialized) return;
    try {
      if (typeof window !== 'undefined' && (window as any).runtime) {
        EventsOn('game-started', (projectDir: string) => {
          update((s) => ({ ...s, isRunning: true }));
          toastStore.info('Game Started', 'Ikemen GO is running');
        });

        EventsOn('game-stopped', (data: any) => {
          update((s) => ({ ...s, isRunning: false }));
          if (data?.error) {
            toastStore.warning('Game Exited', 'Ikemen GO terminated with an error or abnormal code');
          } else {
            toastStore.info('Game Closed', 'Ikemen GO process finished');
          }
        });
        eventsInitialized = true;
      }
    } catch (e) {
      console.warn('Wails EventsOn init error:', e);
    }
  }

  async function loadRecent() {
    initEvents();
    try {
      const recent = await GetRecentProjects();
      update((s) => ({ ...s, recent: recent || [] }));
    } catch (err: any) {
      console.error('Failed to load recent projects:', err);
    }
  }

  async function create(name: string, targetDir: string, engineVersion: string, author: string): Promise<boolean> {
    initEvents();
    update((s) => ({ ...s, loading: true }));
    try {
      const manifest = await CreateProject(name, targetDir, engineVersion, author);
      update((s) => ({ ...s, current: manifest as any, loading: false }));
      toastStore.success('Project Created', `Successfully initialized ${name}`);
      await loadRecent();
      return true;
    } catch (err: any) {
      console.error('Project creation failed:', err);
      toastStore.error('Creation Failed', err?.message || 'Could not scaffold project');
      update((s) => ({ ...s, loading: false }));
      return false;
    }
  }

  async function open(projectDir: string): Promise<boolean> {
    initEvents();
    update((s) => ({ ...s, loading: true }));
    try {
      const manifest = await OpenProject(projectDir);
      const running = await IsProjectRunning(projectDir);
      update((s) => ({ ...s, current: manifest as any, isRunning: running, loading: false }));
      toastStore.success('Project Opened', manifest.name);
      await loadRecent();
      return true;
    } catch (err: any) {
      console.error('Failed to open project:', err);
      toastStore.error('Open Failed', err?.message || 'Could not open Ikemen project');
      update((s) => ({ ...s, loading: false }));
      return false;
    }
  }

  async function selectAndOpen(): Promise<boolean> {
    try {
      const dir = await SelectProjectDirectoryDialog();
      if (dir) {
        return await open(dir);
      }
      return false;
    } catch (err: any) {
      toastStore.error('Dialog Error', err?.message || 'Failed to select directory');
      return false;
    }
  }

  function close() {
    update((s) => ({ ...s, current: null, isRunning: false }));
  }

  async function launch(): Promise<void> {
    initEvents();
    let currentPath = '';
    update((s) => {
      currentPath = s.current?.path || '';
      return s;
    });

    if (!currentPath) {
      toastStore.error('Launch Error', 'No active project to launch');
      return;
    }

    try {
      await LaunchProject(currentPath);
    } catch (err: any) {
      console.error('Failed to launch project:', err);
      toastStore.error('Launch Error', err?.message || 'Failed to start Ikemen GO');
    }
  }

  async function stop(): Promise<void> {
    let currentPath = '';
    update((s) => {
      currentPath = s.current?.path || '';
      return s;
    });

    if (!currentPath) return;

    try {
      await StopProject(currentPath);
      toastStore.info('Game Stopping', 'Terminating game process...');
    } catch (err: any) {
      toastStore.error('Stop Failed', err?.message || 'Could not stop running game');
    }
  }

  async function openFolder(subPath = ''): Promise<void> {
    let rootPath = '';
    update((s) => {
      rootPath = s.current?.path || '';
      return s;
    });

    if (!rootPath) return;

    const fullPath = subPath ? `${rootPath}/${subPath}` : rootPath;
    try {
      await OpenFolderInExplorer(fullPath);
    } catch (err: any) {
      toastStore.error('Explorer Error', err?.message || 'Could not open folder in file manager');
    }
  }

  return {
    subscribe,
    initEvents,
    loadRecent,
    create,
    open,
    selectAndOpen,
    close,
    launch,
    stop,
    openFolder,
  };
}

export const projectStore = createProjectStore();
