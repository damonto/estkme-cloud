package transmitter

type APDU interface {
	Lock() error
	Unlock() error
	Transmit(command string) (string, error)
	Receiver() chan []byte
}
