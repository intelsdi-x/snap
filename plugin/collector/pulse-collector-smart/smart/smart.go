/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

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

package smart

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unicode"
	"unsafe"
)

const (
	hdio_drive_cmd        = 0x031f
	win_smart             = 0xb0
	smart_read_values     = 0xd0
	smart_read_thresholds = 0xd1
	smart_enable          = 0xd8
	nr_attributes         = 30
)

type AttributeFormat int

// Supported formats of attribute's raw data.
const (
	FormatDefault AttributeFormat = iota
	FormatTemperature
	FormatPLPF
	FormatFP1024
	FormatTTS
)

type Attribute struct {
	Name   string
	Format AttributeFormat
}

// Connects attribute ID with its label and format of raw data.
var AttributeMap = map[byte]Attribute{
	0x05: {"reallocatedsectors", FormatDefault},
	0x09: {"poweronhours", FormatDefault},
	0x0c: {"powercyclecount", FormatDefault},
	0xaa: {"availablereservedspace", FormatDefault},
	0xab: {"programfailcount", FormatDefault},
	0xac: {"erasefailcount", FormatDefault},
	0xae: {"unexpectedpowerloss", FormatDefault},
	0xaf: {"powerlossprotectionfailure", FormatPLPF},
	0xb7: {"satadownshifts", FormatDefault},
	0xb8: {"e2eerrors", FormatDefault},
	0xbb: {"uncorrectableerrors", FormatDefault},
	0xbe: {"casetemperature", FormatTemperature},
	0xc0: {"unsafeshutdowns", FormatDefault},
	0xc2: {"internaltemperature", FormatDefault},
	0xc5: {"pendingsectors", FormatDefault},
	0xc7: {"crcerrors", FormatDefault},
	0xe1: {"hostwrites", FormatDefault},
	0xe2: {"timedworkload/mediawear", FormatFP1024},
	0xe3: {"timedworkload/readpercent", FormatDefault},
	0xe4: {"timedworkload/time", FormatDefault},
	0xe8: {"reservedblocks", FormatDefault},
	0xe9: {"wearout", FormatDefault},
	0xeA: {"thermalthrottle", FormatTTS},
	0xf1: {"totallba/written", FormatDefault},
	0xf2: {"totallba/read", FormatDefault},
}

// Data format for single attribute.
type SmartValue struct {
	Id     byte
	Status int16
	Data   byte
	Vendor [8]byte
}

// Data format for smart binary data.
type SmartValues struct {
	Revision          int16
	Values            [nr_attributes]SmartValue
	OfflineStatus     byte
	Vendor1           byte
	OfflineTimeout    int16
	Vendor2           byte
	OfflineCapability byte
	SmartCapability   int16
	Reserved          [16]byte
	Vendor            [125]byte
	Checksum          byte
}

// ReadSmartData_ enables SMART on device and retrieves binary data from it.
// It returns data casted to appropriate Go structure.
func ReadSmartData_(device string, sysutilProvider SysutilProvider) (*SmartValues, error) {
	f, err := sysutilProvider.OpenDevice(device)
	if err != nil {
		return nil, errors.New(device + ": Can't open device")
	}
	defer f.Close()

	if err := enableSmart(f.Fd(), sysutilProvider); err != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", device, err))
	}

	buf := make([]byte, 4+512)
	buf[0] = win_smart
	buf[1] = 0
	buf[2] = smart_read_values
	buf[3] = 1

	if err := sysutilProvider.Ioctl(f.Fd(), hdio_drive_cmd, buf); err != nil {
		return nil, errors.New(fmt.Sprintf(
			"%s: S.M.A.R.T Reading failed, error = %v", device, err))
	}

	values := SmartValues{}
	smart_data := bytes.NewBuffer(buf[4:])
	binary.Read(smart_data, binary.LittleEndian, &values)

	return &values, nil
}

func enableSmart(fd uintptr, sysutilProvider SysutilProvider) error {
	buf := make([]byte, 4+512)
	buf[0] = win_smart
	buf[1] = 0
	buf[2] = smart_enable
	buf[3] = 0
	e := sysutilProvider.Ioctl(fd, hdio_drive_cmd, buf)
	if e != nil {
		return errors.New("Can't enable S.M.A.R.T")
	}
	return nil
}

