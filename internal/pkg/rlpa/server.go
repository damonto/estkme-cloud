package rlpa

import (
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/sqids/sqids-go"
)

type Server interface {
	Listen(address string) error
	Shutdown() error
}

type server struct {
	listener *net.TCPListener
	manager  Manager
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
	slog.Info("rLPA server is running on", "address", address)

	for {
		conn, err := s.listener.AcceptTCP()
		if err != nil {
			if err == io.EOF || errors.Is(err, net.ErrClosed) {
				return nil
			}
			conn.Close()
			return err
		}

		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(30 * time.Second)
		go s.handleConn(conn)
	}
}

func (s *server) handleConn(tcpConn *net.TCPConn) {
	id := s.id(tcpConn)
	conn := NewConnection(id, tcpConn)
	s.manager.Add(id, conn)
	slog.Info("new connection from", "id", id)
	defer conn.Close()
	defer s.manager.Remove(id)

	for {
		tag, data, err := conn.Read()
		if err != nil {
			if err == io.EOF || errors.Is(err, net.ErrClosed) {
				return
			}
			slog.Error("error reading from connection", "error", err)
			continue
		}

		if tag == TagClose {
			slog.Info("client closed connection", "id", id)
			return
		}
		if tag == TagAPDU {
			slog.Info("received data from", "id", id, "tag", tag, "data", hex.EncodeToString(data))
		} else {
			slog.Info("received data from", "id", id, "tag", tag, "data", string(data))
		}
		go conn.Dispatch(tag, data)
	}
}

func (s *server) id(tcpConn *net.TCPConn) string {
	sqid, _ := sqids.New(sqids.Options{
		MinLength: 6,
	})
	netAddr, _ := net.ResolveTCPAddr("tcp", tcpConn.RemoteAddr().String())
	id, _ := sqid.Encode([]uint64{uint64(netAddr.Port)})
	return id
}

func (s *server) Shutdown() error {
	for _, conn := range s.manager.GetAll() {
		conn.Close()
		s.manager.Remove(conn.Id)
	}
	return s.listener.Close()
}
