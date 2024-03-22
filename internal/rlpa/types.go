package rlpa

type Handler = func(conn *Connection, data []byte) error
