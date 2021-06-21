package socketIO

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
)

const SizeLen = 8

func Send(conn net.Conn, bytes []byte) error {
	size := len(bytes)
	// send the total size first
	sizeBytes := make([]byte, SizeLen)
	binary.BigEndian.PutUint64(sizeBytes, uint64(size))
	n, sizeErr := conn.Write(sizeBytes)
	fmt.Println("write size " + strconv.Itoa(n))
	if sizeErr != nil {
		return sizeErr
	}
	if n < len(sizeBytes) {
		return errors.New("send: read error of size")
	}
	// send content
	write := 0
	for write < size {
		n, writeErr := conn.Write(bytes[write:])
		if writeErr != nil {
			return writeErr
		}
		write += n
	}
	return nil
}

func Receive(conn net.Conn) ([]byte, error) {
	sizeBuf := make([]byte, SizeLen)
	n, sizeErr := conn.Read(sizeBuf)
	fmt.Println("read size " + strconv.Itoa(n))
	if sizeErr != nil {
		return nil, sizeErr
	}
	if n < SizeLen {
		return nil, errors.New("receive: read error of size")
	}
	size := binary.BigEndian.Uint64(sizeBuf)
	resp := make([]byte, size)
	read := uint64(0)
	for read < size {
		sizeRead, err := conn.Read(resp[read:])
		if err != nil {
			return nil, err
		}
		read += uint64(sizeRead)
	}
	return resp, nil
}
