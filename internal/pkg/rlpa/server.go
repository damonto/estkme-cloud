package rlpa

import (
	"errors"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/google/uuid"
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
	slog.Info("listening on", "address", address)

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
	connectionId := uuid.New().String()
	conn := NewConnection(connectionId, tcpConn)
	s.manager.Add(connectionId, conn)
	slog.Info("new connection from", "id", connectionId)
	defer conn.Close()
	defer s.manager.Remove(connectionId)

	for {
		tag, data, err := conn.Read()
		if err != nil {
			if err == io.EOF || errors.Is(err, net.ErrClosed) {
				return
			}
			slog.Error("error reading from connection", "error", err)
			continue
		}
		slog.Info("received data from", "id", connectionId, "tag", tag, "data", data)
		go conn.Dispatch(tag, data)
	}
}

func (s *server) Shutdown() error {
	for _, conn := range s.manager.GetAll() {
		conn.Close()
		s.manager.Remove(conn.ConnectionId)
	}
	return s.listener.Close()
}
