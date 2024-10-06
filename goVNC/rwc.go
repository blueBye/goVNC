package main

import (
	"fmt"
	"net"
	"time"

	"github.com/amitbet/vncproxy/logger"
	"github.com/gorilla/websocket"
)

type RWC struct {
	WSC    *websocket.Conn
	Stream chan byte
}

func (c RWC) Write(p []byte) (int, error) {
	logger.Info("[rwc] Write")
	err := c.WSC.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (c RWC) Read(p []byte) (int, error) {
	logger.Info("[rwc] Read")

	n := len(p)
	for i := 0; i < n; i++ {
		b, ok := <-c.Stream
		if !ok {
			return i, fmt.Errorf("[rwc] channel closed, read %d bytes", i)
		}
		p[i] = b
	}
	return n, nil
}

func (c RWC) Close() error {
	logger.Info("[rwc] Close")
	return c.WSC.Close()
}

func (c RWC) LocalAddr() net.Addr {
	logger.Info("[rwc] LocalAddr")
	return nil
}

func (c RWC) RemoteAddr() net.Addr {
	logger.Info("[rwc] RemoteAddr")
	return nil
}

func (c RWC) SetDeadline(t time.Time) error {
	logger.Info("[rwc] SetDeadline")
	return nil
}

func (c RWC) SetReadDeadline(t time.Time) error {
	logger.Info("[rwc] SetReadDeadline")
	return nil
}

func (c RWC) SetWriteDeadline(t time.Time) error {
	return nil
}
