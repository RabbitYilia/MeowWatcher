package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"github.com/xlab/at/sms"
	"log"
	"strconv"
	"time"
)

func DecodeText(DeviceName string, Tittle string,Data string) error {
	log.Println("[", DeviceName, "]", "New SMS:", Tittle, " ", Data)
	PushSC(DeviceName, Tittle, Data)
	PushTG(DeviceName, Tittle, Data)
	AddSMSToDB(DeviceName, Tittle, Data)
	return nil
}

func DecodePDU(DeviceName string, PDU string) error {
	var PhoneNumber string
	if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["PhoneNumber"] != nil {
		PhoneNumber = Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["PhoneNumber"].(string)
	} else {
		PhoneNumber = ""
	}
	PDUData, err := hex.DecodeString(PDU)
	if err != nil {
		log.Println(err)
		return err
	}
	msg := new(sms.Message)
	_, err = msg.ReadFrom(PDUData)
	if err != nil {
		PushSC(DeviceName, DeviceName, PDU)
		PushTG(DeviceName, DeviceName, PDU)
		AddSMSToDB(DeviceName, DeviceName, PDU)
		log.Println(err)
		return nil
	}
	SendTime := ProcessPDUTimestamp(msg.ServiceCenterTime.PDU())
	ReceiveTime := time.Now().Format("2006-01-02 15:04:05")
	From := string(msg.Address)
	To := DeviceName + "@" + PhoneNumber
	Tittle := From + "->" + To
	Data := "From:" + From + "\r\n" + "To:" + To + "\r\n" + "Send:" + SendTime + "\r\n" + "Received:" + ReceiveTime + "\r\n" + msg.Text
	log.Println("[", DeviceName, "]", "New SMS:", Tittle, " ", msg.Text)
	PushSC(DeviceName, Tittle, Data)
	PushTG(DeviceName, Tittle, Data)
	AddSMSToDB(DeviceName, Tittle, Data)
	return nil
}

func ProcessPDUTimestamp(data []byte) string {
	Year := ProcessTimestampAttr(data[0])
	Month := ProcessTimestampAttr(data[1])
	Day := ProcessTimestampAttr(data[2])
	Hour := ProcessTimestampAttr(data[3])
	Minute := ProcessTimestampAttr(data[4])
	Sec := ProcessTimestampAttr(data[5])
	Zone := ProcessTimestampAttr(data[6])
	ZoneInt, err := strconv.Atoi(Zone)
	if err != nil {
		log.Fatal()
	}
	Zone = strconv.Itoa(ZoneInt / 4)
	return "UTC+" + Zone + " 20" + Year + "-" + Month + "-" + Day + " " + Hour + ":" + Minute + ":" + Sec
}

func ProcessTimestampAttr(Data byte) string {
	high := uint(Data >> 4)
	low := uint(Data << 4)
	low = low >> 4
	return strconv.FormatUint(uint64(low), 10) + strconv.FormatUint(uint64(high), 10)
}

func u2s(form string) (to string, err error) {
	bs := []byte(form)
	if err != nil {
		return
	}
	for i, bl, br, r := 0, len(bs), bytes.NewReader(bs), uint16(0); i < bl; i += 2 {
		binary.Read(br, binary.BigEndian, &r)
		to += string(r)
	}
	return
}
