package fileutils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"
)

// WriteFile creates a temporary directory for loading plugins
// Plugins loaded by the cli and from the auto-load directory go through this route of copying the plugin binaries to the temp dir and executing from temp
// WriteFile takes the name of the original file (fileName), path to the original file (filePath) and the content of the file (b)
// Returns temporary file path and error
func WriteFile(fileName, filePath string, b []byte) (string, error) {
	// Create temporary directory
	dir, err := ioutil.TempDir(filePath, "snap-plugin-")
	if err != nil {
		return "", err
	}

	f, err := os.Create(filepath.Join(dir, fileName))
	if err != nil {
		return "", err
	}
	// Close before load
	defer f.Close()

	n, err := f.Write(b)
	log.Debugf("wrote %v to %v", n, f.Name())
	if err != nil {
		return "", err
	}
	if runtime.GOOS != "windows" {
		err = f.Chmod(0700)
		if err != nil {
			return "", err
		}
	}
	return f.Name(), nil
}
