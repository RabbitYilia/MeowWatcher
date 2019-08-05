package main

func DialNumber(Device Device, Number string) {
	SendCommand(&Device, "ATD"+Number+";")
	SendCommand(&Device, "AT^DDSETEX=2")

}
func ReceiveCall(Device Device) {
	SendCommand(&Device, "ATA")
	SendCommand(&Device, "AT^DDSETEX=2")
}
