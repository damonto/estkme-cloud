package rlpa

import (
	"encoding/binary"
	"log/slog"
	"net"
	"sync"

	"github.com/damonto/estkme-rlpa-server/internal/pkg/transmitter"
)

type Connection struct {
	ConnectionId string
	Conn         *net.TCPConn
	APDU         transmitter.APDU
	mutx         sync.Mutex
	handlers     map[byte]Handler
}

func NewConnection(connectionId string, conn *net.TCPConn) *Connection {
	c := &Connection{ConnectionId: connectionId, Conn: conn}
	c.APDU = NewAPDU(c)
	c.registerHandlers()
	return c
}

func (c *Connection) registerHandlers() {
	c.handlers = map[byte]Handler{
		TagManagement: func(conn *Connection, data []byte) error {
			return conn.Send(TagMessageBox, []byte("Welcome! \n You are connected to the server. \n This is your connection id:\n"+conn.ConnectionId))
		},
		TagProcessNotification: func(conn *Connection, data []byte) error {
			defer conn.Close()
			return conn.Send(TagMessageBox, []byte("We strongly recommend you use the management mode to process notifications. \n You are now disconnected. \n Goodbye!"))
		},
		TagDownloadProfile: func(conn *Connection, data []byte) error {
			defer conn.Close()
			conn.Send(TagMessageBox, []byte("Your profile is being downloaded. \n Please wait..."))
			if err := Download(conn, data); err != nil {
				slog.Error("error downloading profile", "error", err)
				return conn.Send(TagMessageBox, []byte("download failed \n"+err.Error()))
			}
			return conn.Send(TagMessageBox, []byte("download successful"))
		},
	}
}

func (c *Connection) Dispatch(tag byte, data []byte) {
	if handler, ok := c.handlers[tag]; ok {
		if err := handler(c, data); err != nil {
			slog.Error("error handling tag", "tag", tag, "data", data, "error", err)
		}
	}
	if tag == TagAPDU {
		c.APDU.Receiver() <- data
	}
}

func (c *Connection) Send(tag byte, data []byte) error {
	c.mutx.Lock()
	defer c.mutx.Unlock()
	packet := c.pack(tag, data)
	slog.Info("sending data", "tag", tag, "packet", packet)
	_, err := c.Conn.Write(packet)
	return err
}

func (c *Connection) Read() (byte, []byte, error) {
	header := make([]byte, 3)
	_, err := c.Conn.Read(header)
	if err != nil {
		return 0, nil, err
	}
	tag := header[0]
	length := binary.LittleEndian.Uint16(header[1:3])
	data := make([]byte, length)

	len, err := c.Conn.Read(data)
	if err != nil {
		return 0, nil, err
	}
	for len < int(length) {
		n, err := c.Conn.Read(data[len:])
		if err != nil {
			return 0, nil, err
		}
		len += n
	}
	return tag, data, nil
}

func (c *Connection) pack(tag byte, data []byte) []byte {
	var packet = make([]byte, len(data)+3)
	packet[0] = tag
	binary.LittleEndian.PutUint16(packet[1:], uint16(len(data)))
	copy(packet[3:], data)
	return packet
}

func (c *Connection) Close() error {
	c.Send(TagClose, nil)
	return c.Conn.Close()
}
