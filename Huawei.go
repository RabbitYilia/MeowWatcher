package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/xlab/at/sms"
	"io"
	"strconv"
	"strings"
)

func Huawei_INIT(DeviceName string) {
	Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["HWVersion"] = Huawei_GET(DeviceName, "HWVersion")
	Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Model"] = Huawei_GET(DeviceName, "Model")
	Huawei_SET(DeviceName, "CellNetworkRegister", "2")
	Huawei_SET(DeviceName, "TECharset", "UCS2")
	Huawei_SET(DeviceName, "MessageStorage", "\"ME\",\"ME\",\"ME\"")
	DeivceSendCommand(DeviceName, "AT^SYSCFG=2,2,3FFFFFFF,2,4")
	Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["PhoneNumber"] = Huawei_GET(DeviceName, "PhoneNumber")
	Huawei_Status_Update(DeviceName)
}

func Huawei_Status_Update(DeviceName string) {
	CellNetworkRegisterStatus := Huawei_GET(DeviceName, "CellNetworkRegisterStatus")
	switch CellNetworkRegisterStatus {
	case "Home":
		Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["RegisterStatus"] = "Home"
	case "Romaning":
		Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["RegisterStatus"] = "Romaning"
		DebugOutput(0, DeviceName, "Romaning")
	case "Denied":
		Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["RegisterStatus"] = "Denied"
		DeviceStop(DeviceName, "Device Network Denied")
	case "No":
		if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["RegisterStatus"] != nil {
			if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["RegisterStatus"].(string) != "No" {
				DeviceError(DeviceName, errors.New("Device Comes to Offline"))
				break
			}
		}
		Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["RegisterStatus"] = "No"
		DeviceError(DeviceName, errors.New("Device Not Registered"))
	}
	OperatorName := Huawei_GET(DeviceName, "GetOperatorName")
	Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["OperatorName"] = OperatorName
	OperationMode := Huawei_GET(DeviceName, "GetOperationMode")
	if OperationMode == "NONE" {
		DeviceError(DeviceName, errors.New("OperationMode = NO SERVICE"))
	}
	Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["OperationMode"] = OperationMode
	//OperationStatus := Huawei_GET(DeviceName, "OperationStatus")
	//if OperationStatus != "Online" {
	//	DeviceError(DeviceName, errors.New("OperationStatus = "+OperationStatus))
	//}
}

func Huawei_GET(DeviceName string, Key string) string {
	MDMHandler := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MDMPortHandler"].(io.ReadWriteCloser)
	switch Key {
	case "HWVersion":
		HWVersion, _, err := SendCommand(MDMHandler, "AT+CGMR")
		if err != nil {
			DeviceError(DeviceName, err)
		}
		if HWVersion == "" {
			DeviceError(DeviceName, errors.New("Illegal Response"))
		}
		return strings.Replace(HWVersion, "+CGMR: ", "", -1)
	case "Model":
		Model, _, err := SendCommand(MDMHandler, "AT+CGMM")
		if err != nil {
			DeviceError(DeviceName, err)
		}
		if Model == "" {
			DeviceError(DeviceName, errors.New("Illegal Response"))
		}
		return strings.Replace(Model, "+CGMM: ", "", -1)
	case "CellNetworkRegisterStatus":
		CellNetworkRegister, _, err := SendCommand(MDMHandler, "AT+CREG?")
		if err != nil {
			DeviceError(DeviceName, err)
		}
		if CellNetworkRegister == "" {
			DeviceError(DeviceName, errors.New("Illegal Response"))
		}
		CellNetworkRegister = strings.Replace(CellNetworkRegister, "+CREG: ", "", -1)
		CellNetworkRegister = strings.Split(CellNetworkRegister, ",")[1]
		switch CellNetworkRegister {
		case "1":
			return "Home"
		case "5":
			return "Romaning"
		case "3":
			return "Denied"
		default:
			return "No"
		}
	case "GetOperatorName":
		OperatorName, _, err := SendCommand(MDMHandler, "AT+COPS?")
		if err != nil {
			DeviceError(DeviceName, err)
		}
		if OperatorName == "" {
			DeviceError(DeviceName, errors.New("Illegal Response"))
		}
		OperatorName = strings.Replace(OperatorName, "+COPS: ", "", -1)
		OperatorName = strings.Split(OperatorName, ",")[2]
		OperatorName = strings.Replace(OperatorName, "\"", "", -1)
		return OperatorName
	case "GetOperationMode":
		OperationMode, _, err := SendCommand(MDMHandler, "AT^GETPORTMODE")
		if err != nil {
			DeviceError(DeviceName, err)
		}
		if OperationMode == "" {
			DeviceError(DeviceName, errors.New("Illegal Response"))
		}
		OperationMode = strings.Replace(OperationMode, "^GETPORTMODE: ", "", -1)
		OperationMode = strings.Split(OperationMode, ",")[0]
		OperationMode = strings.Split(OperationMode, ":")[2]
		OperationMode = strings.Replace(OperationMode, "\"", "", -1)
		return OperationMode
	case "SMSStatus":
		SMSStatus, _, err := SendCommand(MDMHandler, "AT+CPMS?")
		if err != nil {
			DeviceError(DeviceName, err)
		}
		if SMSStatus == "" {
			DeviceError(DeviceName, errors.New("Illegal Response"))
		}
		SMSStatus = strings.Replace(SMSStatus, "+CPMS: ", "", -1)
		return SMSStatus
	case "MessageFormat":
		MessageFormat, _, err := SendCommand(MDMHandler, "AT+CMGF?")
		if err != nil {
			DeviceError(DeviceName, err)
		}
		if MessageFormat == "" {
			DeviceError(DeviceName, errors.New("Illegal Response"))
		}
		MessageFormat = strings.Replace(MessageFormat, "+CMGF: ", "", -1)
		switch MessageFormat {
		case "0":
			return "PDU"
		case "1":
			return "TEXT"
		}
	case "PhoneNumber":
		PhoneNumber, _, err := SendCommand(MDMHandler, "AT+CNUM")
		if err != nil {
			DeviceError(DeviceName, err)
		}
		if PhoneNumber == "" {
			DeviceError(DeviceName, errors.New("Illegal Response"))
		}
		PhoneNumber = strings.Replace(PhoneNumber, "+CNUM: ", "", -1)
		PhoneNumber = strings.Split(PhoneNumber, ",")[1]
		PhoneNumber = strings.Replace(PhoneNumber, "\"", "", -1)
		return PhoneNumber
	}
	return ""
}

