package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/xlab/at/sms"
	"io"
	"strconv"
	"strings"
	"time"
)

func Huawei_INIT(DeviceName string) error {
	var err error
	if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"] != nil {
		Status := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"].(string)
		if Status != "ON" && Status != "READY" {
			return errors.New("Device Not READY")
		}
	}
	OperationMode, err := Huawei_GET(DeviceName, "GetOperationMode")
	if err != nil {
		return err
	}
	if OperationMode == "NONE" {
		return errors.New("OperationMode = NO SERVICE")
	}
	Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["OperationMode"] = OperationMode
	Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["HWVersion"], err = Huawei_GET(DeviceName, "HWVersion")
	if err != nil {
		return err
	}
	Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Model"], err = Huawei_GET(DeviceName, "Model")
	if err != nil {
		return err
	}
	if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["OperationMode"].(string) != "EV-DO" {
		_, err = Huawei_SET(DeviceName, "CellNetworkRegister", "2")
		if err != nil {
			return err
		}
		_, err = Huawei_SET(DeviceName, "TECharset", "UCS2")
		if err != nil {
			return err
		}
		_, _, err = DeivceSendCommand(DeviceName, "AT^SYSCFG=2,2,3FFFFFFF,2,4")
		if err != nil {
			return err
		}
		Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["PhoneNumber"], err = Huawei_GET(DeviceName, "PhoneNumber")
		if err != nil {
			return err
		}
	}
	_, err = Huawei_SET(DeviceName, "MessageStorage", "\"ME\",\"ME\",\"ME\"")
	if err != nil {
		return err
	}
	err = Huawei_Status_Update(DeviceName)
	if err != nil {
		return err
	}
	return nil
}

func Huawei_Status_Update(DeviceName string) error {
	if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"] != nil {
		Status := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"].(string)
		if Status != "ON" && Status != "READY" {
			return errors.New("Device Not READY")
		}
	}
	if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["OperationMode"].(string) != "EV-DO" {
		CellNetworkRegisterStatus, err := Huawei_GET(DeviceName, "CellNetworkRegisterStatus")
		if err != nil {
			return err
		}
		switch CellNetworkRegisterStatus {
		case "Home":
			Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["RegisterStatus"] = "Home"
		case "Romaning":
			Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["RegisterStatus"] = "Romaning"
			DebugOutput(0, DeviceName, "Romaning")
		case "Denied":
			Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["RegisterStatus"] = "Denied"
			DeviceStop(DeviceName, "Device Network Denied")
			return errors.New("Device Stop")
		case "No":
			if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["RegisterStatus"] != nil {
				if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["RegisterStatus"].(string) != "No" {
					return errors.New("Device Comes to Offline")
				}
			}
			Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["RegisterStatus"] = "No"
			return errors.New("Device Not Registered")
		}
		OperatorName, err := Huawei_GET(DeviceName, "GetOperatorName")
		if err != nil {
			return err
		}
		Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["OperatorName"] = OperatorName
	}
	OperationMode, err := Huawei_GET(DeviceName, "GetOperationMode")
	if err != nil {
		return err
	}
	if OperationMode == "NONE" {
		return errors.New("OperationMode = NO SERVICE")
	}
	Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["OperationMode"] = OperationMode
	//OperationStatus := Huawei_GET(DeviceName, "OperationStatus")
	//if OperationStatus != "Online" {
	//	DeviceError(DeviceName, errors.New("OperationStatus = "+OperationStatus))
	//}
	return nil
}

