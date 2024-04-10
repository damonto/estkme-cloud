package driver

type APDU interface {
	Lock() error
	Unlock() error
	Transmit(command string) (string, error)
	Receive() chan []byte
}
