package main

import (
	"MeowWatcher/serial"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/tarm/goserial"
)

func UpdateDeviceInfo(Device *Device) {
	Device.SignalQuality = SendCommand(Device, "AT+CSQ")
	Status := SendCommand(Device, "AT+CREG?")
	Status = strings.Split(Status, ": ")[1]
	Status = strings.Split(Status, ",")[1]
	if Status == "5" {
		Device.Romaning = true
		Device.Status = true
	} else if Status == "2" {
		Device.Romaning = false
		Device.Status = true
	} else {
		Device.Status = false
	}
	Device.Provider = SendCommand(Device, "AT+COPS?")
	log.Println("[", Device.Name, "]", "Info Updated")
}

func InitDevice(configuration *Config) {
	usbports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
	}
	for i := 1; i <= 100; i++ {
		c := make(chan string, 1)
		var DeviceTypeID int
		var HardwareIdentity string
		go func() {
			//log.Println("[", "Detecting", "]", "COM"+strconv.Itoa(i))
			SerialConfig := &serial.Config{Name: "COM" + strconv.Itoa(i), Baud: 115200, ReadTimeout: 5 /*毫秒*/}
			Handler, err := serial.OpenPort(SerialConfig)
			if err != nil {
				c <- ""
				return
			}
			for _, thisport := range usbports {
				if !thisport.IsUSB {
					c <- ""
					return
				}
				if thisport.Name == "COM"+strconv.Itoa(i) {
					args := strings.Split(thisport.DeviceID, "&")
					DeviceTypeID, err = strconv.Atoi(args[len(args)-1])
					HardwareIdentity = args[len(args)-3]
					if err != nil {
						log.Fatal(err)
					}
					break
				}
			}
			if DeviceTypeID == 1 {
				c <- ""
				return
			}
			SendCommandLow(Handler, "ATE1")
			SendCommandLow(Handler, "AT^SYSCFG=2,2,3FFFFFFF,2,4")
			IMEI := SendCommandLow(Handler, "AT+CGSN")
			if IMEI != "" {
				for DeviceNum := range configuration.Devices {
					if configuration.Devices[DeviceNum].IMEI == IMEI {
						if DeviceTypeID == 2 {
							configuration.Devices[DeviceNum].DiagnosePort = "COM" + strconv.Itoa(i)
							configuration.Devices[DeviceNum].DiagnosePortConfig = SerialConfig
							configuration.Devices[DeviceNum].DiagnosePortHandler = Handler
							log.Println("[", "Detect", "]", configuration.Devices[DeviceNum].Name, "DiagnosePort Working on", "COM"+strconv.Itoa(i))
						} else {
							configuration.Devices[DeviceNum].ATPort = "COM" + strconv.Itoa(i)
							configuration.Devices[DeviceNum].ATPortConfig = SerialConfig
							configuration.Devices[DeviceNum].ATPortHandler = Handler
							log.Println("[", "Detect", "]", configuration.Devices[DeviceNum].Name, "ATPort Working on", "COM"+strconv.Itoa(i))
						}
						if HardwareIdentity != "" {
							configuration.Devices[DeviceNum].HWIdentity = HardwareIdentity
						}
					}
				}
				c <- IMEI
			} else {
				Handler.Close()
				c <- ""
			}
		}()
		select {
		case result := <-c:
			if result == "" {
				//log.Println("[", "Detect", "]", "COM"+strconv.Itoa(i), "Offline")
				continue
			}
			log.Println("[", "Detect", "]", "COM"+strconv.Itoa(i), "Online")
		case <-time.After(6 * time.Second):
			//log.Println("[", "Detect", "]", "COM"+strconv.Itoa(i), "Offline")
			continue
		}
	}
	for DeviceNum := range configuration.Devices {
		for _, thisport := range usbports {
			if !thisport.IsUSB {
				continue
			}
			args := strings.Split(thisport.DeviceID, "&")
			HardwareIdentity := args[len(args)-3]
			if configuration.Devices[DeviceNum].HWIdentity == HardwareIdentity {
				configuration.Devices[DeviceNum].VoicePort = thisport.Name
				log.Println("[", "Detect", "]", configuration.Devices[DeviceNum].Name, " Voice Port is ", thisport.Name)
				break
			}
		}
	}
	for DeviceNum := range configuration.Devices {
		if configuration.Devices[DeviceNum].ATPort == "" {
			continue
		}
		SendCommand(&configuration.Devices[DeviceNum], "ATE1")
		Status := SendCommand(&configuration.Devices[DeviceNum], "AT+CREG?")
		Status = strings.Split(Status, ": ")[1]
		Status = strings.Split(Status, ",")[1]
		if Status == "5" {
			configuration.Devices[DeviceNum].Romaning = true
			configuration.Devices[DeviceNum].Status = true
		} else if Status == "2" {
			configuration.Devices[DeviceNum].Romaning = false
			configuration.Devices[DeviceNum].Status = true
		} else {
			configuration.Devices[DeviceNum].Status = false
		}
		configuration.Devices[DeviceNum].IMEI = SendCommand(&configuration.Devices[DeviceNum], "AT+CGSN")
		configuration.Devices[DeviceNum].IMSI = SendCommand(&configuration.Devices[DeviceNum], "AT+CIMI")
		configuration.Devices[DeviceNum].ManufactureIdentification = SendCommand(&configuration.Devices[DeviceNum], "AT+CGMI")
		configuration.Devices[DeviceNum].Model = SendCommand(&configuration.Devices[DeviceNum], "AT+CGMM")
		configuration.Devices[DeviceNum].HWVersion = SendCommand(&configuration.Devices[DeviceNum], "AT+CGMR")
		configuration.Devices[DeviceNum].PhoneNumber = SendCommand(&configuration.Devices[DeviceNum], "AT+CNUM")
		SendCommand(&configuration.Devices[DeviceNum], "AT+CRC=1")
		SendCommand(&configuration.Devices[DeviceNum], "AT+CLIP=1")
		SendCommand(&configuration.Devices[DeviceNum], "AT^CURC=1")
		if configuration.Devices[DeviceNum].PhoneNumber != "" {
			configuration.Devices[DeviceNum].PhoneNumber = strings.Replace(strings.Split(configuration.Devices[DeviceNum].PhoneNumber, ",")[1], "\"", "", -1)
		}
		TECharSet := SendCommand(&configuration.Devices[DeviceNum], "AT+CSCS=?")
		if strings.Contains(TECharSet, "UCS2") {
			SendCommand(&configuration.Devices[DeviceNum], "AT+CSCS=\"UCS2\"")
		}
		SMSMode := SendCommand(&configuration.Devices[DeviceNum], "AT+CMGF=?")
		if strings.Contains(SMSMode, "0") {
			SendCommand(&configuration.Devices[DeviceNum], "AT+CMGF=0")
		}
		SendCommand(&configuration.Devices[DeviceNum], "AT+CPMS= \"ME\",\"ME\",\"ME\"")
		log.Println("[", configuration.Devices[DeviceNum].Name, "]", "Init Finished")
	}
}
func GetFeedback(Device *Device, ReadTime int64) string {
	buffer := make([]byte, MAXRWLEN)
	//发命令之前清空缓冲区
	num, err := Device.DiagnosePortHandler.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}
	StrBuffer := ""
	start := time.Now().Unix()
	for i := 0; i < 128000; i++ {
		num, err = Device.DiagnosePortHandler.Read(buffer)
		if num > 0 {
			StrBuffer += fmt.Sprintf("%s", string(buffer[:num]))
		}
		if num == 0 && time.Now().Unix()-start > ReadTime {
			break
		}
		if strings.LastIndex(StrBuffer, "\r\n") > 0 {
			StrBuffer = strings.TrimLeft(StrBuffer, "\r\n")
			break
		}
	}
	StrBuffer = strings.Replace(StrBuffer, "\r\n\r\n", "\r\n", -1)
	StrBuffer = strings.TrimRight(StrBuffer, "\r\n")
	return StrBuffer
}

