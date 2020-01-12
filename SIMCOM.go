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

func SIMCOM_INIT(DeviceName string) error {
	var err error
	if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"] != nil {
		Status := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"].(string)
		if Status != "ON" && Status != "READY" {
			return errors.New("Device Not READY")
		}
	}
	Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["HWVersion"], err = SIMCOM_GET(DeviceName, "HWVersion")
	if err != nil {
		return err
	}
	Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Model"], err = SIMCOM_GET(DeviceName, "Model")
	if err != nil {
		return err
	}
	_, err = SIMCOM_SET(DeviceName, "CellNetworkRegister", "2")
	if err != nil {
		return err
	}
	_, err = SIMCOM_SET(DeviceName, "TECharset", "UCS2")
	if err != nil {
		return err
	}
	_, err = SIMCOM_SET(DeviceName, "MessageStorage", "\"ME\",\"ME\",\"ME\"")
	if err != nil {
		return err
	}
	Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["PhoneNumber"], err = SIMCOM_GET(DeviceName, "PhoneNumber")
	if err != nil {
		return err
	}
	err = SIMCOM_Status_Update(DeviceName)
	if err != nil {
		return err
	}
	return nil
}

func SIMCOM_Status_Update(DeviceName string) error {
	if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"] != nil {
		Status := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"].(string)
		if Status != "ON" && Status != "READY" {
			return errors.New("Device Not READY")
		}
	}
	CellNetworkRegisterStatus, err := SIMCOM_GET(DeviceName, "CellNetworkRegisterStatus")
	if err != nil {
		return err
	}
	switch CellNetworkRegisterStatus {
	case "":
		return errors.New("Device Not Ready")
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
				break
			}
		}
		Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["RegisterStatus"] = "No"
		return errors.New("Device Not Registered")
	}
	OperatorName, err := SIMCOM_GET(DeviceName, "GetOperatorName")
	if err != nil {
		return err
	}
	Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["OperatorName"] = OperatorName
	OperationMode, err := SIMCOM_GET(DeviceName, "GetOperationMode")
	if err != nil {
		return err
	}
	if OperationMode == "NO SERVICE" {
		return errors.New("OperationMode = NO SERVICE")
	}
	Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["OperationMode"] = OperationMode
	OperationStatus, err := SIMCOM_GET(DeviceName, "OperationStatus")
	if err != nil {
		return err
	}
	if OperationStatus != "Online" {
		return errors.New("OperationStatus = " + OperationStatus)
	}
	return nil
}

func SIMCOM_GET(DeviceName string, Key string) (string, error) {
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
		//+CGMR: LE11B12SIM7600M22
		HWVersion, _, err := SendCommand(MDMHandler, "AT+CGMR")
		if err != nil {
			return "", err
		}
		if HWVersion == "" {
			return "", errors.New("Illegal Response")
		}
		return strings.Replace(HWVersion, "+CGMR: ", "", -1), nil
	case "Model":
		//SIMCOM_SIM7600CE-T
		Model, _, err := SendCommand(MDMHandler, "AT+CGMM")
		if err != nil {
			return "", err
		}
		if Model == "" {
			return "", errors.New("Illegal Response")
		}
		return strings.Replace(Model, "+CGMM: ", "", -1), nil
	case "CellNetworkRegisterStatus":
		//+CREG: 2,1,FFFF,FFFFFFF
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
		//+COPS: 0,0,"Mi Mobile",7
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
		//Dual-Line
		//+CPSI: CDMA,Online,460-03,1019,248,-4.7,-12.8,0,0,-3276.8,14009,65535,6,42414
		//+CPSI: LTE,Online,460-11,0x2C1C,261237426,269,EUTRAN-BAND3,1850,5,5,-85,-1042,-760,14
		//
		OperationMode, _, err := SendCommand(MDMHandler, "AT+CPSI?")
		if err != nil {
			return "", err
		}
		if OperationMode == "" {
			return "", errors.New("Illegal Response")
		}
		OperationModes := strings.Split(OperationMode, "+CPSI: ")
		OperationMode = OperationModes[len(OperationModes)-1]
		OperationMode = strings.Replace(OperationMode, "+CPSI: ", "", -1)
		OperationMode = strings.Split(OperationMode, ",")[0]
		return OperationMode, nil
	case "OperationStatus":
		OperationStatus, _, err := SendCommand(MDMHandler, "AT+CPSI?")
		if err != nil {
			return "", err
		}
		if OperationStatus == "" {
			return "", errors.New("Illegal Response")
		}
		OperationStatuss := strings.Split(OperationStatus, "+CPSI: ")
		OperationStatus = OperationStatuss[len(OperationStatuss)-1]
		OperationStatus = strings.Replace(OperationStatus, "+CPSI: ", "", -1)
		OperationStatus = strings.Split(OperationStatus, ",")[1]
		return OperationStatus, nil
	case "SMSStatus":
		//+CPMS: "ME",0,99,"ME",0,99,"ME",0,99
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
		//+CMGF: 1
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
		//+CNUM: "","+8617010201799",145
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

func SIMCOM_SET(DeviceName string, Key string, Value string) (string, error) {
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
		//PDU
		//+CMGR: 1,,53
		//[PDU]

		//TEXT
		//+CMGR: "REC UNREAD","13800138000","20/01/11,19:33:23+00",,129,14
		//6D4B8BD5D83CDF1A554A554A554A

		SMSResponse, CmdStatus, err := SendCommand(MDMHandler, fmt.Sprintf("AT+CMGR=%s", Value))
		if err != nil {
			return "", err
		}
		if CmdStatus != "OK" {
			return "", errors.New("Illegal Response")
		}
		SMSResponse = strings.Replace(SMSResponse, "+CMGR: ", "", -1)
		return SMSResponse, nil
	case "DeleteMessage":
		_, _, err := SendCommand(MDMHandler, fmt.Sprintf("AT+CMGD=%s", Value))
		if err != nil {
			return "", err
		}
	}
	return "", nil
}

