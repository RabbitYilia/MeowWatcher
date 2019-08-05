package main

import (
	"encoding/hex"
	"github.com/xlab/at/sms"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func CheckSMS(Configure *Config, Device *Device) {
	SendCommand(Device, "AT+CPMS= \"ME\",\"ME\",\"ME\"")
	SMSStatus := SendCommand(Device, "AT+CPMS?")
	if strings.Split(SMSStatus, ",")[1] != "0" {
		sum := 0
		for {
			SMSResponse := SendCommand(Device, "AT+CMGR="+strconv.Itoa(sum))
			if !strings.Contains(SMSResponse, "+CMGR:") {
				sum++
			} else {
				break
			}
		}
		SMSResponse := SendCommand(Device, "AT+CMGR="+strconv.Itoa(sum))
		PDU := strings.Split(SMSResponse, "\r\n")
		PDUData, err := hex.DecodeString(PDU[1])
		if err != nil {
			log.Fatal(err)
		}
		msg := new(sms.Message)
		msg.ReadFrom(PDUData)
		SendTime := ProcessPDUTimestamp(msg.ServiceCenterTime.PDU())
		ReceiveTime := time.Now().Format("2006-01-02 15:04:05")
		From := string(msg.Address)
		To := Device.Name + "@" + Device.PhoneNumber
		Tittle := From + "->" + To
		Data := "From:" + From + "\r\n" + "To:" + To + "\r\n" + "Send:" + SendTime + "\r\n" + "Received:" + ReceiveTime + "\r\n" + msg.Text
		log.Println("[", Device.Name, "]", "New SMS:", Tittle, " ", msg.Text)
		for PushNum := range Device.PushAddrs {
			PushContent := strings.Replace(Device.PushAddrs[PushNum].Body, "{.Tittle}", url.QueryEscape(Tittle), -1)
			PushContent = strings.Replace(PushContent, "{.Content}", url.QueryEscape(Data), -1)
			PushByPost(Configure.client, Device.PushAddrs[PushNum].URL, PushContent)
		}
		log.Println("[", Device.Name, "]", "Delete No", sum, " SMS")
		SendCommand(Device, "AT+CMGD="+strconv.Itoa(sum))
	} else {
		log.Println("[", Device.Name, "]", "No SMS Received")
	}
}

func ProcessTimestampAttr(Data byte) string {
	high := uint(Data >> 4)
	low := uint(Data << 4)
	low = low >> 4
	return strconv.FormatUint(uint64(low), 10) + strconv.FormatUint(uint64(high), 10)
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
