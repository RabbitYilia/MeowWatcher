package main

import (
	"fmt"
	"log"
	"syscall"
	"unsafe"

	"github.com/tarm/goserial"
	"golang.org/x/sys/windows"
)

func SeekOfflineDevice() {
	MDMguid, _ := classGuidsFromName("Modem")
	MDMDeviceSet, _ := MDMguid[0].getDevicesSet()
	Serialguid, _ := classGuidsFromName("Ports")
	SerialDeviceSet, _ := Serialguid[0].getDevicesSet()
	//USB\VID_12D1&PID_140B&MI_00\6&[2C7B652D]&5&0000
	//DeviceIDExp := regexp.MustCompile(`&([0-9A-Fa-f])*&`)
	//VIDExp := regexp.MustCompile("VID_[0-9A-Fa-f]*")
	//MIExp := regexp.MustCompile("MI_[0-9A-Fa-f]*")
	for i := 0; ; i++ {
		MDMdevice, err := MDMDeviceSet.getDeviceInfo(i)
		if err != nil {
			break
		}
		MDMPortName, _ := retrievePortNameFromDevInfo(MDMdevice)
		//MDMInstanseID, _ := MDMdevice.getInstanceID()
		//DeviceID := strings.Replace(DeviceIDExp.FindStringSubmatch(MDMInstanseID)[0], "&", "", -1)
		//VID := strings.Replace(VIDExp.FindStringSubmatch(MDMInstanseID)[0], "VID_", "", -1)
		MDMOnline := false
		for DeviceName, _ := range Config["Devices"].(map[string]interface{}) {
			if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MDMPortName"] == nil {
				continue
			}
			DeivceMDMPortName := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MDMPortName"].(string)
			DeviceStatus := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"].(string)
			if (DeviceStatus == "ON" || DeviceStatus == "READY" || DeviceStatus == "STOP") && DeivceMDMPortName == MDMPortName {
				MDMOnline = true
				break
			}
		}
		if MDMOnline == true {
			continue
		}

		SerialConfig := &serial.Config{Name: MDMPortName, Baud: 115200, ReadTimeout: 5 /*毫秒*/}
		Handler, err := serial.OpenPort(SerialConfig)
		if err != nil {
			log.Println(err)
			continue
		}
		_, _, err = SendCommand(Handler, "ATE1")
		if err != nil {
			log.Println(err)
			Handler.Close()
			continue
		}
		IMEI, _, err := SendCommand(Handler, "AT+CGSN")
		if err != nil {
			log.Println(err)
			Handler.Close()
			continue
		}
		for DeviceName, _ := range Config["Devices"].(map[string]interface{}) {
			DeviceIMEI := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["IMEI"].(string)
			if DeviceIMEI == IMEI {
				Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MDMPortName"] = MDMPortName
				Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MDMPortHandler"] = Handler
				GetManufacture(DeviceName)
				Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"] = "READY"
				MDMOnline = true
				DeviceInit(DeviceName)
			}
			if MDMOnline == false {
				Handler.Close()
			}
		}
		MDMDeviceSet.destroy()
		SerialDeviceSet.destroy()
	}
}

func retrievePortNameFromDevInfo(device *deviceInfo) (string, error) {
	h, err := device.openDevRegKey(0x00000001, 0, 0x00000001, 0x20019)
	if err != nil {
		return "", err
	}
	defer syscall.RegCloseKey(h)

	var name [1024]uint16
	nameP := (*byte)(unsafe.Pointer(&name[0]))
	nameSize := uint32(len(name) * 2)
	if err := syscall.RegQueryValueEx(h, syscall.StringToUTF16Ptr("PortName"), nil, nil, nameP, &nameSize); err != nil {
		return "", err
	}
	return syscall.UTF16ToString(name[:]), nil
}

type deviceInfo struct {
	set  devicesSet
	data devInfoData
}

func (set devicesSet) getDeviceInfo(index int) (*deviceInfo, error) {
	result := &deviceInfo{set: set}

	result.data.size = uint32(unsafe.Sizeof(result.data))
	err := setupDiEnumDeviceInfo(set, uint32(index), &result.data)
	return result, err
}
func (set devicesSet) destroy() {
	setupDiDestroyDeviceInfoList(set)
}

func (dev *deviceInfo) getInstanceID() (string, error) {
	n := uint32(0)
	setupDiGetDeviceInstanceId(dev.set, &dev.data, nil, 0, &n)
	buff := make([]uint16, n)
	if err := setupDiGetDeviceInstanceId(dev.set, &dev.data, unsafe.Pointer(&buff[0]), uint32(len(buff)), &n); err != nil {
		return "", err
	}
	return windows.UTF16ToString(buff[:]), nil
}

func (dev *deviceInfo) openDevRegKey(scope dicsScope, hwProfile uint32, keyType uint32, samDesired regsam) (syscall.Handle, error) {
	return setupDiOpenDevRegKey(dev.set, &dev.data, scope, hwProfile, keyType, samDesired)
}

func classGuidsFromName(className string) ([]guid, error) {
	// Determine the number of GUIDs for className
	n := uint32(0)
	if err := setupDiClassGuidsFromNameInternal(className, nil, 0, &n); err != nil {
		// ignore error: UIDs array size too small
	}
	res := make([]guid, n)
	err := setupDiClassGuidsFromNameInternal(className, &res[0], n, &n)
	return res, err
}

type guid struct {
	data1 uint32
	data2 uint16
	data3 uint16
	data4 [8]byte
}

func (g guid) String() string {
	return fmt.Sprintf("%08x-%04x-%04x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		g.data1, g.data2, g.data3,
		g.data4[0], g.data4[1], g.data4[2], g.data4[3],
		g.data4[4], g.data4[5], g.data4[6], g.data4[7])
}
func (g *guid) getDevicesSet() (devicesSet, error) {
	return setupDiGetClassDevs(g, nil, 0, 0x00000002)
}

type devicesSet syscall.Handle
type deviceProperty uint32
type dicsScope uint32
type regsam uint32

type devInfoData struct {
	size     uint32
	guid     guid
	devInst  uint32
	reserved uintptr
}

var _ unsafe.Pointer

var (
	modsetupapi                           = windows.NewLazySystemDLL("setupapi.dll")
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