func Huawei_GET(DeviceName string, Key string) (string, error) {
	if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"] != nil {
		Status := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"].(string)
		if Status != "ON" && Status != "READY" {
			return "", errors.New("Device Not Ready")
		}
	}
	if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MDMPortHandler"] == nil {
		return "", errors.New("Device Not Ready")
	}
	MDMHandler := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MDMPortHandler"].(io.ReadWriteCloser)
	switch Key {
	case "HWVersion":
		//11.126.10.01.68
		HWVersion, _, err := SendCommand(MDMHandler, "AT+CGMR")
		if err != nil {
			return "", err
		}
		if HWVersion == "" {
			return "", errors.New("Illegal Response")
		}
		return strings.Replace(HWVersion, "+CGMR: ", "", -1), nil
	case "Model":
		//E180
		Model, _, err := SendCommand(MDMHandler, "AT+CGMM")
		if err != nil {
			return "", err
		}
		if Model == "" {
			//return "", errors.New("Illegal Response")
			//Some Huawei device do not provide Model info
			return "", nil
		}
		return strings.Replace(Model, "+CGMM: ", "", -1), nil
	case "CellNetworkRegisterStatus":
		//+CREG: 0,1
		CellNetworkRegister, _, err := SendCommand(MDMHandler, "AT+CREG?")
		if err != nil {
			return "", err
		}
		if CellNetworkRegister == "" {
			return "", errors.New("Illegal Response")
		}
		CellNetworkRegister = strings.Replace(CellNetworkRegister, "+CREG: ", "", -1)
		CellNetworkRegister = strings.Split(CellNetworkRegister, ",")[1]
		switch CellNetworkRegister {
		case "1":
			return "Home", nil
		case "5":
			return "Romaning", nil
		case "3":
			return "Denied", nil
		default:
			return "No", nil
		}
	case "GetOperatorName":
		//+COPS: 0,2,"46001",2
		OperatorName, _, err := SendCommand(MDMHandler, "AT+COPS?")
		if err != nil {
			return "", err
		}
		if OperatorName == "" {
			return "", errors.New("Illegal Response")
		}
		OperatorName = strings.Replace(OperatorName, "+COPS: ", "", -1)
		OperatorName = strings.Split(OperatorName, ",")[2]
		OperatorName = strings.Replace(OperatorName, "\"", "", -1)
		return OperatorName, nil
	case "GetOperationMode":
		//^GETPORTMODE:TYPE:WCDMA:Qualcomm,MDM:0,DIAG:1,PCUI:2
		//^GETPORTMODE:TYPE:EV-DO:Qualcomm,MDM:0,DIAG:1,PCUI:2
		OperationMode, _, err := SendCommand(MDMHandler, "AT^GETPORTMODE")
		if err != nil {
			return "", err
		}
		if OperationMode == "" {
			return "", errors.New("Illegal Response")
		}
		OperationMode = strings.Replace(OperationMode, "^GETPORTMODE: ", "", -1)
		OperationMode = strings.Split(OperationMode, ",")[0]
		OperationMode = strings.Split(OperationMode, ":")[2]
		OperationMode = strings.Replace(OperationMode, "\"", "", -1)
		return OperationMode, nil
	case "SMSStatus":
		//+CPMS: 0,23,0,23,0,23
		SMSStatus, _, err := SendCommand(MDMHandler, "AT+CPMS?")
		if err != nil {
			return "", err
		}
		if SMSStatus == "" {
			return "", errors.New("Illegal Response")
		}
		SMSStatus = strings.Replace(SMSStatus, "+CPMS: ", "", -1)
		return SMSStatus, nil
	case "MessageFormat":
		//+CMGF: 0
		MessageFormat, _, err := SendCommand(MDMHandler, "AT+CMGF?")
		if err != nil {
			return "", err
		}
		if MessageFormat == "" {
			return "", errors.New("Illegal Response")
		}
		MessageFormat = strings.Replace(MessageFormat, "+CMGF: ", "", -1)
		switch MessageFormat {
		case "0":
			return "PDU", nil
		case "1":
			return "TEXT", nil
		}
	case "PhoneNumber":
		//+CNUM: "","+8613800138000",145
		PhoneNumber, _, err := SendCommand(MDMHandler, "AT+CNUM")
		if err != nil {
			return "", err
		}
		if PhoneNumber == "" {
			return "N/A", nil
		}
		PhoneNumber = strings.Replace(PhoneNumber, "+CNUM: ", "", -1)
		PhoneNumber = strings.Split(PhoneNumber, ",")[1]
		PhoneNumber = strings.Replace(PhoneNumber, "\"", "", -1)
		return PhoneNumber, nil
	}
	return "", nil
}

