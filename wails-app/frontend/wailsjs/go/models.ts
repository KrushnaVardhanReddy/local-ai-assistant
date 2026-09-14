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

