package cloud

import (
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"math/rand/v2"
	"net"
	"os"
	"sync"
	"time"
)

type Server interface {
	Listen(address string) error
	Shutdown() error
}

type server struct {
	listener *net.TCPListener
	manager  Manager
	wg       sync.WaitGroup
}

func NewServer(manager Manager) Server {
	return &server{manager: manager}
}

func (s *server) Listen(address string) error {
	var err error
	tcpAddr, _ := net.ResolveTCPAddr("tcp", address)
	s.listener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}
	slog.Info("eSTK.me cloud enhance server is running on", "address", address)

	for {
		conn, err := s.listener.AcceptTCP()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			conn.Close()
			return err
		}

		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(30 * time.Second)
		// TODO: Delete this line when new eSTK.me firmware is released. (which will support the heartbeat tag)
		conn.SetReadDeadline(time.Now().Add(2 * time.Minute))

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleConn(conn)
		}()
	}
}

func (s *server) handleConn(tcpConn *net.TCPConn) {
	id := s.id()
	slog.Info("new connection from", "id", id)
	conn := NewConn(id, tcpConn)
	s.manager.Add(id, conn)
	defer conn.Close()
	defer s.manager.Remove(id)

	for {
		tag, data, err := conn.Read()
		if err != nil {
			if err == io.EOF || errors.Is(err, net.ErrClosed) || os.IsTimeout(err) {
				return
			}
			if !errors.Is(err, ErrorTagUnknown) {
				slog.Error("error reading from connection", "error", err)
			}
			continue
		}

		if tag == TagClose {
			slog.Info("client closed connection", "id", id)
			return
		}
		if tag == TagAPDU {
			slog.Debug("received data from", "id", id, "tag", tag, "data", hex.EncodeToString(data))
		} else {
			slog.Debug("received data from", "id", id, "tag", tag, "data", string(data))
		}
		go conn.Handle(tag, data)
	}
}

func (s *server) id() string {
	seeds := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	id := make([]rune, 6)
	for i := range id {
		id[i] = seeds[rand.IntN(len(seeds))]
	}
	if _, err := s.manager.Get(string(id)); errors.Is(err, ErrConnNotFound) {
		return string(id)
	} else {
		return s.id()
	}
}

func (s *server) Shutdown() error {
	if err := s.listener.Close(); err != nil {
		return err
	}
	slog.Info("waiting for all connections to close, please wait...", "count", s.manager.Len())
	s.wg.Wait()
	for _, conn := range s.manager.GetAll() {
		conn.Close()
		s.manager.Remove(conn.Id)
	}
	return nil
}