func Huawei_SET(DeviceName string, Key string, Value string) (string, error) {
	if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"] != nil {
		Status := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"].(string)
		if Status != "ON" && Status != "READY" {
			return "", errors.New("Device Not Ready")
		}
	}
	if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MDMPortHandler"] == nil {
		return "", errors.New("Device Not Ready")
	}
	MDMHandler := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MDMPortHandler"].(io.ReadWriteCloser)
	switch Key {
	case "CellNetworkRegister":
		_, _, err := SendCommand(MDMHandler, fmt.Sprintf("AT+CREG=%s", Value))
		if err != nil {
			return "", err
		}
	case "TECharset":
		_, _, err := SendCommand(MDMHandler, fmt.Sprintf("AT+CSCS=\"%s\"", Value))
		if err != nil {
			return "", err
		}
	case "MessageStorage":
		_, _, err := SendCommand(MDMHandler, fmt.Sprintf("AT+CPMS=%s", Value))
		if err != nil {
			return "", err
		}
	case "MessageFormat":
		_, _, err := SendCommand(MDMHandler, fmt.Sprintf("AT+CMGF=%s", Value))
		if err != nil {
			return "", err
		}
	case "ReadMessage":
		//TEXT
		//+CMGR: "REC READ","13800138000",,"20/01/11,19:46:33+32"
		//6D4B8BD56D4B8BD5006100620063D83CDF1A
		//PDU
		//+CMGR: 0,,37
		//[PDU]
		SMSResponse, CmdStatus, err := SendCommand(MDMHandler, fmt.Sprintf("AT+CMGR=%s", Value))
		if err != nil {
			return "", err
		}
		if CmdStatus != "OK" {
			return "", errors.New("Illegal Response")
		}
		SMSResponse = strings.Replace(SMSResponse, "+CMGR: ", "", -1)
		return SMSResponse, nil
	case "ReadMessageCDMA":
		//+CMS ERROR: Invalid index
		//
		//^HCMGR:13800138000,2020,01,11,19,54,31,0,6,14,0,0,0,1
		//mK▒▒▒<▒UJUJUJ
		SMSResponse, _, err := SendCommand(MDMHandler, fmt.Sprintf("AT^HCMGR=%s", Value))
		if err != nil {
			return "", err
		}
		//if CmdStatus != "OK" {
		//	return "", errors.New("Illegal Response")
		//}
		SMSResponse = strings.Replace(SMSResponse, "^HCMGR:", "", -1)
		return SMSResponse, nil
	case "DeleteMessage":
		_, _, err := SendCommand(MDMHandler, fmt.Sprintf("AT+CMGD=%s", Value))
		if err != nil {
			return "", err
		}
	}
	return "", nil
}

func Huawei_Get_SMS(DeviceName string) error {
	OperationMode := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["OperationMode"].(string)
	if OperationMode == "EV-DO" {
		_, err := Huawei_SET(DeviceName, "MessageFormat", "1")
		if err != nil {
			return err
		}
	} else {
		_, err := Huawei_SET(DeviceName, "MessageFormat", "0")
		if err != nil {
			return err
		}
	}
	for {
		SMSStatus, err := Huawei_GET(DeviceName, "SMSStatus")
		if err != nil {
			return err
		}
		SMSTotal, err := strconv.Atoi(strings.Split(SMSStatus, ",")[1])
		if err != nil {
			return err
		}
		Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MessageFormat"], err = Huawei_GET(DeviceName, "MessageFormat")
		if err != nil {
			return err
		}
		if SMSTotal != 0 {
			switch OperationMode {
			default:
				err := Huawei_Get_SMS_Common(DeviceName)
				if err != nil {
					return err
				}
			case "CDMA":
				err := Huawei_Get_SMS_CDMA(DeviceName)
				if err != nil {
					return err
				}
			case "EV-DO":
				err := Huawei_Get_SMS_CDMA(DeviceName)
				if err != nil {
					return err
				}
			}
		} else {
			break
		}
	}
	return nil
}

