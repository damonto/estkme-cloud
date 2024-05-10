package cloud

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"log/slog"
	"net"
	"sync"

	"github.com/damonto/estkme-cloud/internal/driver"
)

type Handler = func(ctx context.Context, conn *Conn, data []byte) error

type Conn struct {
	Id       string
	Conn     *net.TCPConn
	APDU     driver.APDU
	lock     sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	handlers map[Tag]Handler
}

var (
	ErrTagUnknown = errors.New("unknown tag")
)

func NewConn(id string, conn *net.TCPConn) *Conn {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Conn{
		Id:       id,
		Conn:     conn,
		handlers: make(map[Tag]Handler, len(KnownTags)),
		ctx:      ctx,
		cancel:   cancel,
	}
	c.APDU = NewAPDU(c)
	c.registerHandlers()
	return c
}

func (c *Conn) registerHandlers() {
	c.RegisterHandler(TagManagement, func(ctx context.Context, conn *Conn, data []byte) error {
		return conn.Send(TagMessageBox, []byte("Welcome! \n You are connected to the server. \n Here is your PIN\n"+conn.Id))
	})

	c.RegisterHandler(TagProcessNotification, func(ctx context.Context, conn *Conn, data []byte) error {
		defer conn.Close()
		conn.Send(TagMessageBox, []byte("Processing notifications..."))
		if err := processNotification(ctx, conn); err != nil {
			slog.Error("failed to process notification", "error", err)
			return conn.Send(TagMessageBox, []byte("Process failed \n"+ToTitle(err.Error())))
		}
		return conn.Send(TagMessageBox, []byte("All notifications have been processed successfully"))
	})

	c.RegisterHandler(TagDownloadProfile, func(ctx context.Context, conn *Conn, data []byte) error {
		defer conn.Close()
		conn.Send(TagMessageBox, []byte("Your profile is being downloaded. \n Please wait..."))
		if err := downloadProfile(ctx, conn, data); err != nil {
			slog.Error("failed to download profile", "error", err)
			return conn.Send(TagMessageBox, []byte("Download failed \n"+ToTitle(err.Error())))
		}
		return conn.Send(TagMessageBox, []byte("Your profile has been downloaded successfully"))
	})
}

func (c *Conn) RegisterHandler(tag Tag, handler Handler) error {
	if !c.isKnownTag(tag) {
		return ErrTagUnknown
	}
	c.handlers[tag] = handler
	return nil
}

func (c *Conn) Handle(tag Tag, data []byte) {
	if tag == TagAPDU {
		c.APDU.Receive() <- data
	}
	if handler, ok := c.handlers[tag]; ok {
		if err := handler(c.ctx, c, data); err != nil {
			slog.Error("error handling tag", "tag", tag, "data", data, "error", err)
		}
	}
}

func (c *Conn) Send(tag Tag, data []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	packet := c.pack(tag, data)
	if tag == TagAPDU {
		slog.Debug("sending data to", "id", c.Id, "tag", tag, "packet", hex.EncodeToString(packet))
	} else {
		slog.Debug("sending data to", "id", c.Id, "tag", tag, "data", string(data))
	}
	_, err := c.Conn.Write(packet)
	return err
}

func (c *Conn) Read() (Tag, []byte, error) {
	header := make([]byte, 3)
	_, err := c.Conn.Read(header)
	if err != nil {
		return 0, nil, err
	}
	if !c.isKnownTag(Tag(header[0])) {
		return 0, nil, ErrTagUnknown
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
	return Tag(header[0]), data, nil
}

func (c *Conn) isKnownTag(tag Tag) bool {
	for _, knownTag := range KnownTags {
		if tag == knownTag {
			return true
		}
	}
	return false
}

func (c *Conn) pack(tag Tag, data []byte) []byte {
	var packet = make([]byte, len(data)+3)
	packet[0] = byte(tag)
	binary.LittleEndian.PutUint16(packet[1:], uint16(len(data)))
	copy(packet[3:], data)
	return packet
}

func (c *Conn) Close() error {
	select {
	case <-c.ctx.Done():
		return nil
	default:
		err := c.Send(TagClose, nil)
		if err != nil {
			return c.Conn.Close()
		}
		c.cancel()
		return err
	}
}
