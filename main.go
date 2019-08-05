package main

import (
	"encoding/json"
	"github.com/tarm/goserial"
	"io"
	"log"
	"net/http"
	"os"
)

const MAXRWLEN = 128000

type Device struct {
	Name                      string
	User                      string
	PushAddrs                 []PushAddr
	DiagnosePort              string
	ATPort                    string
	IMEI                      string
	IMSI                      string
	Model                     string
	ManufactureIdentification string
	SignalQuality             string
	Provider                  string
	Status                    string
	HWVersion                 string
	PhoneNumber               string
	DiagnosePortConfig        *serial.Config
	ATPortConfig              *serial.Config
	ATPortHandler             io.ReadWriteCloser
	DiagnosePortHandler       io.ReadWriteCloser
}

type Config struct {
	Devices []Device
	Proxy   string
	client  *http.Client
}
type PushAddr struct {
	URL  string
	Body string
}

func main() {
	file, _ := os.Open("conf.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Config{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Fatal(err)
	}

	SetProxy(&configuration)
	InitDevice(&configuration)
	log.Println(configuration)
	for {
		for DeviceNum := range configuration.Devices {
			if configuration.Devices[DeviceNum].ATPort == "" {
				continue
			}
			UpdateDeviceInfo(&configuration.Devices[DeviceNum])
			CheckSMS(&configuration, &configuration.Devices[DeviceNum])
			if configuration.Devices[DeviceNum].DiagnosePort == "" {
				continue
			}
			SendCommandLow(configuration.Devices[DeviceNum].DiagnosePortHandler, "ATE")
			ProcessFeedBack(&configuration.Devices[DeviceNum], GetFeedback(&configuration.Devices[DeviceNum], 5))
		}

	}
}
