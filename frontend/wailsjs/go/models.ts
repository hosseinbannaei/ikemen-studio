export namespace config {
	
	export class ConfigIssue {
	    key: string;
	    currentValue: string;
	    suggestedValue: string;
	    severity: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigIssue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.currentValue = source["currentValue"];
	        this.suggestedValue = source["suggestedValue"];
	        this.severity = source["severity"];
	        this.description = source["description"];
	    }
	}
	export class ConfigInspectionResult {
	    isValid: boolean;
	    hasLegacyRenderMode: boolean;
	    issues: ConfigIssue[];
	    totalKeys: number;
	    configPath: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigInspectionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isValid = source["isValid"];
	        this.hasLegacyRenderMode = source["hasLegacyRenderMode"];
	        this.issues = this.convertValues(source["issues"], ConfigIssue);
	        this.totalKeys = source["totalKeys"];
	        this.configPath = source["configPath"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Settings {
	    enginesDir: string;
	    theme: string;
	    recentProjects: string[];
	    defaultChannel: string;
	    registeredVaults: string[];
	    defaultLinkStrategy: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enginesDir = source["enginesDir"];
	        this.theme = source["theme"];
	        this.recentProjects = source["recentProjects"];
	        this.defaultChannel = source["defaultChannel"];
	        this.registeredVaults = source["registeredVaults"];
	        this.defaultLinkStrategy = source["defaultLinkStrategy"];
	    }
	}

}

export namespace engine {
	
