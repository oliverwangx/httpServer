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
		Connections:   make(chan net.Conn, connectionNum),
		Errors:        make(chan error),
		Closed:        false,
		lock:          &sync.Mutex{},
	}
	fmt.Println(connectionNum)
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
	if pool.Closed {
		err = errors.New("connection pool is closed")
		return
	}
	conn = <-pool.Connections
	return
}

// TODO: put connection no need to handle error. For utility, sometimes no need to arise error, based on condition. Use log or warning
func (pool *ConnectionPool) PutConnection(conn net.Conn) (err error) {
	var closed bool
	pool.lock.Lock()
	closed = pool.Closed
	pool.lock.Unlock()
	if closed {
		err = errors.New("connection pool is closed")
		return
	}
	fmt.Println("put connection")
	pool.Connections <- conn
	return
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
				pool.Errors <- err
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
