package main

import (
	"github.com/tarm/goserial"
	"log"
)

func DialNumber(Device Device, Number string) {
	SendCommand(&Device, "ATD"+Number+";")
	SendCommand(&Device, "AT^DDSETEX=2")
	SerialConfig := &serial.Config{Name: Device.VoicePort, Baud: 115200, ReadTimeout: 5 /*毫秒*/}
	Handler, err := serial.OpenPort(SerialConfig)
	if err == nil {
		log.Fatal(err)
	}
	Handler.Close()

}
func ReceiveCall(Device Device) {
	SendCommand(&Device, "ATA")
	SendCommand(&Device, "AT^DDSETEX=2")
	SerialConfig := &serial.Config{Name: Device.VoicePort, Baud: 115200, ReadTimeout: 5 /*毫秒*/}
	Handler, err := serial.OpenPort(SerialConfig)
	if err == nil {
		log.Fatal(err)
	}
	Handler.Close()
}
