package rlpa

type Handler = func(conn *Conn, data []byte) error