func ProcessFeedBack(Device *Device, FeedBacks string) {
	if FeedBacks == "" {
		return
	}
	log.Println("[", "Feedback", "]", Device.Name, ":", FeedBacks)
}

func SendCommand(Device *Device, Command string) string {
	buffer := make([]byte, MAXRWLEN)
	//发命令之前清空缓冲区
	num, err := Device.ATPortHandler.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}
	num, err = Device.ATPortHandler.Write([]byte(Command + "\r\n"))
	if err != nil {
		log.Fatal(err)
	}
	StrBuffer := ""
	for i := 0; i < 128000; i++ {
		num, err = Device.ATPortHandler.Read(buffer)
		if num > 0 {
			StrBuffer += fmt.Sprintf("%s", string(buffer[:num]))
		}
		if strings.LastIndex(StrBuffer, "\r\nCOMMAND NOT SUPPORT\r\n") > 0 {
			StrBuffer = strings.Split(StrBuffer, "\r\nCOMMAND NOT SUPPORT\r\n")[0]
			break
		}
		if strings.LastIndex(StrBuffer, "\r\nOK\r\n") > 0 {
			StrBuffer = strings.Split(StrBuffer, "\r\nOK\r\n")[0]
			break
		}
	}
	StrBuffer = strings.Replace(StrBuffer, Command+"\r", Command+"\r\n", -1)
	StrBuffer = strings.Replace(StrBuffer, "\r\n\r\n", "\r\n", -1)
	StrBuffer = strings.Replace(StrBuffer, Command+"\r\n", "", -1)
	StrBuffer = strings.TrimRight(StrBuffer, "\r\n")
	return StrBuffer
}

func SendCommandLow(Handler io.ReadWriteCloser, Command string) string {
	buffer := make([]byte, MAXRWLEN)
	//发命令之前清空缓冲区
	num, err := Handler.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}
	num, err = Handler.Write([]byte(Command + "\r\n"))
	if err != nil {
		log.Fatal(err)
	}
	StrBuffer := ""
	start := time.Now().Unix()
	for i := 0; i < 128000; i++ {
		num, err = Handler.Read(buffer)
		if num > 0 {
			StrBuffer += fmt.Sprintf("%s", string(buffer[:num]))
		}
		if strings.LastIndex(StrBuffer, "\r\nCOMMAND NOT SUPPORT\r\n") > 0 {
			StrBuffer = strings.Split(StrBuffer, "\r\nCOMMAND NOT SUPPORT\r\n")[0]
			break
		}
		if strings.LastIndex(StrBuffer, "\r\nOK\r\n") > 0 {
			StrBuffer = strings.Split(StrBuffer, "\r\nOK\r\n")[0]
			break
		}
		if time.Now().Unix()-start > 5 {
			break
		}
	}
	StrBuffer = strings.Replace(StrBuffer, Command+"\r", Command+"\r\n", -1)
	StrBuffer = strings.Replace(StrBuffer, "\r\n\r\n", "\r\n", -1)
	StrBuffer = strings.Replace(StrBuffer, Command+"\r\n", "", -1)
	StrBuffer = strings.TrimRight(StrBuffer, "\r\n")
	return StrBuffer
}