func Huawei_Get_SMS_Common(DeviceName string) error {
	MessageFormat := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MessageFormat"]
	count := -1
	for {
		count++
		SMSResponse, err := Huawei_SET(DeviceName, "ReadMessage", strconv.Itoa(count))
		if err != nil {
			return err
		}
		if SMSResponse == "" {
			continue
		}
		switch MessageFormat {
		case "TEXT":
			err = Huawei_RECV_TEXT(DeviceName, SMSResponse)
			if err != nil {
				return err
			}
		default:
			PDU := strings.Split(SMSResponse, "\r\n")[1]
			err = DecodePDU(DeviceName, PDU)
			if err != nil {
				return err
			}
		}
		_, err = Huawei_SET(DeviceName, "DeleteMessage", strconv.Itoa(count))
		if err != nil {
			return err
		}
		break
	}
	return nil
}
func Huawei_Get_SMS_CDMA(DeviceName string) error {
	MessageFormat := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MessageFormat"]
	count := -1
	for {
		count++
		SMSResponse, err := Huawei_SET(DeviceName, "ReadMessageCDMA", strconv.Itoa(count))
		if err != nil {
			return err
		}
		if SMSResponse == "" {
			continue
		}
		switch MessageFormat {
		case "TEXT":
			err = Huawei_RECV_TEXT(DeviceName, SMSResponse)
			if err != nil {
				return err
			}
		default:
			PDU := strings.Split(SMSResponse, "\r\n")[1]
			err = Huawei_RECV_PDU(DeviceName, PDU)
			if err != nil {
				return err
			}
		}
		_, err = Huawei_SET(DeviceName, "DeleteMessage", strconv.Itoa(count))
		if err != nil {
			return err
		}
		break
	}
	//Prevent CPMS return empty with OK
	time.Sleep(time.Second)
	return nil
}

func Huawei_RECV_TEXT(DeviceName string, SMSResponse string) error {
	var PhoneNumber string
	if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["PhoneNumber"] != nil {
		PhoneNumber = Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["PhoneNumber"].(string)
	} else {
		PhoneNumber = ""
	}
	Arg := strings.Split(SMSResponse, "\r\n")[0]
	Args := strings.Split(Arg, ",")
	From := Args[0]
	To := DeviceName + "@" + PhoneNumber
	Tittle := From + "->" + To
	ReceiveTime := time.Now().Format("2006-01-02 15:04:05")
	SendTime := Args[1] + "-" + Args[2] + "-" + Args[3] + " " + Args[4] + ":" + Args[5] + ":" + Args[6]
	Body := strings.Split(SMSResponse, "\r\n")[1]
	Body, _ = u2s(Body)
	Data := "From:" + From + "\r\n" + "To:" + To + "\r\n" + "Send:" + SendTime + "\r\n" + "Received:" + ReceiveTime + "\r\n" + Body
	err := DecodeText(DeviceName, Tittle, Data)
	if err != nil {
		return err
	}
	return nil
}

func Huawei_RECV_PDU(DeviceName string, PDU string) error {
	err := DecodePDU(DeviceName, PDU)
	if err != nil {
		return err
	}
	return nil
}