func Huawei_SET(DeviceName string, Key string, Value string) string {
	MDMHandler := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MDMPortHandler"].(io.ReadWriteCloser)
	switch Key {
	case "CellNetworkRegister":
		_, _, err := SendCommand(MDMHandler, fmt.Sprintf("AT+CREG=%s", Value))
		if err != nil {
			DeviceError(DeviceName, err)
		}
	case "TECharset":
		_, _, err := SendCommand(MDMHandler, fmt.Sprintf("AT+CSCS=\"%s\"", Value))
		if err != nil {
			DeviceError(DeviceName, err)
		}
	case "MessageStorage":
		_, _, err := SendCommand(MDMHandler, fmt.Sprintf("AT+CPMS=%s", Value))
		if err != nil {
			DeviceError(DeviceName, err)
		}
	case "MessageFormat":
		_, _, err := SendCommand(MDMHandler, fmt.Sprintf("AT+CMGF=%s", Value))
		if err != nil {
			DeviceError(DeviceName, err)
		}
	case "ReadMessage":
		SMSResponse, CmdStatus, err := SendCommand(MDMHandler, fmt.Sprintf("AT+CMGR=%s", Value))
		if err != nil {
			DeviceError(DeviceName, err)
		}
		if CmdStatus != "OK" {
			DeviceError(DeviceName, errors.New("Illegal Response"))
		}
		SMSResponse = strings.Replace(SMSResponse, "+CMGR: ", "", -1)
		return SMSResponse
	case "DeleteMessage":
		_, _, err := SendCommand(MDMHandler, fmt.Sprintf("AT+CMGD=%s", Value))
		if err != nil {
			DeviceError(DeviceName, err)
		}
	}
	return ""
}

func Huawei_Get_SMS(DeviceName string) {
	OperationMode := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["OperationMode"].(string)
	Huawei_SET(DeviceName, "MessageFormat", "0")
	for {
		SMSStatus := Huawei_GET(DeviceName, "SMSStatus")
		SMSTotal, _ := strconv.Atoi(strings.Split(SMSStatus, ",")[1])
		Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MessageFormat"] = Huawei_GET(DeviceName, "MessageFormat")
		if SMSTotal != 0 {
			switch OperationMode {
			default:
				Huawei_Get_SMS_Common(DeviceName)
			case "CDMA":
				Huawei_Get_SMS_CDMA(DeviceName)
			case "EV-DO":
				Huawei_Get_SMS_CDMA(DeviceName)
			}
		} else {
			break
		}
	}
}

func Huawei_Get_SMS_Common(DeviceName string) {
	count := -1
	for {
		count++
		SMSResponse := Huawei_SET(DeviceName, "ReadMessage", strconv.Itoa(count))
		if SMSResponse == "" {
			continue
		}
		PDU := strings.Split(SMSResponse, "\r\n")[1]
		DecodePDU(DeviceName, PDU)
		Huawei_SET(DeviceName, "DeleteMessage", strconv.Itoa(count))
		break
	}
}
func Huawei_Get_SMS_CDMA(DeviceName string) {

}

func Huawei_SEND_SMS(DeviceName string, DstPhone string, Content string) error {
	SMS := sms.Message{
		Text:     Content,
		Encoding: sms.Encodings.UCS2,
		Type:     sms.MessageTypes.Submit,
		Address:  sms.PhoneNumber(DstPhone),
	}
	n, PDU, err := SMS.PDU()
	hexPDU := strings.Replace(hex.EncodeToString(PDU), "00010005a1", "0001000581", 1)
	if err != nil {
		return nil
	}
	DeivceSendCommand(DeviceName, "AT+CMGS="+strconv.Itoa(n))
	Response, _, _ := DeivceSendPDU(DeviceName, hexPDU)
	if strings.Contains(Response, "+CMGS:") {
		return nil
	} else {
		return errors.New("Send Failed")
	}
}