func SIMCOM_Get_SMS(DeviceName string) error {
	if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"] != nil {
		Status := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"].(string)
		if Status != "ON" && Status != "READY" {
			return errors.New("Device Not Ready")
		}
	}
	OperationMode := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["OperationMode"].(string)
	switch OperationMode {
	case "EVDO":
		_, err := SIMCOM_SET(DeviceName, "MessageFormat", "1")
		if err != nil {
			return err
		}
	case "CDMA":
		_, err := SIMCOM_SET(DeviceName, "MessageFormat", "1")
		if err != nil {
			return err
		}
	default:
		_, err := SIMCOM_SET(DeviceName, "MessageFormat", "0")
		if err != nil {
			return err
		}
	}
	for {
		SMSStatus, err := SIMCOM_GET(DeviceName, "SMSStatus")
		if err != nil {
			return err
		}
		SMSTotal, err := strconv.Atoi(strings.Split(SMSStatus, ",")[1])
		if err != nil {
			return err
		}
		Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MessageFormat"], err = SIMCOM_GET(DeviceName, "MessageFormat")
		if err != nil {
			return err
		}
		if SMSTotal != 0 {
			err = SIMCOM_Get_SMS_Common(DeviceName)
			if err != nil {
				return err
			}
		} else {
			break
		}
	}
	return nil
}

func SIMCOM_Get_SMS_Common(DeviceName string) error {
	MessageFormat := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MessageFormat"]
	if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"] != nil {
		Status := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"].(string)
		if Status != "ON" && Status != "READY" {
			return errors.New("Device Not Ready")
		}
	}
	count := -1
	for {
		count++
		SMSResponse, err := SIMCOM_SET(DeviceName, "ReadMessage", strconv.Itoa(count))
		if err != nil {
			return err
		}
		if SMSResponse == "" {
			continue
		}
		switch MessageFormat {
		case "TEXT":
			err = SIMCOM_RECV_TEXT(DeviceName, SMSResponse)
			if err != nil {
				return err
			}
		default:
			PDU := strings.Split(SMSResponse, "\r\n")[1]
			err = SIMCOM_RECV_PDU(DeviceName, PDU)
			if err != nil {
				return err
			}
		}
		_, err = SIMCOM_SET(DeviceName, "DeleteMessage", strconv.Itoa(count))
		if err != nil {
			return err
		}
		break
	}
	return nil
}

func SIMCOM_RECV_TEXT(DeviceName string, SMSResponse string) error {
	var PhoneNumber string
	if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["PhoneNumber"] != nil {
		PhoneNumber = Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["PhoneNumber"].(string)
	} else {
		PhoneNumber = ""
	}
	Arg := strings.Split(SMSResponse, "\r\n")[0]
	Args := strings.Split(Arg, ",")
	From := Args[1]
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

func SIMCOM_RECV_PDU(DeviceName string, PDU string) error {
	err := DecodePDU(DeviceName, PDU)
	if err != nil {
		return err
	}
	return nil
}

func SIMCOM_SEND_SMS(DeviceName string, DstPhone string, Content string) error {
	OperationMode := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["OperationMode"].(string)
	switch OperationMode {
	case "EVDO":
		_, err := SIMCOM_SET(DeviceName, "MessageFormat", "1")
		if err != nil {
			return err
		}
	case "CDMA":
		_, err := SIMCOM_SET(DeviceName, "MessageFormat", "1")
		if err != nil {
			return err
		}
	default:
		_, err := SIMCOM_SET(DeviceName, "MessageFormat", "0")
		if err != nil {
			return err
		}
	}
	MessageFormat := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MessageFormat"]
	switch MessageFormat {
	case "TEXT":
		err := SIMCOM_SEND_SMS_TEXT(DeviceName, DstPhone, Content)
		if err != nil {
			return err
		}
	default:
		err := SIMCOM_SEND_SMS_PDU(DeviceName, DstPhone, Content)
		if err != nil {
			return err
		}
	}
	return nil
}
func SIMCOM_SEND_SMS_TEXT(DeviceName string, DstPhone string, Content string) error {
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

func SIMCOM_SEND_SMS_PDU(DeviceName string, DstPhone string, Content string) error {
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
