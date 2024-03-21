package rlpa

import (
	"errors"
	"sync"
)

type Manager interface {
	Add(id string, connection *Connection)
	Remove(id string)
	Get(id string) (*Connection, error)
	GetAll() []*Connection
	Len() int
}

const (
	ErrConnectionNotFound = "connection not found"
)

type manager struct {
	connections sync.Map
}

func NewManager() Manager {
	return &manager{}
}

func (m *manager) Add(id string, connection *Connection) {
	m.connections.Store(id, connection)
}

func (m *manager) Remove(id string) {
	m.connections.Delete(id)
}

func (m *manager) Get(id string) (*Connection, error) {
	conn, ok := m.connections.Load(id)
	if !ok {
		return nil, errors.New(ErrConnectionNotFound)
	}
	return conn.(*Connection), nil
}

func (m *manager) GetAll() []*Connection {
	var connections []*Connection
	m.connections.Range(func(_, value interface{}) bool {
		connections = append(connections, value.(*Connection))
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
