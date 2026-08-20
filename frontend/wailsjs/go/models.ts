export namespace config {
	
	export class Settings {
	    enginesDir: string;
	    theme: string;
	    recentProjects: string[];
	    defaultChannel: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enginesDir = source["enginesDir"];
	        this.theme = source["theme"];
	        this.recentProjects = source["recentProjects"];
	        this.defaultChannel = source["defaultChannel"];
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

export namespace project {
	
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
	export class VerificationReport {
	    totalChecked: number;
	    missingCount: number;
	    repairedCount: number;
	    repairedFiles: string[];
	    logFilePath: string;
	    success: boolean;
	    errorMessage?: string;
	
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
	    }
	}

}

