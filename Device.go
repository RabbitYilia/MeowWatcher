package main

import (
	"fmt"
	"io"
)

func DeviceInit(DeviceName string) {
	switch Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Manufacture"].(string) {
	case "SIMCOM INCORPORATED":
		SIMCOM_INIT(DeviceName)
	}
	Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"] = "ON"
	DebugOutput(0, fmt.Sprintf("%s Online", DeviceName))
}

func DeviceStatusUpdate(DeviceName string) {

}

func GetManufacture(DeviceName string) {
	Manufacture, _, err := SendCommand(Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MDMPortHandler"].(io.ReadWriteCloser), "AT+CGMI")
	if err != nil {
		DeviceError(DeviceName, err)
	}
	Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Manufacture"] = Manufacture
}

func DeviceError(DeviceName string, err error) {
	DebugOutput(0, DeviceName, err)
	if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MDMPortHandler"] != nil {
		Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MDMPortHandler"].(io.ReadWriteCloser).Close()
		delete(Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{}), "MDMPortHandler")
	}
	Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"] = "Error"
}

func DeviceStop(DeviceName string, Output string) {
	DebugOutput(0, DeviceName, Output)
	if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MDMPortHandler"] != nil {
		Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MDMPortHandler"].(io.ReadWriteCloser).Close()
		delete(Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{}), "MDMPortHandler")
	}
	Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"] = "STOP"
}

func DeviceGetSMS(DeviceName string) {
	switch Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Manufacture"].(string) {
	case "SIMCOM INCORPORATED":
		SIMCOM_Get_SMS(DeviceName)
	}
}

func DeviceSendSMS(DeviceName string, DstPhone string, Content string)error {
	//switch Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Manufacture"].(string) {
	//case "SIMCOM INCORPORATED":
	//	return SIMCOM_SEND_SMS(DeviceName,DstPhone,Content)
	//}
	//return errors.New("Unknown")
	return SIMCOM_SEND_SMS(DeviceName,DstPhone,Content)
}
