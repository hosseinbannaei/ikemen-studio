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

export interface ToastMessage {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message?: string;
  duration?: number;
}
