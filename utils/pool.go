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

// pool.Close add lock or use atomic
// select <- pool.Connections

func NewConnectionPool(connectionNum int, createConnection func() (net.Conn, error)) *ConnectionPool {
	pool := &ConnectionPool{
		ConnectionNum: connectionNum,
		ConnCreator:   createConnection,
		Connections:   make(chan net.Conn),
		Errors:        make(chan error),
		Closed:        false,
		lock:          &sync.Mutex{},
	}
	for i := 0; i < connectionNum; i++ {
		go func() {
			conn, err := pool.ConnCreator()
			if err != nil {
				pool.Errors <- err
				return
			}
			pool.Connections <- conn
		}()
	}
	return pool
}

func (pool *ConnectionPool) FetchConnection() (net.Conn, error) {
	if pool.Closed {
		return nil, errors.New("connection pool is closed")
	}
	connection := <-pool.Connections
	return connection, nil
}

// TODO: put connection no need to handle error. For utility, sometimes no need to arise error, based on condition. Use log or warning
func (pool *ConnectionPool) PutConnection(conn net.Conn) error {
	var closed bool
	pool.lock.Lock()
	closed = pool.Closed
	pool.lock.Unlock()
	if closed {
		return errors.New("connection pool is closed")
	}
	fmt.Println("put connection")
	select {
	case <-pool.Connections:
		return nil
	case err := <-pool.Errors:
		return err
	}
}

func (pool *ConnectionPool) Close() {
	if pool.Closed {
		return
	}
	pool.lock.Lock()
	pool.Closed = true
	pool.lock.Unlock()
	if pool.Closed {
		for i := 0; i < pool.ConnectionNum; i++ {
			conn := <-pool.Connections
			err := conn.Close()
			if err != nil {
				// TODO: save the error, log
				continue
			}
		}
	}
}

func (pool *ConnectionPool) LogErrors() {
	select {
	case err := <-pool.Errors:
		fmt.Println(err.Error())
	}
}
