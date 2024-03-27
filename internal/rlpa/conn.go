package rlpa

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"log/slog"
	"net"
	"sync"

	"github.com/damonto/estkme-rlpa-server/internal/transmitter"
)

type Conn struct {
	Id       string
	Conn     *net.TCPConn
	APDU     transmitter.APDU
	lock     sync.Mutex
	handlers map[byte]Handler
}

var ErrorTagUnknown = errors.New("unknown tag")

func NewConn(id string, conn *net.TCPConn) *Conn {
	c := &Conn{Id: id, Conn: conn}
	c.APDU = NewAPDU(c)
	c.registerHandlers()
	return c
}

func (c *Conn) registerHandlers() {
	c.handlers = map[byte]Handler{
		TagManagement: func(conn *Conn, data []byte) error {
			return conn.Send(TagMessageBox, []byte("Welcome! \n You are connected to the server. \n Here is your PIN\n"+conn.Id))
		},
		TagProcessNotification: func(conn *Conn, data []byte) error {
			defer conn.Close()
			conn.Send(TagMessageBox, []byte("Processing notifications..."))
			if err := processNotification(conn); err != nil {
				slog.Error("error processing notification", "error", err)
				return conn.Send(TagMessageBox, []byte("Process failed \n"+err.Error()))
			}
			return conn.Send(TagMessageBox, []byte("All notifications have been processed successfully"))
		},
		TagDownloadProfile: func(conn *Conn, data []byte) error {
			defer conn.Close()
			conn.Send(TagMessageBox, []byte("Your profile is being downloaded. \n Please wait..."))
			if err := downloadProfile(conn, data); err != nil {
				slog.Error("error downloading profile", "error", err)
				return conn.Send(TagMessageBox, []byte("Download failed \n"+err.Error()))
			}
			return conn.Send(TagMessageBox, []byte("Your profile has been downloaded successfully"))
		},
	}
}

func (c *Conn) Dispatch(tag byte, data []byte) {
	if handler, ok := c.handlers[tag]; ok {
		if err := handler(c, data); err != nil {
			slog.Error("error handling tag", "tag", tag, "data", data, "error", err)
		}
	}
	if tag == TagAPDU {
		c.APDU.Receiver() <- data
	}
}

func (c *Conn) Send(tag byte, data []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	packet := c.pack(tag, data)
	if tag == TagAPDU {
		slog.Info("sending data to", "id", c.Id, "tag", tag, "packet", hex.EncodeToString(packet))
	} else {
		slog.Info("sending data to", "id", c.Id, "tag", tag, "data", string(data))
	}
	_, err := c.Conn.Write(packet)
	return err
}

func (c *Conn) Read() (byte, []byte, error) {
	header := make([]byte, 3)
	_, err := c.Conn.Read(header)
	if err != nil {
		return 0, nil, err
	}
	if !c.isKnownTag(header[0]) {
		return 0, nil, ErrorTagUnknown
	}

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
	return header[0], data, nil
}

func (c *Conn) isKnownTag(tag byte) bool {
	for _, knownTag := range KnownTags {
		if tag == knownTag {
			return true
		}
	}
	return false
}

func (c *Conn) pack(tag byte, data []byte) []byte {
	var packet = make([]byte, len(data)+3)
	packet[0] = tag
	binary.LittleEndian.PutUint16(packet[1:], uint16(len(data)))
	copy(packet[3:], data)
	return packet
}

func (c *Conn) Close() error {
	defer c.Conn.Close()
	return c.Send(TagClose, nil)
}
