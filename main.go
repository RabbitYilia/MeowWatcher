package main

import (
	"encoding/json"
	"github.com/tarm/goserial"
	"log"
	"os"
	"strconv"
	"time"
)

var configuration Config

func main() {
	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	configuration = Config{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	SetProxy(&configuration)
	for {
		start := time.Now().Unix()
		GetModemPort()
		log.Println(configuration)
		for DeviceNum := range configuration.Devices {
			if configuration.Devices[DeviceNum].Status == "READY" {
				if configuration.Devices[DeviceNum].MDMPortName == "" {
					configuration.Devices[DeviceNum].Status = "OFF"
					continue
				}
				MDMBaudRate, _ := strconv.Atoi(configuration.Devices[DeviceNum].MDMPortBaudRate)
				MDMSerialConfig := &serial.Config{Name: configuration.Devices[DeviceNum].MDMPortName, Baud: MDMBaudRate, ReadTimeout: 5 /*毫秒*/}
				MDMHandler, err := serial.OpenPort(MDMSerialConfig)
				if err == nil {
					configuration.Devices[DeviceNum].MDMPortHandler = MDMHandler
				} else {
					log.Println(err)
					configuration.Devices[DeviceNum].Status = "OFF"
					continue
				}
				if configuration.Devices[DeviceNum].PCUIPortName != "" {
					PCUISerialConfig := &serial.Config{Name: configuration.Devices[DeviceNum].PCUIPortName, Baud: 115200, ReadTimeout: 5 /*毫秒*/}
					PCUIHandler, err := serial.OpenPort(PCUISerialConfig)
					if err == nil {
						configuration.Devices[DeviceNum].PCUIPortHandler = PCUIHandler
					}
				}
				if configuration.Devices[DeviceNum].PCUIPortName != "" {
					DIAGSerialConfig := &serial.Config{Name: configuration.Devices[DeviceNum].DIAGPortName, Baud: 115200, ReadTimeout: 5 /*毫秒*/}
					DIAGHandler, err := serial.OpenPort(DIAGSerialConfig)
					if err == nil {
						configuration.Devices[DeviceNum].DIAGPortHandler = DIAGHandler
					}
				}
				configuration.Devices[DeviceNum].Status = "ON"
			}
			if configuration.Devices[DeviceNum].Status == "ON" {
				CheckSMS(&configuration, &configuration.Devices[DeviceNum])
			}
		}
		end := time.Now().Unix()
		if end-start < 5 {
			time.Sleep(time.Second * time.Duration(5-end+start))
		}
	}
}
