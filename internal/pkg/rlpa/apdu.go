package rlpa

import (
	"encoding/hex"
	"sync"

	"github.com/damonto/estkme-rlpa-server/internal/pkg/transmitter"
)

type apdu struct {
	mutex    sync.Mutex
	conn     *Connection
	receiver chan []byte
}

func NewAPDU(conn *Connection) transmitter.APDU {
	return &apdu{conn: conn, receiver: make(chan []byte, 1)}
}

func (a *apdu) Lock() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.conn.Send(TagAPDULock, nil)
}

func (a *apdu) Unlock() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.conn.Send(TagAPDUUnlock, nil)
}

func (a *apdu) Transmit(command string) (string, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	b, _ := hex.DecodeString(command)
	if err := a.conn.Send(TagAPDU, b); err != nil {
		return "", err
	}
	return hex.EncodeToString(<-a.receiver), nil
}

func (a *apdu) Receiver() chan []byte {
	return a.receiver
}
