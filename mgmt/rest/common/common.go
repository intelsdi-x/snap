package common

import (
	"io/ioutil"
	"os"
	"path"
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/control"
)

func WriteFile(filename string, b []byte) (string, error) {
	// Create temporary directory
	dir, err := ioutil.TempDir(control.GetDefaultConfig().TempDirPath, "snap-plugin-")
	if err != nil {
		return "", err
	}

	f, err := os.Create(path.Join(dir, filename))
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
