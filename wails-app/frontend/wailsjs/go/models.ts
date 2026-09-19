export namespace audio {
	
	export class AudioDevice {
	    id: number;
	    name: string;
	    isInput: boolean;
	    isLoopback: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AudioDevice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.isInput = source["isInput"];
	        this.isLoopback = source["isLoopback"];
	    }
	}

}

export namespace backend {
	
	export class CacheItem {
	    id: string;
	    question: string;
	    answer: string;
	    tokensSaved: number;
	
	    static createFrom(source: any = {}) {
	        return new CacheItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.question = source["question"];
	        this.answer = source["answer"];
	        this.tokensSaved = source["tokensSaved"];
	    }
	}

}

export namespace driving {
	
	export class FileNode {
	    id: string;
	    name: string;
	    path: string;
	    isDir: boolean;
	    extension: string;
	    size: number;
	    children?: FileNode[];
	
	    static createFrom(source: any = {}) {
	        return new FileNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.isDir = source["isDir"];
	        this.extension = source["extension"];
	        this.size = source["size"];
	        this.children = this.convertValues(source["children"], FileNode);
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
	export class WorkspaceDocument {
	    id: string;
	    path: string;
	    name: string;
	    content: string;
	    loadedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceDocument(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.path = source["path"];
	        this.name = source["name"];
	        this.content = source["content"];
	        this.loadedAt = source["loadedAt"];
	    }
	}

}

export namespace engine {
	
	export class SummaryRequest {
	    id: string;
	    prompt: string;
	
	    static createFrom(source: any = {}) {
	        return new SummaryRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.prompt = source["prompt"];
	    }
	}

}

