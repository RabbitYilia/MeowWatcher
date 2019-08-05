package enumerator

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var _ unsafe.Pointer

var (
	modsetupapi = windows.NewLazySystemDLL("setupapi.dll")

	procSetupDiClassGuidsFromNameW        = modsetupapi.NewProc("SetupDiClassGuidsFromNameW")
	procSetupDiGetClassDevsW              = modsetupapi.NewProc("SetupDiGetClassDevsW")
	procSetupDiDestroyDeviceInfoList      = modsetupapi.NewProc("SetupDiDestroyDeviceInfoList")
	procSetupDiEnumDeviceInfo             = modsetupapi.NewProc("SetupDiEnumDeviceInfo")
	procSetupDiGetDeviceInstanceIdW       = modsetupapi.NewProc("SetupDiGetDeviceInstanceIdW")
	procSetupDiOpenDevRegKey              = modsetupapi.NewProc("SetupDiOpenDevRegKey")
	procSetupDiGetDeviceRegistryPropertyW = modsetupapi.NewProc("SetupDiGetDeviceRegistryPropertyW")
)

func setupDiClassGuidsFromNameInternal(class string, guid *guid, guidSize uint32, requiredSize *uint32) (err error) {
	var _p0 *uint16
	_p0, err = syscall.UTF16PtrFromString(class)
	if err != nil {
		return
	}
	return _setupDiClassGuidsFromNameInternal(_p0, guid, guidSize, requiredSize)
}

func _setupDiClassGuidsFromNameInternal(class *uint16, guid *guid, guidSize uint32, requiredSize *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procSetupDiClassGuidsFromNameW.Addr(), 4, uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(guid)), uintptr(guidSize), uintptr(unsafe.Pointer(requiredSize)), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func setupDiGetClassDevs(guid *guid, enumerator *string, hwndParent uintptr, flags uint32) (set devicesSet, err error) {
	r0, _, e1 := syscall.Syscall6(procSetupDiGetClassDevsW.Addr(), 4, uintptr(unsafe.Pointer(guid)), uintptr(unsafe.Pointer(enumerator)), uintptr(hwndParent), uintptr(flags), 0, 0)
	set = devicesSet(r0)
	if set == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func setupDiDestroyDeviceInfoList(set devicesSet) (err error) {
	r1, _, e1 := syscall.Syscall(procSetupDiDestroyDeviceInfoList.Addr(), 1, uintptr(set), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func setupDiEnumDeviceInfo(set devicesSet, index uint32, info *devInfoData) (err error) {
	r1, _, e1 := syscall.Syscall(procSetupDiEnumDeviceInfo.Addr(), 3, uintptr(set), uintptr(index), uintptr(unsafe.Pointer(info)))
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func setupDiGetDeviceInstanceId(set devicesSet, devInfo *devInfoData, devInstanceId unsafe.Pointer, devInstanceIdSize uint32, requiredSize *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procSetupDiGetDeviceInstanceIdW.Addr(), 5, uintptr(set), uintptr(unsafe.Pointer(devInfo)), uintptr(devInstanceId), uintptr(devInstanceIdSize), uintptr(unsafe.Pointer(requiredSize)), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func setupDiOpenDevRegKey(set devicesSet, devInfo *devInfoData, scope dicsScope, hwProfile uint32, keyType uint32, samDesired regsam) (hkey syscall.Handle, err error) {
	r0, _, e1 := syscall.Syscall6(procSetupDiOpenDevRegKey.Addr(), 6, uintptr(set), uintptr(unsafe.Pointer(devInfo)), uintptr(scope), uintptr(hwProfile), uintptr(keyType), uintptr(samDesired))
	hkey = syscall.Handle(r0)
	if hkey == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func setupDiGetDeviceRegistryProperty(set devicesSet, devInfo *devInfoData, property deviceProperty, propertyType *uint32, outValue *byte, outSize *uint32, reqSize *uint32) (res bool) {
	r0, _, _ := syscall.Syscall9(procSetupDiGetDeviceRegistryPropertyW.Addr(), 7, uintptr(set), uintptr(unsafe.Pointer(devInfo)), uintptr(property), uintptr(unsafe.Pointer(propertyType)), uintptr(unsafe.Pointer(outValue)), uintptr(unsafe.Pointer(outSize)), uintptr(unsafe.Pointer(reqSize)), 0, 0)
	res = r0 != 0
	return
}
