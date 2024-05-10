package cloud

import (
	"encoding/hex"
	"sync"
	"time"

	"github.com/damonto/estkme-cloud/internal/driver"
)

const (
	APDUCardDead = "6FFF"
)

type apdu struct {
	lock     sync.Mutex
	conn     *Conn
	receiver chan []byte
}

func NewAPDU(conn *Conn) driver.APDU {
	return &apdu{conn: conn, receiver: make(chan []byte, 1)}
}

func (a *apdu) Lock() error {
	a.lock.Lock()
	return a.conn.Send(TagAPDULock, nil)
}

func (a *apdu) Unlock() error {
	defer a.lock.Unlock()
	return a.conn.Send(TagAPDUUnlock, nil)
}

func (a *apdu) Transmit(command string) (string, error) {
	b, _ := hex.DecodeString(command)
	if err := a.conn.Send(TagAPDU, b); err != nil {
		return "", err
	}

	select {
	case r := <-a.receiver:
		return hex.EncodeToString(r), nil
	case <-time.After(15 * time.Second): // If response is not received in 15 seconds, return card dead.
		return hex.EncodeToString([]byte(APDUCardDead)), nil
	}
}

func (a *apdu) Receive() chan []byte {
	return a.receiver
}
