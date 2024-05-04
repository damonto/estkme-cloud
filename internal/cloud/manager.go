package cloud

import (
	"errors"
	"sync"
)

type Manager interface {
	Add(id string, conn *Conn)
	Remove(id string)
	Get(id string) (*Conn, error)
	GetAll() []*Conn
	Len() int
}

var (
	ErrConnNotFound = errors.New("conn not found")
)

type manager struct {
	conns sync.Map
}

func NewManager() Manager {
	return &manager{}
}

func (m *manager) Add(id string, conn *Conn) {
	m.conns.Store(id, conn)
}

func (m *manager) Remove(id string) {
	m.conns.Delete(id)
}

func (m *manager) Get(id string) (*Conn, error) {
	conn, ok := m.conns.Load(id)
	if !ok {
		return nil, ErrConnNotFound
	}
	return conn.(*Conn), nil
}

func (m *manager) GetAll() []*Conn {
	var conns []*Conn
	m.conns.Range(func(_, value interface{}) bool {
		conns = append(conns, value.(*Conn))
		return true
	})
	return conns
}

func (m *manager) Len() int {
	length := 0
	m.conns.Range(func(_, _ interface{}) bool {
		length++
		return true
	})
	return length
}
