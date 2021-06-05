package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
)

type ConnectionPool struct {
	lock           *sync.Mutex
	MinConnections int
	MaxConnections int
	ConnectionNum  int
	ConnCreator    func() (net.Conn, error)
	Closed         bool
	Connections    chan net.Conn
}

// pool.Close add lock or use atomic
// select <- pool.Connections

// TODO: implement newConnectionPool() and return a pool, put craete connection into a goroutine

func NewConnectionPool(minConnections int, maxConnections int, createConnection func() (net.Conn, error)) *ConnectionPool {
	pool := &ConnectionPool{}
	pool.MinConnections = minConnections
	pool.ConnectionNum = minConnections
	pool.MaxConnections = maxConnections
	pool.Connections = make(chan net.Conn, maxConnections)
	pool.Closed = false
	pool.ConnCreator = createConnection
	pool.lock = &sync.Mutex{}
	for i := 0; i < minConnections; i++ {
		go func() {
			conn, err := pool.ConnCreator()
			if err != nil {
				fmt.Println("Connection Pool, init: " + err.Error())
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
	// errChan = make(chan error) handle error
	// concurrent
	// var num int
	// pool.lock.Lock()
	// num = pool.ConnectionNum

	// pool.lock.Unlock()

	// created connection fetched by other goroutine

	if pool.ConnectionNum < pool.MaxConnections {
		go func() {
			conn, err := pool.ConnCreator()
			if err != nil {
				fmt.Println("Connection Pool, fetch connection: " + err.Error())
				// errChan <- err
			}
			fmt.Println("Create a new connection")
			pool.Connections <- conn
			pool.lock.Lock()
			pool.ConnectionNum++
			pool.lock.Unlock()

		}()
	}
	fmt.Println("Connetion pool number :" + strconv.Itoa(pool.ConnectionNum) + " maxConnection: " + strconv.Itoa(pool.MaxConnections))
	connection := <-pool.Connections
	return connection, nil
}

// TODO: put connection no need to handle error. For utility, sometimes no need to arise error, based on condition. Use log or warning
func (pool *ConnectionPool) PutConnection(conn net.Conn) error {
	if pool.Closed {
		return errors.New("connection pool is closed")
	}
	// TODO: conn.Closed instead of conn == nil
	if conn == nil {
		fmt.Println("put connection: connection is nil")
		pool.lock.Lock()
		pool.ConnectionNum -= 1
		pool.lock.Unlock()
		return errors.New("nil connection detected")
	}
	fmt.Println("put connection")
	pool.Connections <- conn
	return nil
}

// TODO: same as put connection
func (pool *ConnectionPool) Close() {
	if pool.Closed {
		return
	}
	pool.Closed = true
	for i := 0; i < pool.ConnectionNum; i++ {
		conn := <-pool.Connections
		err := conn.Close()
		if err != nil {
			// return err

			// save the error, log
			continue
		}
	}
}
