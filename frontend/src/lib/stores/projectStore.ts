import { writable } from 'svelte/store';
import type { ProjectManifest, GameState, VerificationReport, CrashDiagnosticInfo } from '../types';
import {
  CreateProject,
  OpenProject,
  GetRecentProjects,
  SelectProjectDirectoryDialog,
  LaunchProject,
  StopProject,
  IsProjectRunning,
  OpenFolderInExplorer,
  VerifyAndRepairProject,
  OpenProjectLogsFolder,
} from '../../../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime';
import { toastStore } from './toastStore';

interface ProjectState {
  current: ProjectManifest | null;
  recent: string[];
  isRunning: boolean;
  gameState: GameState;
  canStop: boolean;
  loading: boolean;
  isVerifying: boolean;
  activeCrash: CrashDiagnosticInfo | null;
}

let eventsInitialized = false;
let stopCooldownTimer: any = null;

function createProjectStore() {
  const { subscribe, update, set } = writable<ProjectState>({
    current: null,
    recent: [],
    isRunning: false,
    gameState: 'idle',
    canStop: false,
    loading: false,
    isVerifying: false,
    activeCrash: null,
  });

  function initEvents() {
    if (eventsInitialized) return;
    try {
      if (typeof window !== 'undefined' && (window as any).runtime) {
        try {
          EventsOff('game-started');
          EventsOff('game-stopped');
          EventsOff('game-crashed');
        } catch {}

        EventsOn('game-started', (projectDir: string) => {
          update((s) => ({
            ...s,
            isRunning: true,
            gameState: 'running',
            canStop: false,
            activeCrash: null,
          }));

          if (stopCooldownTimer) clearTimeout(stopCooldownTimer);
          stopCooldownTimer = setTimeout(() => {
            update((s) => ({ ...s, canStop: true }));
          }, 1500);

          toastStore.info('Game Started', 'Ikemen GO is running');
        });

        EventsOn('game-crashed', (data: CrashDiagnosticInfo) => {
          update((s) => ({
            ...s,
            activeCrash: data,
          }));
        });

        EventsOn('game-stopped', (data: any) => {
          if (stopCooldownTimer) clearTimeout(stopCooldownTimer);
          update((s) => ({
            ...s,
            isRunning: false,
            gameState: 'idle',
            canStop: false,
          }));

          if (data?.userTerminated) {
            toastStore.info('Game Stopped', 'Game process terminated');
          } else if (!data?.error) {
            toastStore.info('Game Closed', 'Ikemen GO session ended');
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
      update((s) => ({
        ...s,
        current: manifest as any,
        isRunning: running,
        gameState: running ? 'running' : 'idle',
        canStop: running,
        loading: false,
      }));
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
    update((s) => ({ ...s, current: null, isRunning: false, gameState: 'idle', canStop: false, activeCrash: null }));
  }

  async function launch(): Promise<void> {
    initEvents();
    let currentPath = '';
    update((s) => {
      currentPath = s.current?.path || '';
      return {
        ...s,
        gameState: 'starting',
        isRunning: true,
        canStop: false,
        activeCrash: null,
      };
    });

    if (!currentPath) {
      toastStore.error('Launch Error', 'No active project to launch');
      update((s) => ({ ...s, gameState: 'idle', isRunning: false }));
      return;
    }

    try {
      await LaunchProject(currentPath);
      setTimeout(() => {
        update((s) => {
          if (s.gameState === 'starting') {
            return { ...s, gameState: 'running', canStop: true };
          }
          return s;
        });
      }, 1500);
    } catch (err: any) {
      console.error('Failed to launch project:', err);
      toastStore.error('Launch Error', err?.message || 'Failed to start Ikemen GO');
      update((s) => ({ ...s, gameState: 'idle', isRunning: false, canStop: false }));
    }
  }

  async function stop(): Promise<void> {
    let currentPath = '';
    update((s) => {
      currentPath = s.current?.path || '';
      return {
        ...s,
        gameState: 'stopping',
      };
    });

    if (!currentPath) return;

    try {
      await StopProject(currentPath);
      toastStore.info('Game Stopping', 'Terminating game process...');
    } catch (err: any) {
      toastStore.error('Stop Failed', err?.message || 'Could not stop running game');
      update((s) => ({ ...s, gameState: 'running' }));
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

  async function verifyAndRepair(projectDir?: string): Promise<VerificationReport | null> {
    let targetPath = projectDir || '';
    if (!targetPath) {
      update((s) => {
        targetPath = s.current?.path || '';
        return s;
      });
    }

    if (!targetPath) {
      toastStore.error('Verification Error', 'No project selected for verification');
      return null;
    }

    update((s) => ({ ...s, isVerifying: true }));
    try {
      const report = await VerifyAndRepairProject(targetPath);
      update((s) => ({ ...s, isVerifying: false }));

      if (report.repairedCount > 0) {
        toastStore.success(
          'Verification Complete',
          `Restored ${report.repairedCount} missing file${report.repairedCount > 1 ? 's' : ''} (Checked ${report.totalChecked} files)`
        );
      } else {
        toastStore.success('Verification Complete', `All ${report.totalChecked} engine runtime files are intact`);
      }

      return report as any;
    } catch (err: any) {
      update((s) => ({ ...s, isVerifying: false }));
      console.error('Verification failed:', err);
      toastStore.error('Verification Failed', err?.message || 'Could not complete file verification');
      return null;
    }
  }

  async function openLogs(projectDir?: string): Promise<void> {
    let targetPath = projectDir || '';
    if (!targetPath) {
      update((s) => {
        targetPath = s.current?.path || '';
        return s;
      });
    }

    if (!targetPath) return;

    try {
      await OpenProjectLogsFolder(targetPath);
    } catch (err: any) {
      toastStore.error('Logs Error', err?.message || 'Could not open logs folder');
    }
  }

  function dismissCrash() {
    update((s) => ({ ...s, activeCrash: null }));
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
    verifyAndRepair,
    openLogs,
    dismissCrash,
  };
}

export const projectStore = createProjectStore();
