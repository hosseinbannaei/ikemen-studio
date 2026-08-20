import { writable, derived } from 'svelte/store';
import type { VaultInfo, VaultAsset, AssetCategory, AssetMetadataUpdate, IngestResult } from '../types';
import {
  GetVaults,
  CreateVault,
  RegisterVault,
  UnregisterVault,
  GetVaultAssets,
  UpdateVaultAsset,
  DeleteVaultAsset,
  IngestAsset,
  IngestAssets,
  AddVaultAssetToProject,
  RemoveVaultAssetFromProject,
  SelectArchiveDialog,
  SelectMultipleArchivesDialog,
  SelectDirectoryDialog,
} from '../../../wailsjs/go/main/App';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import { toastStore } from './toastStore';

interface VaultState {
  vaults: VaultInfo[];
  activeVaultId: string; // 'all' or specific ID
  assets: VaultAsset[];
  loading: boolean;
  isIngesting: boolean;
  ingestMessage: string;
  searchQuery: string;
  selectedCategory: AssetCategory | 'all';
  selectedTag: string | null;
  selectedAsset: VaultAsset | null;
}

let eventsInitialized = false;

function createVaultStore() {
  const { subscribe, update, set } = writable<VaultState>({
    vaults: [],
    activeVaultId: 'all',
    assets: [],
    loading: false,
    isIngesting: false,
    ingestMessage: '',
    searchQuery: '',
    selectedCategory: 'all',
    selectedTag: null,
    selectedAsset: null,
  });

  function initDropEvents() {
    if (eventsInitialized) return;
    try {
      if (typeof window !== 'undefined' && (window as any).runtime) {
        EventsOn('wails-file-drop', async (paths: string[]) => {
          if (paths && paths.length > 0) {
            toastStore.info('Files Dropped', `Importing ${paths.length} file(s)/folder(s)...`);
            let targetVault = 'vault-default';
            update((s) => {
              targetVault = s.activeVaultId !== 'all' ? s.activeVaultId : 'vault-default';
              return s;
            });
            await ingestMultiple(paths, targetVault, 'auto');
          }
        });
        eventsInitialized = true;
      }
    } catch (e) {
      console.warn('Wails file drop event init error:', e);
    }
  }

  async function loadVaults() {
    initDropEvents();
    update((s) => ({ ...s, loading: true }));
    try {
      const vaults = await GetVaults();
      update((s) => ({
        ...s,
        vaults: vaults || [],
        loading: false,
      }));
    } catch (err: any) {
      console.error('Failed to load vaults:', err);
      toastStore.error('Vault Error', err?.message || 'Could not load vaults');
      update((s) => ({ ...s, loading: false }));
    }
  }

  async function loadAssets(vaultId?: string) {
    initDropEvents();
    let targetVault = vaultId;
    if (!targetVault) {
      update((s) => {
        targetVault = s.activeVaultId;
        return s;
      });
    }

    update((s) => ({ ...s, loading: true }));
    try {
      const assets = await GetVaultAssets(targetVault || 'all');
      update((s) => ({
        ...s,
        assets: (assets as unknown as VaultAsset[]) || [],
        loading: false,
      }));
    } catch (err: any) {
      console.error('Failed to load vault assets:', err);
      toastStore.error('Asset Error', err?.message || 'Could not load vault assets');
      update((s) => ({ ...s, loading: false }));
    }
  }

  async function selectVault(vaultId: string) {
    update((s) => ({ ...s, activeVaultId: vaultId, selectedAsset: null }));
    await loadAssets(vaultId);
  }

  async function create(name: string, description: string, targetPath: string): Promise<boolean> {
    update((s) => ({ ...s, loading: true }));
    try {
      const v = await CreateVault(name, description, targetPath);
      toastStore.success('Vault Created', `Created vault: ${name}`);
      await loadVaults();
      if (v?.id) {
        await selectVault(v.id);
      }
      return true;
    } catch (err: any) {
      toastStore.error('Creation Failed', err?.message || 'Could not create vault');
      update((s) => ({ ...s, loading: false }));
      return false;
    }
  }

  async function register(path: string): Promise<boolean> {
    update((s) => ({ ...s, loading: true }));
    try {
      const v = await RegisterVault(path);
      toastStore.success('Vault Mounted', `Registered vault: ${v.name}`);
      await loadVaults();
      if (v?.id) {
        await selectVault(v.id);
      }
      return true;
    } catch (err: any) {
      toastStore.error('Registration Failed', err?.message || 'Could not register vault');
      update((s) => ({ ...s, loading: false }));
      return false;
    }
  }

  async function unregister(vaultId: string): Promise<boolean> {
    try {
      await UnregisterVault(vaultId);
      toastStore.info('Vault Unmounted', 'Vault removed from list (files preserved on disk)');
      await loadVaults();
      await selectVault('all');
      return true;
    } catch (err: any) {
      toastStore.error('Unregister Failed', err?.message || 'Could not unmount vault');
      return false;
    }
  }

  async function updateAsset(vaultId: string, assetKey: string, updates: AssetMetadataUpdate): Promise<boolean> {
    try {
      await UpdateVaultAsset(vaultId, assetKey, updates as any);
      toastStore.success('Asset Updated', 'Saved metadata changes');
      await loadAssets(vaultId);
      // Update selected asset if viewing
      update((s) => {
        if (s.selectedAsset && s.selectedAsset.key === assetKey) {
          const updated = s.assets.find((a) => a.key === assetKey);
          return { ...s, selectedAsset: updated || null };
        }
        return s;
      });
      return true;
    } catch (err: any) {
      toastStore.error('Update Failed', err?.message || 'Could not save asset changes');
      return false;
    }
  }

  async function deleteAsset(vaultId: string, assetKey: string): Promise<boolean> {
    try {
      await DeleteVaultAsset(vaultId, assetKey);
      toastStore.info('Asset Deleted', 'Removed asset from vault');
      update((s) => ({
        ...s,
        selectedAsset: s.selectedAsset?.key === assetKey ? null : s.selectedAsset,
      }));
      await loadAssets(vaultId);
      await loadVaults();
      return true;
    } catch (err: any) {
      toastStore.error('Delete Failed', err?.message || 'Could not delete asset');
      return false;
    }
  }

  async function ingest(filePath: string, vaultId: string, targetMode = 'auto'): Promise<IngestResult | null> {
    return ingestMultiple([filePath], vaultId, targetMode);
  }

  async function ingestMultiple(filePaths: string[], vaultId: string, targetMode = 'auto'): Promise<IngestResult | null> {
    if (filePaths.length === 0) return null;

    update((s) => ({
      ...s,
      isIngesting: true,
      ingestMessage: `Unpacking and scanning ${filePaths.length} item(s)...`,
    }));

    try {
      let res: any;
      if (filePaths.length === 1) {
        res = await IngestAsset(vaultId, filePaths[0], targetMode);
      } else {
        res = await IngestAssets(vaultId, filePaths, targetMode);
      }

      update((s) => ({ ...s, isIngesting: false, ingestMessage: '' }));

      if (res && res.imported_count > 0) {
        toastStore.success(
          'Assets Ingested',
          `Successfully imported ${res.imported_count} character(s)/stage(s) into Vault!`
        );
        await loadVaults();
        await loadAssets(vaultId);
      } else {
        toastStore.warning('Import Warning', 'No compatible fighters or stages found in archive(s)');
      }

      return res as IngestResult;
    } catch (err: any) {
      update((s) => ({ ...s, isIngesting: false, ingestMessage: '' }));
      console.error('Ingest error:', err);
      toastStore.error('Import Failed', err?.message || 'Could not process asset archive(s)');
      return null;
    }
  }

  async function linkToProject(projectDir: string, vaultId: string, assetKey: string): Promise<boolean> {
    try {
      await AddVaultAssetToProject(projectDir, vaultId, assetKey);
      toastStore.success('Added to Project', `Connected ${assetKey.split('/').pop()} to project roster`);
      return true;
    } catch (err: any) {
      toastStore.error('Link Failed', err?.message || 'Could not link asset to project');
      return false;
    }
  }

  async function unlinkFromProject(projectDir: string, assetKey: string): Promise<boolean> {
    try {
      await RemoveVaultAssetFromProject(projectDir, assetKey);
      toastStore.info('Removed from Project', `Unlinked ${assetKey.split('/').pop()} from project`);
      return true;
    } catch (err: any) {
      toastStore.error('Removal Error', err?.message || 'Could not remove asset from project');
      return false;
    }
  }

  async function browseArchive(): Promise<string> {
    try {
      return await SelectArchiveDialog();
    } catch {
      return '';
    }
  }

  async function browseArchives(): Promise<string[]> {
    try {
      return await SelectMultipleArchivesDialog();
    } catch {
      return [];
    }
  }

  async function browseFolder(): Promise<string> {
    try {
      return await SelectDirectoryDialog('Select Asset Folder or Vault');
    } catch {
      return '';
    }
  }

  function setSelectedAsset(asset: VaultAsset | null) {
    update((s) => ({ ...s, selectedAsset: asset }));
  }

  function setSearchQuery(query: string) {
    update((s) => ({ ...s, searchQuery: query }));
  }

  function setCategory(cat: AssetCategory | 'all') {
    update((s) => ({ ...s, selectedCategory: cat }));
  }

  function toggleTag(tag: string) {
    update((s) => ({
      ...s,
      selectedTag: s.selectedTag === tag ? null : tag,
    }));
  }

  return {
    subscribe,
    initDropEvents,
    loadVaults,
    loadAssets,
    selectVault,
    create,
    register,
    unregister,
    updateAsset,
    deleteAsset,
    ingest,
    ingestMultiple,
    linkToProject,
    unlinkFromProject,
    browseArchive,
    browseArchives,
    browseFolder,
    setSelectedAsset,
    setSearchQuery,
    setCategory,
    toggleTag,
  };
}

export const vaultStore = createVaultStore();

// Derived store for filtered & searched assets
export const filteredAssets = derived(vaultStore, ($v) => {
  let list = $v.assets;

  // Category filter
  if ($v.selectedCategory !== 'all') {
    list = list.filter((a) => a.category === $v.selectedCategory);
  }

  // Tag filter
  if ($v.selectedTag) {
    list = list.filter((a) => a.tags && a.tags.includes($v.selectedTag!));
  }

  // Search query
  if ($v.searchQuery.trim()) {
    const q = $v.searchQuery.toLowerCase().trim();
    list = list.filter(
      (a) =>
        (a.display_name && a.display_name.toLowerCase().includes(q)) ||
        (a.author && a.author.toLowerCase().includes(q)) ||
        (a.key && a.key.toLowerCase().includes(q)) ||
        (a.tags && a.tags.some((t) => t.toLowerCase().includes(q)))
    );
  }

  return list;
});
