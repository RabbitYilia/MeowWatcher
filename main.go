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
	FeedbackPort              string
	ManagerPort               string
	IMEI                      string
	IMSI                      string
	Model                     string
	ManufactureIdentification string
	SignalQuality             string
	Provider                  string
	Status                    string
	HWVersion                 string
	PhoneNumber               string
	FeedbackPortConfig        *serial.Config
	ManagerPortConfig         *serial.Config
	ManagerPortHandler        io.ReadWriteCloser
	FeedbackPortHandler       io.ReadWriteCloser
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
		for DeviceNum, _ := range configuration.Devices {
			if configuration.Devices[DeviceNum].ManagerPort == "" {
				continue
			}
			UpdateDeviceInfo(&configuration.Devices[DeviceNum])
			CheckSMS(&configuration, &configuration.Devices[DeviceNum])
			if configuration.Devices[DeviceNum].FeedbackPort == "" {
				continue
			}
			ProcessFeedBack(&configuration.Devices[DeviceNum], GetFeedback(&configuration.Devices[DeviceNum], 5))
		}

	}
}
