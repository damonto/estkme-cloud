package rlpa

import (
	"errors"
	"sync"
)

type Manager interface {
	Add(id string, connection *Conn)
	Remove(id string)
	Get(id string) (*Conn, error)
	GetAll() []*Conn
	Len() int
}

var (
	ErrConnNotFound = errors.New("connection not found")
)

type manager struct {
	connections sync.Map
}

func NewManager() Manager {
	return &manager{}
}

func (m *manager) Add(id string, connection *Conn) {
	m.connections.Store(id, connection)
}

func (m *manager) Remove(id string) {
	m.connections.Delete(id)
}

func (m *manager) Get(id string) (*Conn, error) {
	conn, ok := m.connections.Load(id)
	if !ok {
		return nil, ErrConnNotFound
	}
	return conn.(*Conn), nil
}

func (m *manager) GetAll() []*Conn {
	var connections []*Conn
	m.connections.Range(func(_, value interface{}) bool {
		connections = append(connections, value.(*Conn))
		return true
	})
	return connections
}

func (m *manager) Len() int {
	length := 0
	m.connections.Range(func(_, _ interface{}) bool {
		length++
		return true
	})
	return length
}
