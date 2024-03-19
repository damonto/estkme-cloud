package rlpa

import (
	"errors"
	"sync"
)

type Manager interface {
	Add(connectionId string, connection *Connection)
	Remove(connectionId string)
	Get(connectionId string) (*Connection, error)
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

func (m *manager) Add(connectionId string, connection *Connection) {
	m.connections.Store(connectionId, connection)
}

func (m *manager) Remove(connectionId string) {
	m.connections.Delete(connectionId)
}

func (m *manager) Get(connectionId string) (*Connection, error) {
	conn, ok := m.connections.Load(connectionId)
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
