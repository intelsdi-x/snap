package ipmi

import (
	"fmt"
	"sync"
	"unsafe"
)

// #include "linux_inband.h"
import "C"

// Implements communication with openipmi driver on linux
type LinuxInband struct {
	Device string
	mutex  sync.Mutex
}

// Performs batch of requests to given device.
// nSim - number of request that are allowed to be 'in processing'.
// Returns array of responses in order corresponding to requests.
// Error is returned when any of requests failed.
func (al *LinuxInband) BatchExecRaw(requests []IpmiRequest, nSim int) ([]IpmiResponse, error) {
	al.mutex.Lock()
	defer al.mutex.Unlock()

	n := len(requests)
	info := C.struct_IpmiStatusInfo{}
	inputs := make([]C.struct_IpmiCommandInput, n)
	outputs := make([]C.struct_IpmiCommandOutput, n)

	for i, r := range requests {
		for j, b := range r.Data {
			inputs[i].data[j] = C.char(b)
		}
		inputs[i].data_len = C.int(len(r.Data))
		inputs[i].channel = C.short(r.Channel)
		inputs[i].slave = C.uchar(r.Slave)
	}

	errcode := C.IPMI_BatchCommands(C.CString(al.Device), &inputs[0], &outputs[0],
		C.int(n), C.int(nSim), &info)

	switch {
	case errcode < 0:
		return nil, fmt.Errorf("%d : Invalid call", errcode)
	case errcode > 0:
		return nil, fmt.Errorf("%d : System error [%d : %s]", errcode,
			info.system_error, C.GoString(&info.error_str[0]))
	}

	results := make([]IpmiResponse, n)

	for i, r := range outputs {
		results[i].Data = C.GoBytes(unsafe.Pointer(&r.data[0]), r.data_len)
	}

	return results, nil
}
