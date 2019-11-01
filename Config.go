package main

import (
	"io"
	"net/http"
)

type Device struct {
	Name      string
	User      string
	PushAddrs []PushAddr
	VID       string
	DeviceID  string
	IMEI      string

	IMSI                      string
	Model                     string
	ManufactureIdentification string
	SignalQuality             string
	Provider                  string
	HWVersion                 string
	PhoneNumber               string
	Romaning                  bool

	MDMPortName     string
	MDMPortBaudRate string
	PCUIPortName    string
	DIAGPortName    string
	TYPE            string
	PCUI            int
	DIAG            int

	MDMPortHandler  io.ReadWriteCloser
	PCUIPortHandler io.ReadWriteCloser
	DIAGPortHandler io.ReadWriteCloser
	Status          string
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

func (Dev *Device) CloseDevice() {
	Dev.Status = "OFF"
	if Dev.MDMPortHandler != nil {
		Dev.MDMPortHandler.Close()
		Dev.MDMPortHandler = nil
	}
	if Dev.PCUIPortHandler != nil {
		Dev.PCUIPortHandler.Close()
		Dev.PCUIPortHandler = nil
	}
	if Dev.DIAGPortHandler != nil {
		Dev.DIAGPortHandler.Close()
		Dev.DIAGPortHandler = nil
	}
}
