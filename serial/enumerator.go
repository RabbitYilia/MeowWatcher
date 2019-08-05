package enumerator

type PortDetails struct {
	Name     string
	IsUSB    bool
	DeviceID string
}

func GetDetailedPortsList() ([]*PortDetails, error) {
	return nativeGetDetailedPortsList()
}

type PortEnumerationError struct {
	causedBy error
}

// Error returns the complete error code with details on the cause of the error
func (e PortEnumerationError) Error() string {
	reason := "Error while enumerating serial ports"
	if e.causedBy != nil {
		reason += ": " + e.causedBy.Error()
	}
	return reason
}