// Parses 8 bytes of raw data in a way specific to this format.
// It returns map of values. Main value is accessible using empty string.
// Additional values are accessible using "/[additonal value]"
func (a AttributeFormat) ParseRaw(data [8]byte) map[string]interface{} {
	switch a {
	case FormatDefault:
		return map[string]interface{}{"": uint64(data[1]) + uint64(data[2])<<8 +
			uint64(data[3])<<16 + uint64(data[4])<<24 +
			uint64(data[5])<<32 + uint64(data[5])<<40}
	case FormatFP1024:
		return map[string]interface{}{"": float64(uint64(data[1])+uint64(data[2])<<8+
			uint64(data[3])<<16+uint64(data[4])<<24+
			uint64(data[5])<<32+uint64(data[5])<<40) / 1024}
	case FormatPLPF:
		return map[string]interface{}{"": uint64(data[1]) + uint64(data[2])<<8,
			"/sincelast": uint64(data[3]) + uint64(data[4])<<8,
			"/tests":     uint64(data[5]) + uint64(data[6])<<8}
	case FormatTemperature:
		return map[string]interface{}{"": uint64(data[1]) + uint64(data[2])<<8,
			"/min": uint64(data[3]), "/max": uint64(data[4]),
			"/overcounter": uint64(data[5]) + uint64(data[6])<<8,
		}
	case FormatTTS:
		return map[string]interface{}{"": uint64(data[1]),
			"/eventcount": uint64(data[2]) + uint64(data[3])<<8 +
				uint64(data[4])<<16 + uint64(data[5])<<24,
		}
	}

	return nil
}

// Introduced to make mocking possible. See ReadSmartData_.
var ReadSmartData = ReadSmartData_

// GetKeys returns list of keys that can be used to access parsed values
// of particular format.
func (a AttributeFormat) GetKeys() []string {
	ret := []string{}
	switch a {
	case FormatPLPF:
		ret = []string{"/sincelast", "/tests"}
	case FormatTemperature:
		ret = []string{"/min", "/max", "/overcounter"}
	case FormatTTS:
		ret = []string{"/eventcount"}
	}
	return append([]string{"", "/normalized"}, ret...)
}

// GetAttributes transforms smart data structure to map containing attributes'
// values. Main value is accessed using label.
// Additional values are accessed using "[label]/[additonal value]".
func (sv SmartValues) GetAttributes() map[string]interface{} {
	ret_val := map[string]interface{}{}
	for i := 0; i < nr_attributes; i++ {
		a, ok := AttributeMap[sv.Values[i].Id]
		if ok {
			ret_val[a.Name+"/normalized"] = sv.Values[i].Data
			attrib_content := a.Format.ParseRaw(sv.Values[i].Vendor)
			for k, v := range attrib_content {
				ret_val[a.Name+k] = v
			}
		}
	}
	return ret_val
}

// Returns list of keys that can be used to access all values of all formats.
// Which is cross product of attributes' label and (sub)value keys.
func ListAllKeys() []string {
	keys := []string{}
	for _, v := range AttributeMap {
		for _, f := range v.Format.GetKeys() {
			keys = append(keys, v.Name+f)
		}
	}

	return keys
}

// Represents OS abstraction layer, currently used for mocking.
type SysutilProvider interface {
	OpenDevice(device string) (*os.File, error)
	Ioctl(fd uintptr, cmd uint, buf []byte) error
	ListDevices() ([]string, error)
}

type sysutilProviderLinux struct {
}

func (s *sysutilProviderLinux) OpenDevice(device string) (*os.File, error) {
	f, err := os.OpenFile("/dev/"+device, os.O_RDWR, 0)
	return f, err
}

func (s *sysutilProviderLinux) Ioctl(fd uintptr, cmd uint, buf []byte) error {
	ptr := uintptr(unsafe.Pointer(&buf[0]))
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(cmd), ptr)

	if e != 0 {
		return e
	}

	return nil
}

func (s *sysutilProviderLinux) ListDevices() ([]string, error) {
	result := []string{}

	f, err := os.Open("/proc/partitions")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scan := bufio.NewScanner(f)
	for scan.Scan() {
		if len(scan.Text()) < 1 {
			continue
		}

		table := strings.Fields(scan.Text())
		if table[0] == "8" && strings.IndexFunc(table[3], unicode.IsDigit) < 0 {
			result = append(result, table[3])
		}

	}

	return result, nil

}

func NewSysutilProvider() SysutilProvider {
	return &sysutilProviderLinux{}
}
