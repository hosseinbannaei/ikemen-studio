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

export type ThemeId = 'mkx' | 'obsidian' | 'cyber' | 'capcom' | 'light';
export type RadiusStyle = 'sharp' | 'subtle' | 'rounded';

export interface ThemePreset {
  id: ThemeId;
  name: string;
  subtitle: string;
  description: string;
  accentHex: string;
  bgHex: string;
  cardHex: string;
  borderHex: string;
  tag: string;
}

export interface Settings {
  enginesDir: string;
  theme: string;
  radiusStyle?: RadiusStyle;
  recentProjects: string[];
  defaultChannel: string;
  registeredVaults?: string[];
  defaultLinkStrategy?: string;
}

export type AssetCategory = 'fighters' | 'stages' | 'motifs' | 'lifebars' | 'sounds' | 'fonts' | 'storyboards';
export type LinkStrategy = 'symlink' | 'hardlink' | 'copy';

export interface ProjectStageInfo {
  relative_path: string;
  display_name: string;
  author: string;
  version: string;
  bgm_path: string;
  preview_base64: string;
  is_extra_stage: boolean;
  assigned_characters: string[];
  xscale: number;
  yscale: number;
  zoffset: number;
  is_linked_from_vault: boolean;
}

export interface ProjectMotifInfo {
  key: string;
  display_name: string;
  author: string;
  version: string;
  resolution: string;
  grid_columns: number;
  grid_rows: number;
  total_slots: number;
  is_active: boolean;
  sprite_file: string;
  sound_file: string;
  preview_base64: string;
  is_linked_from_vault: boolean;
}

export interface ProjectLifebarInfo {
  key: string;
  display_name: string;
  author: string;
  version: string;
  is_active: boolean;
  sprite_file: string;
  sound_file: string;
  font_count: number;
  preview_base64: string;
  is_linked_from_vault: boolean;
}

export interface ProjectAudioInfo {
  relative_path: string;
  file_name: string;
  format: string;
  size_bytes: number;
  assigned_events: string[];
  assigned_stages: string[];
  is_linked_from_vault: boolean;
}

export interface ProjectFontInfo {
  relative_path: string;
  file_name: string;
  font_type: string;
  size_bytes: number;
  system_slot_mappings: string[];
  is_linked_from_vault: boolean;
}

export interface ProjectStoryboardInfo {
  relative_path: string;
  display_name: string;
  scene_count: number;
  bgm_path: string;
  sprite_file: string;
  assigned_slots: string[];
  preview_base64: string;
  is_linked_from_vault: boolean;
}

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

export interface VaultCleanReport {
  vault_id: string;
  removed_duplicates: number;
  cleaned_contaminations: number;
  regenerated_previews: number;
  pruned_missing: number;
  total_assets_now: number;
  details: string[];
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

export interface RosterCharacterSlot {
  index: number;
  type: 'character' | 'randomselect' | 'empty' | 'disabled';
  character?: string;
  display_name?: string;
  author?: string;
  portrait_base64?: string;
  home_stage?: string;
  music?: string;
  order?: number;
  include_in_arcade?: boolean;
  raw_line?: string;
}

export interface RosterGridInfo {
  rows: number;
  columns: number;
  wrapping: boolean;
  show_empty_boxes: boolean;
}

export interface RosterAvailableCharacter {
  name: string;
  display_name: string;
  author: string;
  portrait_base64: string;
  is_linked: boolean;
}

export interface ProjectRoster {
  grid: RosterGridInfo;
  slots: RosterCharacterSlot[];
  extra_stages: string[];
  available_characters: RosterAvailableCharacter[];
  available_stages: string[];
}

export interface FileSectionSummary {
  name: string;
  line_start: number;
  line_end: number;
  item_count: number;
}

export interface AnimActionSummary {
  action_no: number;
  description?: string;
  frame_count: number;
  total_ticks: number;
  has_loop: boolean;
  has_hitbox: boolean;
  has_hurtbox: boolean;
}

export interface CommandEntrySummary {
  name: string;
  command: string;
  time: number;
  buffer_time: number;
}

export interface StateDefSummary {
  state_no: number;
  name?: string;
  type: string;
  move_type: string;
  physics: string;
  anim: number;
  controller_count: number;
}

export interface FileInspectionResult {
  rel_path: string;
  file_type: string;
  category: string;
  display_name: string;
  total_lines: number;
  size_bytes: number;
  sections: FileSectionSummary[];
  key_values: Record<string, string>;
  anim_actions?: AnimActionSummary[];
  commands?: CommandEntrySummary[];
  state_defs?: StateDefSummary[];
  raw_content: string;
  is_editable: boolean;
  syntax_mode: 'ini' | 'zss' | 'lua' | 'glsl' | 'plain';
}



