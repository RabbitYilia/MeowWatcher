package main

import (
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
	Device.Status = SendCommand(Device, "AT+CPAS")
	Device.Provider = SendCommand(Device, "AT+COPS?")
	log.Println("[", Device.Name, "]", "Info Updated")
}

func InitDevice(configuration *Config) {
	for i := 1; i <= 100; i++ {
		c := make(chan string, 1)
		go func() {
			log.Println("[", "Detecting", "]", "COM"+strconv.Itoa(i))
			SerialConfig := &serial.Config{Name: "COM" + strconv.Itoa(i), Baud: 115200, ReadTimeout: 5 /*毫秒*/}
			Handler, err := serial.OpenPort(SerialConfig)
			if err != nil {
				c <- ""
				return
			}
			SendCommandLow(Handler, "ATE1")
			IMEI := SendCommandLow(Handler, "AT+CGSN")
			if IMEI != "" {
				Resp := DetectedPortType(Handler)
				for DeviceNum := range configuration.Devices {
					if configuration.Devices[DeviceNum].IMEI == IMEI {
						if Resp != "" {
							configuration.Devices[DeviceNum].FeedbackPort = "COM" + strconv.Itoa(i)
							configuration.Devices[DeviceNum].FeedbackPortConfig = SerialConfig
							configuration.Devices[DeviceNum].FeedbackPortHandler = Handler
							log.Println("[", "Detect", "]", configuration.Devices[DeviceNum].Name, "FeedBack Working on", "COM"+strconv.Itoa(i))
						} else {
							configuration.Devices[DeviceNum].ManagerPort = "COM" + strconv.Itoa(i)
							configuration.Devices[DeviceNum].ManagerPortConfig = SerialConfig
							configuration.Devices[DeviceNum].ManagerPortHandler = Handler
							log.Println("[", "Detect", "]", configuration.Devices[DeviceNum].Name, "Manager Working on", "COM"+strconv.Itoa(i))
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
				log.Println("[", "Detect", "]", "COM"+strconv.Itoa(i), "Offline")
				continue
			}
			log.Println("[", "Detect", "]", "COM"+strconv.Itoa(i), "Online")
		case <-time.After(10 * time.Second):
			log.Println("[", "Detect", "]", "COM"+strconv.Itoa(i), "Offline")
			continue
		}
	}

	for DeviceNum := range configuration.Devices {
		if configuration.Devices[DeviceNum].ManagerPort == "" {
			continue
		}
		SendCommand(&configuration.Devices[DeviceNum], "ATE1")
		configuration.Devices[DeviceNum].IMEI = SendCommand(&configuration.Devices[DeviceNum], "AT+CGSN")
		configuration.Devices[DeviceNum].IMSI = SendCommand(&configuration.Devices[DeviceNum], "AT+CIMI")
		configuration.Devices[DeviceNum].ManufactureIdentification = SendCommand(&configuration.Devices[DeviceNum], "AT+CGMI")
		configuration.Devices[DeviceNum].Model = SendCommand(&configuration.Devices[DeviceNum], "AT+CGMM")
		configuration.Devices[DeviceNum].HWVersion = SendCommand(&configuration.Devices[DeviceNum], "AT+CGMR")
		configuration.Devices[DeviceNum].PhoneNumber = SendCommand(&configuration.Devices[DeviceNum], "AT+CNUM")
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
	num, err := Device.FeedbackPortHandler.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}
	StrBuffer := ""
	start := time.Now().Unix()
	for i := 0; i < 128000; i++ {
		num, err = Device.FeedbackPortHandler.Read(buffer)
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
	num, err := Device.ManagerPortHandler.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}
	num, err = Device.ManagerPortHandler.Write([]byte(Command + "\r\n"))
	if err != nil {
		log.Fatal(err)
	}
	StrBuffer := ""
	for i := 0; i < 128000; i++ {
		num, err = Device.ManagerPortHandler.Read(buffer)
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

func DetectedPortType(Handler io.ReadWriteCloser) string {
	buffer := make([]byte, MAXRWLEN)
	//发命令之前清空缓冲区
	num, err := Handler.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}
	StrBuffer := ""
	sum := 0
	start := time.Now().Unix()
	for i := 0; i < 128000; i++ {
		num, err = Handler.Read(buffer)
		sum += num
		if num > 0 {
			StrBuffer += fmt.Sprintf("%s", string(buffer[:num]))
		}
		if sum != 0 {
			break
		}
		if time.Now().Unix()-start > 6 {
			break
		}
	}
	StrBuffer = strings.Replace(StrBuffer, "\r\n\r\n", "\r\n", -1)
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
