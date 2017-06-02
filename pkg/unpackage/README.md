Plugin Unpackaging
==================
This package is used to unpackage plugins loaded in the [ACI](https://github.com/appc/spec/blob/master/spec/aci.md) format created by CoreOS. 

Currently we do not have a Pulse plugin packager, but it is a planned future work. 

## Overview  
1. The ACI file is checked (`GetContentType`) if it's the format of gzip. If so, the ACI is unzipped (`Uncompress`). Other compressed extensions
 may be added later.  
2. It is untarred (`Untar`) as underlying ACI files are tars. Untar creates a directory with the same name as the ACI without the extension and extracts all files and directories into it.  

	```
	/plugin-aci-name/
	/plugin-aci-name/manifest
	/plugin-aci-name/rootfs
	/plugin-aci-name/rootfs/plugin-binary
	```
3. The manifest is then unmars
haled into a struct (`UnmarshalJSON`). 

	```go
	type Labels struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	type App struct {
		Exec  []string `json:"exec"`
		Group int64    `json:"group,string"`
		User  int64    `json:"user,string"`
	}

	type Manifest struct {
		AcKind    string   `json:"acKind"`
		AcVersion string   `json:"acVersion"`
		Name      string   `json:"name"`
		Labels    []Labels `json:"labels"`
		App       App      `json:"app"`
	}
	```
