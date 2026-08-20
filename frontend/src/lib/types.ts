export interface EngineConfig {
  version: string;
  channel: string;
}

export interface ProjectManifest {
  name: string;
  version: string;
  engine: EngineConfig;
  created_at: string;
  updated_at: string;
  author: string;
  path: string;
}

export interface ReleaseAsset {
  name: string;
  size: number;
  downloadUrl: string;
  os: string;
  arch: string;
}

export interface ReleaseInfo {
  tag: string;
  name: string;
  publishedAt: string;
  isPrerelease: boolean;
  body: string;
  htmlUrl: string;
  assets: ReleaseAsset[];
}

export interface InstalledEngine {
  version: string;
  path: string;
  executablePath: string;
  installedAt: string;
  channel: string;
  size: number;
}

export interface Settings {
  enginesDir: string;
  theme: string;
  recentProjects: string[];
  defaultChannel: string;
  registeredVaults?: string[];
  defaultLinkStrategy?: string;
}

export type AssetCategory = 'fighters' | 'stages' | 'motifs' | 'sounds';
export type LinkStrategy = 'symlink' | 'hardlink' | 'copy';

export interface VaultAsset {
  key: string;
  category: AssetCategory;
  display_name: string;
  author: string;
  version_date: string;
  mugen_version: string;
  ikemen_version: string;
  source_url: string;
  source_package: string;
  license: string;
  tags: string[];
  preview_image: string;
  preview_base64?: string;
  notes: string;
  size_bytes: number;
  added_at: string;
}

export interface VaultInfo {
  id: string;
  name: string;
  description: string;
  path: string;
  asset_count: number;
  size_bytes: number;
  is_default: boolean;
  created_at: string;
}

export interface AssetMetadataUpdate {
  display_name?: string;
  author?: string;
  source_url?: string;
  license?: string;
  tags?: string[];
  notes?: string;
}

export interface IngestResult {
  vault_id: string;
  detected_assets: VaultAsset[];
  is_multi_asset: boolean;
  source_package: string;
  imported_count: number;
  warnings: string[];
}

export interface DownloadProgress {
  version: string;
  percent: number;
  downloadedBytes: number;
  totalBytes: number;
  status: 'downloading' | 'extracting' | 'completed' | 'cancelled' | 'error';
  error?: string;
}

export type GameState = 'idle' | 'starting' | 'running' | 'stopping';

export interface VerificationReport {
  totalChecked: number;
  missingCount: number;
  repairedCount: number;
  repairedFiles: string[];
  logFilePath: string;
  success: boolean;
  errorMessage?: string;
  mode?: string;
}

export interface CrashDiagnosticInfo {
  projectDir: string;
  errorSummary: string;
  logFilePath: string;
  canRepair: boolean;
}

export interface ExistingGameInspection {
  isValid: boolean;
  detectedName: string;
  characterCount: number;
  stageCount: number;
  hasSelectDef: boolean;
  hasSystemDef: boolean;
  hasConfigIni: boolean;
  sourcePath: string;
  detectedEngineVersion?: string;
}

export interface ImportOptions {
  sourceDir: string;
  targetDir: string;
  projectName: string;
  engineVersion: string;
  engineChannel?: string;
  enginePath?: string;
  baselineEnginePath?: string;
  author?: string;
  mode?: 'rebuild' | 'diff_upgrade' | 'legacy_match';
  includeChars?: boolean;
  includeStages?: boolean;
  includeSound?: boolean;
  includeFonts?: boolean;
  includeRoster?: boolean;
  includeLegacySystem?: boolean;
}

export interface EngineBackupInfo {
  id: string;
  version: string;
  timestamp: string;
  path: string;
}

export interface ToastMessage {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message?: string;
  duration?: number;
}

export interface ConfigIssue {
  key: string;
  currentValue: string;
  suggestedValue: string;
  severity: 'error' | 'warning' | 'info';
  description: string;
}

export interface ConfigInspectionResult {
  isValid: boolean;
  hasLegacyRenderMode: boolean;
  issues: ConfigIssue[];
  totalKeys: number;
  configPath: string;
}

export interface CategoryDiff {
  category: string;
  title: string;
  description: string;
  itemCount: number;
  status: 'outdated' | 'missing' | 'clean' | 'custom';
  files: string[];
}

export interface ProjectDiffSummary {
  categories: CategoryDiff[];
  totalDiscrepancies: number;
  engineVersion: string;
}

export interface AssetSyncOptions {
  projectDir: string;
  engineDir?: string;
  syncStockChars?: boolean;
  syncStockStages?: boolean;
  syncScreenpack?: boolean;
  syncFonts?: boolean;
  syncSound?: boolean;
  syncRuntime?: boolean;
  resetConfig?: boolean;
}

export interface ProjectFightersAndStages {
  characters: string[];
  stages: string[];
}

