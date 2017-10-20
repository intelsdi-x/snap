/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package aci

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	specaci "github.com/appc/spec/aci"
	"github.com/appc/spec/schema"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrChmod - Error message for error changing file permission
	ErrChmod = errors.New("Error changing file permissions")
	// ErrCopyingFile - Error message for error copying file
	ErrCopyingFile = errors.New("Error copying file")
	// ErrCreatingFile - Error message for error creating file
	ErrCreatingFile = errors.New("Error creating file")
	// ErrCreatingSymLink - Error message for creating Symlink
	ErrCreatingSymLink = errors.New("Error creating symlink")
	// ErrMkdirAll - Error message for error making directory
	ErrMkdirAll = errors.New("Error making directory")
	// ErrNext - Error message for error interating through tar file
	ErrNext = errors.New("Error iterating through tar file")
	// ErrUntar - Error message for error untarring file
	ErrUntar = errors.New("Error untarring file")
)

var aciLogger = log.WithField("_module", "aci")

// Manifest returns the ImageManifest inside the ACI file
func Manifest(f io.ReadSeeker) (*schema.ImageManifest, error) {
	m, err := specaci.ManifestFromImage(f)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Extract expands the ACI file to a temporary directory, returning
// the directory path where the ACI was expanded or an error
func Extract(f io.ReadSeeker) (string, error) {
	fileMode := os.FileMode(0755)

	tr, err := specaci.NewCompressedTarReader(f)
	if err != nil {
		return "", err
	}
	defer tr.Close()

	// Extract archive to temporary directory
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	aciLogger.WithField("directory", dir).Debugf(
		"Extracting archive to temporary directory")
	for {
		hdr, err := tr.Reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("%v\n%v", ErrNext, err)
		}
		file := filepath.Join(dir, hdr.Name)

		switch hdr.Typeflag {
		case tar.TypeReg:
			w, err := os.Create(file)
			if err != nil {
				return "", fmt.Errorf("%v: %v\n%v", ErrCreatingFile, file, err)
			}
			_, err = io.Copy(w, tr)
			if err != nil {
				w.Close()
				return "", fmt.Errorf("%v: %v\n%v", ErrCopyingFile, file, err)
			}
			w.Close()
			err = os.Chmod(file, fileMode)
			if err != nil {
				return "", fmt.Errorf("%v: %v\n%v", ErrChmod, file, err)
			}
		case tar.TypeDir:
			err = os.MkdirAll(file, fileMode)
			if err != nil {
				return "", fmt.Errorf("%v: %v\n%v", ErrMkdirAll, file, err)
			}
		case tar.TypeSymlink:
			err := os.Symlink(
				filepath.Join(dir, filepath.Dir(hdr.Name), hdr.Linkname),
				filepath.Join(dir, hdr.Name))
			if err != nil {
				return "", fmt.Errorf("%v: name: %v Linkname: %v \n%v",
					ErrCreatingSymLink, hdr.Name, hdr.Linkname, err)
			}
		default:
			return "", fmt.Errorf("%v (type: %d): %v", ErrUntar, hdr.Typeflag, hdr.Name)
		}
	}
	return dir, nil
}

// Validate makes sure the archive is valid. Otherwise,
// an error is returned
func Validate(f io.ReadSeeker) error {
	tr, err := specaci.NewCompressedTarReader(f)
	defer tr.Close()
	if err != nil {
		return err
	}

	if err := specaci.ValidateArchive(tr.Reader); err != nil {
		return err
	}
	return nil
}