	export class InstalledEngine {
	    version: string;
	    path: string;
	    executablePath: string;
	    // Go type: time
	    installedAt: any;
	    channel: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new InstalledEngine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.path = source["path"];
	        this.executablePath = source["executablePath"];
	        this.installedAt = this.convertValues(source["installedAt"], null);
	        this.channel = source["channel"];
	        this.size = source["size"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ReleaseAsset {
	    name: string;
	    size: number;
	    downloadUrl: string;
	    os: string;
	    arch: string;
	
	    static createFrom(source: any = {}) {
	        return new ReleaseAsset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.size = source["size"];
	        this.downloadUrl = source["downloadUrl"];
	        this.os = source["os"];
	        this.arch = source["arch"];
	    }
	}
	export class ReleaseInfo {
	    tag: string;
	    name: string;
	    // Go type: time
	    publishedAt: any;
	    isPrerelease: boolean;
	    body: string;
	    htmlUrl: string;
	    assets: ReleaseAsset[];
	
	    static createFrom(source: any = {}) {
	        return new ReleaseInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tag = source["tag"];
	        this.name = source["name"];
	        this.publishedAt = this.convertValues(source["publishedAt"], null);
	        this.isPrerelease = source["isPrerelease"];
	        this.body = source["body"];
	        this.htmlUrl = source["htmlUrl"];
	        this.assets = this.convertValues(source["assets"], ReleaseAsset);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace main {
	
	export class ProjectFightersAndStages {
	    characters: string[];
	    stages: string[];
	
	    static createFrom(source: any = {}) {
	        return new ProjectFightersAndStages(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.characters = source["characters"];
	        this.stages = source["stages"];
	    }
	}

}

export namespace project {
	
	export class AssetSyncOptions {
	    projectDir: string;
	    engineDir: string;
	    syncStockChars: boolean;
	    syncStockStages: boolean;
	    syncScreenpack: boolean;
	    syncFonts: boolean;
	    syncSound: boolean;
	    syncRuntime: boolean;
	    resetConfig: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AssetSyncOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectDir = source["projectDir"];
	        this.engineDir = source["engineDir"];
	        this.syncStockChars = source["syncStockChars"];
	        this.syncStockStages = source["syncStockStages"];
	        this.syncScreenpack = source["syncScreenpack"];
	        this.syncFonts = source["syncFonts"];
	        this.syncSound = source["syncSound"];
	        this.syncRuntime = source["syncRuntime"];
	        this.resetConfig = source["resetConfig"];
	    }
	}
	export class CategoryDiff {
	    category: string;
	    title: string;
	    description: string;
	    itemCount: number;
	    status: string;
	    files: string[];
	
	    static createFrom(source: any = {}) {
	        return new CategoryDiff(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = source["category"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.itemCount = source["itemCount"];
	        this.status = source["status"];
	        this.files = source["files"];
	    }
	}
	export class EngineBackupInfo {
	    id: string;
	    version: string;
	    // Go type: time
	    timestamp: any;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new EngineBackupInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.version = source["version"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.path = source["path"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class EngineConfig {
	    version: string;
	    channel: string;
	
	    static createFrom(source: any = {}) {
	        return new EngineConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.channel = source["channel"];
	    }
	}
	export class ExistingGameInspection {
	    isValid: boolean;
	    detectedName: string;
	    characterCount: number;
	    stageCount: number;
	    hasSelectDef: boolean;
	    hasSystemDef: boolean;
	    hasConfigIni: boolean;
	    sourcePath: string;
	    detectedEngineVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new ExistingGameInspection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isValid = source["isValid"];
	        this.detectedName = source["detectedName"];
	        this.characterCount = source["characterCount"];
	        this.stageCount = source["stageCount"];
	        this.hasSelectDef = source["hasSelectDef"];
	        this.hasSystemDef = source["hasSystemDef"];
	        this.hasConfigIni = source["hasConfigIni"];
	        this.sourcePath = source["sourcePath"];
	        this.detectedEngineVersion = source["detectedEngineVersion"];
	    }
	}
	export class ImportOptions {
	    sourceDir: string;
	    targetDir: string;
	    projectName: string;
	    engineVersion: string;
	    engineChannel: string;
	    enginePath: string;
	    baselineEnginePath: string;
	    author: string;
	    mode: string;
	    includeChars: boolean;
	    includeStages: boolean;
	    includeSound: boolean;
	    includeFonts: boolean;
	    includeRoster: boolean;
	    includeLegacySystem: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ImportOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourceDir = source["sourceDir"];
	        this.targetDir = source["targetDir"];
	        this.projectName = source["projectName"];
	        this.engineVersion = source["engineVersion"];
	        this.engineChannel = source["engineChannel"];
	        this.enginePath = source["enginePath"];
	        this.baselineEnginePath = source["baselineEnginePath"];
	        this.author = source["author"];
	        this.mode = source["mode"];
	        this.includeChars = source["includeChars"];
	        this.includeStages = source["includeStages"];
	        this.includeSound = source["includeSound"];
	        this.includeFonts = source["includeFonts"];
	        this.includeRoster = source["includeRoster"];
	        this.includeLegacySystem = source["includeLegacySystem"];
	    }
	}
	export class ProjectDiffSummary {
	    categories: CategoryDiff[];
	    totalDiscrepancies: number;
	    engineVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectDiffSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.categories = this.convertValues(source["categories"], CategoryDiff);
	        this.totalDiscrepancies = source["totalDiscrepancies"];
	        this.engineVersion = source["engineVersion"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProjectManifest {
	    name: string;
	    version: string;
	    engine: EngineConfig;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	    author: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectManifest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.engine = this.convertValues(source["engine"], EngineConfig);
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.author = source["author"];
	        this.path = source["path"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RosterAvailableCharacter {
	    name: string;
	    display_name: string;
	    author: string;
	    portrait_base64: string;
	    is_linked: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RosterAvailableCharacter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.display_name = source["display_name"];
	        this.author = source["author"];
	        this.portrait_base64 = source["portrait_base64"];
	        this.is_linked = source["is_linked"];
	    }
	}
	export class RosterCharacterSlot {
	    index: number;
	    type: string;
	    character: string;
	    display_name: string;
	    author: string;
	    portrait_base64: string;
	    home_stage: string;
	    music: string;
	    order: number;
	    include_in_arcade: boolean;
	    raw_line: string;
	
	    static createFrom(source: any = {}) {
	        return new RosterCharacterSlot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.type = source["type"];
	        this.character = source["character"];
	        this.display_name = source["display_name"];
	        this.author = source["author"];
	        this.portrait_base64 = source["portrait_base64"];
	        this.home_stage = source["home_stage"];
	        this.music = source["music"];
	        this.order = source["order"];
	        this.include_in_arcade = source["include_in_arcade"];
	        this.raw_line = source["raw_line"];
	    }
	}
	export class RosterGridInfo {
	    rows: number;
	    columns: number;
	    wrapping: boolean;
	    show_empty_boxes: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RosterGridInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rows = source["rows"];
	        this.columns = source["columns"];
	        this.wrapping = source["wrapping"];
	        this.show_empty_boxes = source["show_empty_boxes"];
	    }
	}
	export class ProjectRoster {
	    grid: RosterGridInfo;
	    slots: RosterCharacterSlot[];
	    extra_stages: string[];
	    available_characters: RosterAvailableCharacter[];
	    available_stages: string[];
	
	    static createFrom(source: any = {}) {
	        return new ProjectRoster(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.grid = this.convertValues(source["grid"], RosterGridInfo);
	        this.slots = this.convertValues(source["slots"], RosterCharacterSlot);
	        this.extra_stages = source["extra_stages"];
	        this.available_characters = this.convertValues(source["available_characters"], RosterAvailableCharacter);
	        this.available_stages = source["available_stages"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class VerificationReport {
	    totalChecked: number;
	    missingCount: number;
	    repairedCount: number;
	    repairedFiles: string[];
	    logFilePath: string;
	    success: boolean;
	    errorMessage?: string;
	    mode: string;
	
	    static createFrom(source: any = {}) {
	        return new VerificationReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalChecked = source["totalChecked"];
	        this.missingCount = source["missingCount"];
	        this.repairedCount = source["repairedCount"];
	        this.repairedFiles = source["repairedFiles"];
	        this.logFilePath = source["logFilePath"];
	        this.success = source["success"];
	        this.errorMessage = source["errorMessage"];
	        this.mode = source["mode"];
	    }
	}

}

export namespace vault {
	
	export class AssetMetadataUpdate {
	    display_name: string;
	    author: string;
	    source_url: string;
	    license: string;
	    tags: string[];
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new AssetMetadataUpdate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.display_name = source["display_name"];
	        this.author = source["author"];
	        this.source_url = source["source_url"];
	        this.license = source["license"];
	        this.tags = source["tags"];
	        this.notes = source["notes"];
	    }
	}
	export class VaultAsset {
	    key: string;
	    category: string;
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
	    preview_base64: string;
	    notes: string;
	    size_bytes: number;
	    // Go type: time
	    added_at: any;
	
	    static createFrom(source: any = {}) {
	        return new VaultAsset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.category = source["category"];
	        this.display_name = source["display_name"];
	        this.author = source["author"];
	        this.version_date = source["version_date"];
	        this.mugen_version = source["mugen_version"];
	        this.ikemen_version = source["ikemen_version"];
	        this.source_url = source["source_url"];
	        this.source_package = source["source_package"];
	        this.license = source["license"];
	        this.tags = source["tags"];
	        this.preview_image = source["preview_image"];
	        this.preview_base64 = source["preview_base64"];
	        this.notes = source["notes"];
	        this.size_bytes = source["size_bytes"];
	        this.added_at = this.convertValues(source["added_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class IngestResult {
	    vault_id: string;
	    detected_assets: VaultAsset[];
	    is_multi_asset: boolean;
	    source_package: string;
	    imported_count: number;
	    warnings: string[];
	
	    static createFrom(source: any = {}) {
	        return new IngestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.vault_id = source["vault_id"];
	        this.detected_assets = this.convertValues(source["detected_assets"], VaultAsset);
	        this.is_multi_asset = source["is_multi_asset"];
	        this.source_package = source["source_package"];
	        this.imported_count = source["imported_count"];
	        this.warnings = source["warnings"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class VaultInfo {
	    id: string;
	    name: string;
	    description: string;
	    path: string;
	    asset_count: number;
	    size_bytes: number;
	    is_default: boolean;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new VaultInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.path = source["path"];
	        this.asset_count = source["asset_count"];
	        this.size_bytes = source["size_bytes"];
	        this.is_default = source["is_default"];
	        this.created_at = this.convertValues(source["created_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

