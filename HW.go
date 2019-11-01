package main

import (
	"io"
	"regexp"
	"strconv"
	"strings"
)

func Huawei_Startup(Handler io.ReadWriteCloser, MDMDevice *Device) error {
	err := Huawei_Startup_COMMON_INIT(Handler, MDMDevice)
	if err != nil {
		MDMDevice.CloseDevice()
		return err
	}
	TYPE, _, DIAG, PCUI, err := Huawei_GET_PORTSEQ(Handler, MDMDevice)
	if err != nil {
		MDMDevice.CloseDevice()
		return err
	}
	switch TYPE {
	case "WCDMA":
		err = Huawei_Startup_WCDMA_INIT(Handler, MDMDevice)
		if err != nil {
			MDMDevice.CloseDevice()
			return err
		}
	case "EV-DO":
		err = Huawei_Startup_CDMA_INIT(Handler, MDMDevice)
		if err != nil {
			MDMDevice.CloseDevice()
			return err
		}
	}
	MDMDevice.TYPE = TYPE
	MDMDevice.DIAG, _ = strconv.Atoi(DIAG)
	MDMDevice.PCUI, _ = strconv.Atoi(PCUI)
	return nil
}
func Huawei_Startup_COMMON_INIT(Handler io.ReadWriteCloser, MDMDevice *Device) error {
	_, _, err := SendCommandLow(Handler, "ATE1")
	if err != nil {
		MDMDevice.CloseDevice()
		return err
	}
	_, _, err = SendCommandLow(Handler, "ATZ")
	if err != nil {
		MDMDevice.CloseDevice()
		return err
	}
	_, _, err = SendCommandLow(Handler, "AT&F")
	if err != nil {
		MDMDevice.CloseDevice()
		return err
	}
	_, _, err = SendCommandLow(Handler, "ATE1")
	if err != nil {
		MDMDevice.CloseDevice()
		return err
	}
	_, _, err = SendCommandLow(Handler, "AT+CMEE=2")
	if err != nil {
		MDMDevice.CloseDevice()
		return err
	}
	MDMDevice.IMSI, _, err = SendCommandLow(Handler, "AT+CIMI")
	if err != nil {
		MDMDevice.CloseDevice()
		return err
	}
	MDMDevice.ManufactureIdentification, _, err = SendCommandLow(Handler, "AT+CGMI")
	if err != nil {
		MDMDevice.CloseDevice()
		return err
	}
	MDMDevice.ManufactureIdentification = strings.Replace(MDMDevice.ManufactureIdentification, "+GMI: ", "", -1)
	MDMDevice.Model, _, err = SendCommandLow(Handler, "AT+CGMM")
	if err != nil {
		MDMDevice.CloseDevice()
		return err
	}
	_, _, err = SendCommandLow(Handler, "AT+CRC=1")
	if err != nil {
		MDMDevice.CloseDevice()
		return err
	}
	SMSMode, _, err := SendCommandLow(Handler, "AT+CMGF=?")
	if err != nil {
		MDMDevice.CloseDevice()
		return err
	}
	if strings.Contains(SMSMode, "0") {
		_, _, err = SendCommandLow(Handler, "AT+CMGF=0")
		if err != nil {
			MDMDevice.CloseDevice()
			return err
		}
	} else if strings.Contains(SMSMode, "1") {
		_, _, err = SendCommandLow(Handler, "AT+CMGF=1")
		if err != nil {
			MDMDevice.CloseDevice()
			return err
		}
	}
	_, _, err = SendCommandLow(Handler, "AT+CPMS= \"ME\",\"ME\",\"ME\"")
	if err != nil {
		MDMDevice.CloseDevice()
		return err
	}
	MDMDevice.Status = "READY"
	return nil
}
func Huawei_Startup_WCDMA_INIT(Handler io.ReadWriteCloser, MDMDevice *Device) error {
	var err error
	PhoneNumberExp := regexp.MustCompile(`,\"[\+0-9]+\",`)
	MDMDevice.HWVersion, _, err = SendCommandLow(Handler, "AT+CGMR")
	if err != nil {
		MDMDevice.CloseDevice()
		return err
	}
	PhoneNumber, _, err := SendCommandLow(Handler, "AT+CNUM")
	if err != nil {
		MDMDevice.CloseDevice()
		return err
	}
	if PhoneNumber != "" {
		MDMDevice.PhoneNumber = strings.Replace(PhoneNumberExp.FindStringSubmatch(PhoneNumber)[0], "\"", "", -1)
	}
	_, _, err = SendCommandLow(Handler, "AT^SYSCFG=2,2,3FFFFFFF,2,4")
	if err != nil {
		MDMDevice.CloseDevice()
		return err
	}
	_, _, err = SendCommandLow(Handler, "AT+CLIP=1")
	if err != nil {
		MDMDevice.CloseDevice()
		return err
	}
	_, _, err = SendCommandLow(Handler, "AT^CURC=1")
	if err != nil {
		MDMDevice.CloseDevice()
		return err
	}
	TECharSet, _, err := SendCommandLow(Handler, "AT+CSCS=?")
	if err != nil {
		MDMDevice.CloseDevice()
		return err
	}
	if strings.Contains(TECharSet, "UCS2") {
		_, _, err = SendCommandLow(Handler, "AT+CSCS=\"UCS2\"")
		if err != nil {
			MDMDevice.CloseDevice()
			return err
		}
	}
	return nil
}
func Huawei_Startup_CDMA_INIT(Handler io.ReadWriteCloser, MDMDevice *Device) error {
	return nil
}
func Huawei_GET_PORTSEQ(Handler io.ReadWriteCloser, MDMDevice *Device) (string, string, string, string, error) {
	//^GETPORTMODE:TYPE:EV-DO:Qualcomm,MDM:0,DIAG:1,PCUI:2,CDROM:3
	//^GETPORTMODE:TYPE:WCDMA:Qualcomm,MDM:0,DIAG:1,PCUI:2
	var err error
	Data, _, err := SendCommandLow(Handler, "AT^GETPORTMODE")
	if err != nil {
		MDMDevice.CloseDevice()
		return "", "", "", "", err
	}
	Data = strings.Replace(Data, "^GETPORTMODE:", "", -1)
	Attrs := strings.Split(Data, ",")
	TYPE := ""
	MDM := ""
	DIAG := ""
	PCUI := ""
	for _, Attr := range Attrs {
		Key := strings.Split(Attr, ":")[0]
		Value := strings.Split(Attr, ":")[1]
		switch Key {
		case "TYPE":
			TYPE = Value
		case "MDM":
			MDM = Value
		case "DIAG":
			DIAG = Value
		case "PCUI":
			PCUI = Value
		}
	}
	return TYPE, MDM, DIAG, PCUI, nil
}
