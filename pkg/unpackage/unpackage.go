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

package unpackage

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
)

var (
	ErrOpen           = errors.New("Error opening file")
	ErrRead           = errors.New("Error reading file")
	ErrSeek           = errors.New("Error seeking file")
	ErrCreatingFile   = errors.New("Error creating file")
	ErrCopyingFile    = errors.New("Error copying file")
	ErrUnzip          = errors.New("Error unzipping file")
	ErrReadManifest   = errors.New("Error reading manifest")
	ErrMkdirAll       = errors.New("Error making directory")
	ErrChmod          = errors.New("Error changing file permissions")
	ErrFileRead       = errors.New("Error reading file")
	ErrNoExecFiles    = errors.New("Error no executable files found")
	ErrNext           = errors.New("Error iterating through tar file")
	ErrGetContentType = errors.New("Error getting content type")
	ErrUnmarshalJSON  = errors.New("Error unmarshaling JSON")
	ErrUntar          = errors.New("Error untarring file")
)

var upkgLogger *log.Entry = log.WithFields(log.Fields{
	"_module": "unpackage",
})

//If package is an ACI, Untar package and Unmarshal ACI manifest
func Unpackager(path string) (string, *Manifest, error) {
	upkgLogger = upkgLogger.WithField("_block", "Unpackager")

	//Check if file has the extension .aci
	if !strings.Contains(path, ".aci") {
		return path, nil, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return "", nil, fmt.Errorf("%v: %v\n%v", ErrOpen, path, err)
	}
	defer f.Close()

	//Untar loaded package
	data, err := Uncompress(path, f)
	if err != nil {
		return "", nil, err
	}
	//Read JSON manifest
	manifest, err := UnmarshalJSON(data)
	if err != nil {
		return "", nil, err
	}

	dir := strings.Replace(path, ".aci", "", -1) + "/rootfs"

	//Get executable file name and update path
	if len(manifest.App.Exec) < 1 {
		return "", nil, ErrNoExecFiles
	}
	exec := manifest.App.Exec[0]
	path = dir + exec

	return path, manifest, nil
}

func Uncompress(path string, f *os.File) ([]byte, error) {
	upkgLogger = upkgLogger.WithField("_block", "Uncompress")
	upkgLogger.Info("Uncompressing Files")

	fileType, err := GetContentType(path, f)
	if err != nil {
		return nil, err
	}
	if fileType == "application/x-gzip" {
		reader, err := Unzip(path, f)
		if err != nil {
			return nil, err
		}
		return Untar(path, reader)
	}
	return Untar(path, bufio.NewReader(f))
}

//Untar files and get manifest
func Untar(path string, reader *bufio.Reader) ([]byte, error) {
	upkgLogger = upkgLogger.WithField("_block", "Untar")
	upkgLogger.Info("Untarring Files")

	fileMode := os.FileMode(0755)
	mdata := []byte{} //manifest data
	tr := tar.NewReader(reader)
	dir := filepath.Join(strings.Replace(path, ".aci", "", -1), "/")
	err := os.MkdirAll(dir, fileMode)
	if err != nil {
		return nil, fmt.Errorf("%v: %v\n%v", ErrMkdirAll, dir, err)
	}
	upkgLogger.Info("Creating directory and writing files to: ", dir)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%v\n%v", ErrNext, err)
		}
		file := filepath.Join(dir, hdr.Name)

		//Check if header type is a file or directory
		switch hdr.Typeflag {
		case tar.TypeReg:
			//Create files
			fmt.Println("x", hdr.Name)
			w, err := os.Create(file)
			if err != nil {
				return nil, fmt.Errorf("%v: %v\n%v", ErrCreatingFile, file, err)
			}
			_, err = io.Copy(w, tr)
			if err != nil {
				return nil, fmt.Errorf("%v: %v\n%v", ErrCopyingFile, file, err)
			}
			err = os.Chmod(file, fileMode)
			if err != nil {
				return nil, fmt.Errorf("%v: %v\n%v", ErrChmod, file, err)
			}
			//Get manifest data for unmarshaling
			if hdr.Name == "manifest" {
				mdata, err = ioutil.ReadFile(file)
				if err != nil {
					return nil, fmt.Errorf("%v: %v\n%v", ErrReadManifest, file, err)
				}
			}
			w.Close()
		case tar.TypeDir:
			//Create directory
			fmt.Println("x", hdr.Name)
			err = os.MkdirAll(file, fileMode)
			if err != nil {
				return nil, fmt.Errorf("%v: %v\n%v", ErrMkdirAll, file, err)
			}
		default:
			return nil, fmt.Errorf("%v: %v", ErrUntar, hdr.Name)
		}
	}
	return mdata, nil
}

//Get the content type of the file
func GetContentType(path string, file *os.File) (string, error) {
	upkgLogger = upkgLogger.WithField("_block", "GetContentType")

	b := make([]byte, 512)
	_, err := file.Read(b)
	if err != nil {
		return "", fmt.Errorf("%v: %v\n%v", ErrRead, path, err)
	}
	fileType := http.DetectContentType(b)
	_, err = file.Seek(0, 0)
	if err != nil {
		return "", fmt.Errorf("%v: %v\n%v", ErrSeek, path, err)
	}
	return fileType, nil
}

func Unzip(path string, f *os.File) (*bufio.Reader, error) {
	upkgLogger = upkgLogger.WithField("_block", "Unzip")
	reader, err := gzip.NewReader(bufio.NewReader(f))
	if err != nil {
		return nil, fmt.Errorf("%v: %v\n%v", ErrUnzip, path, err)
	}
	return bufio.NewReader(reader), nil
}

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

//Unmarshal ACI manifest
func UnmarshalJSON(data []byte) (*Manifest, error) {
	upkgLogger = upkgLogger.WithField("_block", "UnmarshalJSON")
	upkgLogger.Info("Unmarshaling JSON")

	manifest := &Manifest{}
	err := json.Unmarshal(data, manifest)
	if err != nil {
		return nil, fmt.Errorf("%v\n%v", ErrUnmarshalJSON, err)
	}
	return manifest, nil
}
