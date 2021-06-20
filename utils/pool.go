package utils

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

type ConnectionPool struct {
	lock          *sync.Mutex
	ConnectionNum int
	ConnCreator   func() (net.Conn, error)
	Closed        bool
	Connections   chan net.Conn
	Errors        chan error
}

func NewConnectionPool(connectionNum int, createConnection func() (net.Conn, error)) *ConnectionPool {
	pool := &ConnectionPool{
		ConnectionNum: connectionNum,
		ConnCreator:   createConnection,
		Connections:   make(chan net.Conn, connectionNum),
		Errors:        make(chan error),
		Closed:        false,
		lock:          &sync.Mutex{},
	}
	for i := 0; i < connectionNum; i++ {
		go func() {
			conn, err := pool.ConnCreator()
			if err != nil {
				fmt.Println(err)
				pool.Errors <- err
				return
			}
			pool.Connections <- conn
		}()
	}
	return pool
}

func (pool *ConnectionPool) FetchConnection() (conn net.Conn, err error) {
	var closed bool
	pool.lock.Lock()
	closed = pool.Closed
	pool.lock.Unlock()
	if closed {
		err = errors.New("connection pool is closed")
		return
	}
	conn = <-pool.Connections
	return
}

func (pool *ConnectionPool) PutConnection(conn net.Conn) {
	var closed bool
	pool.lock.Lock()
	closed = pool.Closed
	pool.lock.Unlock()
	if closed {
		fmt.Println("Connection pool is closed")
		return
	}
	fmt.Println("put connection")
	pool.Connections <- conn
}

func (pool *ConnectionPool) Close() {
	if pool.Closed {
		return
	}
	pool.lock.Lock()
	if pool.ConnectionNum > 0 {
		pool.Closed = true
		for i := 0; i < pool.ConnectionNum; i++ {
			conn := <-pool.Connections
			err := conn.Close()
			if err != nil {
				pool.Errors <- err
				continue
			}
		}
		pool.ConnectionNum = 0
	}
	pool.lock.Unlock()
}

func (pool *ConnectionPool) LogErrors() {
	select {
	case err := <-pool.Errors:
		fmt.Println(err.Error())
	}
}
