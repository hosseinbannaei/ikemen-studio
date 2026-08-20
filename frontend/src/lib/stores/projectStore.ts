import { writable } from 'svelte/store';
import type {
  ProjectManifest,
  GameState,
  VerificationReport,
  CrashDiagnosticInfo,
  ExistingGameInspection,
  EngineBackupInfo,
  ImportOptions,
  ProjectDiffSummary,
  AssetSyncOptions,
  ConfigInspectionResult,
} from '../types';
import {
  CreateProject,
  OpenProject,
  GetRecentProjects,
  SelectProjectDirectoryDialog,
  LaunchProject,
  LaunchProjectWithOptions,
  StopProject,
  IsProjectRunning,
  OpenFolderInExplorer,
  VerifyAndRepairProject,
  VerifyAndRepairProjectWithMode,
  OpenProjectLogsFolder,
  DetectExistingGame,
  ImportExistingGame,
  ImportExistingGameWithOptions,
  SwitchProjectEngine,
  GetEngineBackups,
  RollbackProjectEngine,
  GetGameConfig,
  SaveGameConfig,
  RemoveRecentProject,
  InspectProjectDifferences,
  SyncProjectAssets,
  InspectGameConfig,
  RepairGameConfig,
  ResetGameConfig,
  GetProjectLogs,
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
  backups: EngineBackupInfo[];
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
    backups: [],
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

  async function loadBackups(projectDir?: string) {
    let targetPath = projectDir || '';
    if (!targetPath) {
      update((s) => {
        targetPath = s.current?.path || '';
        return s;
      });
    }
    if (!targetPath) return;

    try {
      const backups = await GetEngineBackups(targetPath);
      update((s) => ({ ...s, backups: (backups as any) || [] }));
    } catch (e) {
      console.error('Failed to load engine backups:', e);
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
      await loadBackups(manifest.path);
      return true;
    } catch (err: any) {
      console.error('Project creation failed:', err);
      toastStore.error('Creation Failed', err?.message || 'Could not scaffold project');
      update((s) => ({ ...s, loading: false }));
      return false;
    }
  }

  async function open(projectDir: string, silent = false): Promise<boolean> {
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
      if (!silent) {
        toastStore.success('Project Opened', manifest.name);
      }
      await loadRecent();
      await loadBackups(projectDir);
      return true;
    } catch (err: any) {
      if (!silent) {
        console.error('Failed to open project:', err);
        toastStore.error('Open Failed', err?.message || 'Could not open Ikemen project');
      }
      update((s) => ({ ...s, loading: false }));
      return false;
    }
  }

  async function checkRawGame(dir: string): Promise<ExistingGameInspection | null> {
    try {
      const inspection = await DetectExistingGame(dir);
      return inspection as any;
    } catch {
      return null;
    }
  }

  async function importExisting(
    srcDir: string,
    targetDir: string,
    name: string,
    engineVersion: string,
    author: string
  ): Promise<boolean> {
    return importExistingWithOptions({
      sourceDir: srcDir,
      targetDir,
      projectName: name,
      engineVersion,
      author,
      mode: 'rebuild',
      includeChars: true,
      includeStages: true,
      includeSound: true,
      includeFonts: true,
      includeRoster: true,
    });
  }

  async function importExistingWithOptions(opts: ImportOptions): Promise<boolean> {
    initEvents();
    update((s) => ({ ...s, loading: true }));
    try {
      const manifest = await ImportExistingGameWithOptions(opts as any);
      update((s) => ({ ...s, current: manifest as any, loading: false }));
      toastStore.success('Game Imported', `Successfully imported ${manifest.name}`);
      await loadRecent();
      await loadBackups(manifest.path);
      return true;
    } catch (err: any) {
      console.error('Import failed:', err);
      toastStore.error('Import Failed', err?.message || 'Could not import game');
      update((s) => ({ ...s, loading: false }));
      return false;
    }
  }

  async function removeRecent(projectDir: string): Promise<void> {
    try {
      await RemoveRecentProject(projectDir);
      toastStore.info('Removed from Recent', 'Project removed from workspace list');
      await loadRecent();
    } catch (err: any) {
      toastStore.error('Removal Error', err?.message || 'Could not remove project from recent list');
    }
  }

  async function selectAndOpen(): Promise<{ opened: boolean; inspection?: ExistingGameInspection; selectedPath?: string }> {
    try {
      const dir = await SelectProjectDirectoryDialog();
      if (!dir) return { opened: false };

      // Attempt silent open for studio project
      const ok = await open(dir, true);
      if (ok) {
        toastStore.success('Project Opened');
        return { opened: true };
      }

      // If manifest open failed, check if it is a raw game folder to import
      const inspection = await checkRawGame(dir);
      if (inspection && inspection.isValid) {
        toastStore.info('Existing Game Detected', 'Configure import settings');
        return { opened: false, inspection, selectedPath: dir };
      }

      toastStore.warning('No Project Found', 'Selected directory does not contain an Ikemen project or assets');
      return { opened: false };
    } catch (err: any) {
      toastStore.error('Dialog Error', err?.message || 'Failed to select directory');
      return { opened: false };
    }
  }

  function close() {
    update((s) => ({
      ...s,
      current: null,
      isRunning: false,
      gameState: 'idle',
      canStop: false,
      activeCrash: null,
      backups: [],
    }));
  }

  async function launch(): Promise<void> {
    await launchWithOptions([]);
  }

  async function launchWithOptions(args: string[] = []): Promise<void> {
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
      await LaunchProjectWithOptions(currentPath, args);
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

  async function switchEngine(newVersion: string): Promise<boolean> {
    let currentPath = '';
    update((s) => {
      currentPath = s.current?.path || '';
      return s;
    });

    if (!currentPath) return false;

    try {
      await SwitchProjectEngine(currentPath, newVersion);
      await open(currentPath, true);
      toastStore.success('Engine Switched', `Project updated to ${newVersion} with safety backup`);
      return true;
    } catch (err: any) {
      toastStore.error('Switch Failed', err?.message || 'Could not switch engine');
      return false;
    }
  }

  async function rollbackEngine(backupId: string): Promise<boolean> {
    let currentPath = '';
    update((s) => {
      currentPath = s.current?.path || '';
      return s;
    });

    if (!currentPath) return false;

    try {
      await RollbackProjectEngine(currentPath, backupId);
      await open(currentPath, true);
      toastStore.success('Engine Rolled Back', 'Successfully restored previous engine runtime');
      return true;
    } catch (err: any) {
      toastStore.error('Rollback Failed', err?.message || 'Could not rollback engine');
      return false;
    }
  }

  async function getGameConfig(): Promise<Record<string, string> | null> {
    let currentPath = '';
    update((s) => {
      currentPath = s.current?.path || '';
      return s;
    });

    if (!currentPath) return null;

    try {
      const cfg = await GetGameConfig(currentPath);
      return (cfg as any) || null;
    } catch (err) {
      console.error('Failed to get game config:', err);
      return null;
    }
  }

  async function saveGameConfig(updates: Record<string, string>): Promise<boolean> {
    let currentPath = '';
    update((s) => {
      currentPath = s.current?.path || '';
      return s;
    });

    if (!currentPath) return false;

    try {
      await SaveGameConfig(currentPath, updates);
      toastStore.success('Settings Saved', 'Updated save/config.ini');
      return true;
    } catch (err: any) {
      toastStore.error('Save Failed', err?.message || 'Failed to update config.ini');
      return false;
    }
  }

  async function verifyAndRepair(projectDir?: string): Promise<VerificationReport | null> {
    return verifyAndRepairWithMode(false, projectDir);
  }

  async function verifyAndRepairWithMode(updateCoreSystem: boolean, projectDir?: string): Promise<VerificationReport | null> {
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
      const report = await VerifyAndRepairProjectWithMode(targetPath, updateCoreSystem);
      update((s) => ({ ...s, isVerifying: false }));

      if (report.repairedCount > 0) {
        toastStore.success(
          updateCoreSystem ? 'Core Engine Updated' : 'Verification Complete',
          `Updated/restored ${report.repairedCount} file${report.repairedCount > 1 ? 's' : ''} (Checked ${report.totalChecked} files)`
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

  async function inspectDiff(projectDir?: string): Promise<ProjectDiffSummary | null> {
    let targetPath = projectDir || '';
    if (!targetPath) {
      update((s) => {
        targetPath = s.current?.path || '';
        return s;
      });
    }
    if (!targetPath) return null;

    try {
      const diff = await InspectProjectDifferences(targetPath);
      return diff as any;
    } catch (err: any) {
      console.error('Diff inspection error:', err);
      return null;
    }
  }

  async function syncAssets(opts: AssetSyncOptions): Promise<VerificationReport | null> {
    try {
      const report = await SyncProjectAssets(opts as any);
      toastStore.success('Asset Sync Complete', `Updated ${report.repairedCount} component(s)`);
      return report as any;
    } catch (err: any) {
      toastStore.error('Sync Error', err?.message || 'Failed to sync assets');
      return null;
    }
  }

  async function inspectConfig(projectDir?: string): Promise<ConfigInspectionResult | null> {
    let targetPath = projectDir || '';
    if (!targetPath) {
      update((s) => {
        targetPath = s.current?.path || '';
        return s;
      });
    }
    if (!targetPath) return null;

    try {
      const res = await InspectGameConfig(targetPath);
      return res as any;
    } catch (err: any) {
      console.error('Config inspection error:', err);
      return null;
    }
  }

  async function repairConfig(projectDir?: string): Promise<boolean> {
    let targetPath = projectDir || '';
    if (!targetPath) {
      update((s) => {
        targetPath = s.current?.path || '';
        return s;
      });
    }
    if (!targetPath) return false;

    try {
      await RepairGameConfig(targetPath);
      toastStore.success('Config Repaired', 'Fixed invalid RenderMode and normalized keys in save/config.ini');
      return true;
    } catch (err: any) {
      toastStore.error('Config Repair Failed', err?.message || 'Could not repair config.ini');
      return false;
    }
  }

  async function resetConfig(projectDir?: string): Promise<boolean> {
    let targetPath = projectDir || '';
    if (!targetPath) {
      update((s) => {
        targetPath = s.current?.path || '';
        return s;
      });
    }
    if (!targetPath) return false;

    try {
      await ResetGameConfig(targetPath);
      toastStore.success('Config Reset', 'Restored save/config.ini to clean engine defaults');
      return true;
    } catch (err: any) {
      toastStore.error('Config Reset Failed', err?.message || 'Could not reset config.ini');
      return false;
    }
  }

  async function getLogs(projectDir?: string): Promise<string> {
    let targetPath = projectDir || '';
    if (!targetPath) {
      update((s) => {
        targetPath = s.current?.path || '';
        return s;
      });
    }
    if (!targetPath) return 'No project selected';

    try {
      return await GetProjectLogs(targetPath);
    } catch (err) {
      return 'Error loading log file.';
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
    loadBackups,
    create,
    open,
    checkRawGame,
    importExisting,
    importExistingWithOptions,
    removeRecent,
    selectAndOpen,
    close,
    launch,
    launchWithOptions,
    stop,
    openFolder,
    switchEngine,
    rollbackEngine,
    getGameConfig,
    saveGameConfig,
    verifyAndRepair,
    verifyAndRepairWithMode,
    inspectDiff,
    syncAssets,
    inspectConfig,
    repairConfig,
    resetConfig,
    getLogs,
    openLogs,
    dismissCrash,
  };
}

export const projectStore = createProjectStore();