func Huawei_SEND_SMS(DeviceName string, DstPhone string, Content string) error {
	OperationMode := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["OperationMode"].(string)
	if OperationMode == "EV-DO" {
		_, err := Huawei_SET(DeviceName, "MessageFormat", "1")
		if err != nil {
			return err
		}
	} else {
		_, err := Huawei_SET(DeviceName, "MessageFormat", "0")
		if err != nil {
			return err
		}
	}
	MessageFormat := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MessageFormat"]
	switch MessageFormat {
	case "TEXT":
		err := Huawei_SEND_SMS_TEXT(DeviceName, DstPhone, Content)
		if err != nil {
			return err
		}
	default:
		err := Huawei_SEND_SMS_PDU(DeviceName, DstPhone, Content)
		if err != nil {
			return err
		}
	}
	return nil
}
func Huawei_SEND_SMS_TEXT(DeviceName string, DstPhone string, Content string) error {
	OperationMode := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["OperationMode"].(string)
	switch OperationMode {
	default:
		err := Huawei_SEND_SMS_COMMON_TEXT(DeviceName, DstPhone, Content)
		if err != nil {
			return err
		}
	case "CDMA":
		err := Huawei_SEND_SMS_CDMA_TEXT(DeviceName, DstPhone, Content)
		if err != nil {
			return err
		}
	case "EV-DO":
		err := Huawei_SEND_SMS_CDMA_TEXT(DeviceName, DstPhone, Content)
		if err != nil {
			return err
		}
	}
	return nil
}
func Huawei_SEND_SMS_COMMON_TEXT(DeviceName string, DstPhone string, Content string) error {
	_, _, err := DeivceSendCommand(DeviceName, "AT+CMGS=\""+DstPhone+"\"")
	if err != nil {
		return err
	}
	Response, _, _ := DeivceSendPDU(DeviceName, Content)
	if strings.Contains(Response, "+CMGS:") {
		return nil
	} else {
		return errors.New("Send Failed")
	}
}
func Huawei_SEND_SMS_CDMA_TEXT(DeviceName string, DstPhone string, Content string) error {
	_, _, err := DeivceSendCommand(DeviceName, "AT^HCMGS=\""+DstPhone+"\"")
	if err != nil {
		return err
	}
	Response, _, _ := DeivceSendPDU(DeviceName, Content)
	if strings.Contains(Response, "^HCMGSS:") {
		return nil
	} else {
		return errors.New("Send Failed")
	}
}

func Huawei_SEND_SMS_PDU(DeviceName string, DstPhone string, Content string) error {
	OperationMode := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["OperationMode"].(string)
	switch OperationMode {
	default:
		err := Huawei_SEND_SMS_COMMON_PDU(DeviceName, DstPhone, Content)
		if err != nil {
			return nil
		}
	case "CDMA":
		err := Huawei_SEND_SMS_CDMA_PDU(DeviceName, DstPhone, Content)
		if err != nil {
			return err
		}
	case "EV-DO":
		err := Huawei_SEND_SMS_CDMA_PDU(DeviceName, DstPhone, Content)
		if err != nil {
			return err
		}
	}
	return nil
}

func Huawei_SEND_SMS_COMMON_PDU(DeviceName string, DstPhone string, Content string) error {
	SMS := sms.Message{
		Text:     Content,
		Encoding: sms.Encodings.UCS2,
		Type:     sms.MessageTypes.Submit,
		Address:  sms.PhoneNumber(DstPhone),
	}
	n, PDU, err := SMS.PDU()
	if err != nil {
		return err
	}
	hexPDU := strings.Replace(hex.EncodeToString(PDU), "00010005a1", "0001000581", 1)
	if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"] != nil {
		Status := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"].(string)
		if Status != "ON" && Status != "READY" {
			return errors.New("Device Not Ready")
		}
	}
	_, _, err = DeivceSendCommand(DeviceName, "AT+CMGS="+strconv.Itoa(n))
	if err != nil {
		return err
	}
	Response, _, err := DeivceSendPDU(DeviceName, hexPDU)
	if err != nil {
		return err
	}
	if strings.Contains(Response, "+CMGS:") {
		return nil
	} else {
		return errors.New("Send Failed")
	}
}

func Huawei_SEND_SMS_CDMA_PDU(DeviceName string, DstPhone string, Content string) error {
	SMS := sms.Message{
		Text:     Content,
		Encoding: sms.Encodings.UCS2,
		Type:     sms.MessageTypes.Submit,
		Address:  sms.PhoneNumber(DstPhone),
	}
	n, PDU, err := SMS.PDU()
	if err != nil {
		return err
	}
	hexPDU := strings.Replace(hex.EncodeToString(PDU), "00010005a1", "0001000581", 1)
	if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"] != nil {
		Status := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"].(string)
		if Status != "ON" && Status != "READY" {
			return errors.New("Device Not Ready")
		}
	}
	_, _, err = DeivceSendCommand(DeviceName, "AT^HCMGS="+strconv.Itoa(n))
	if err != nil {
		return err
	}
	Response, _, err := DeivceSendPDU(DeviceName, hexPDU)
	if err != nil {
		return err
	}
	if strings.Contains(Response, "^HCMGSS:") {
		return nil
	} else {
		return errors.New("Send Failed")
	}
}
