package client

import (
	"net"
	"time"

	"github.com/amitbet/vncproxy/logger"
	"github.com/gorilla/websocket"
)

type RWC struct {
	WSC         *websocket.Conn
	Buffer      chan []byte
	InputStream chan []byte
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
	logger.Debug("[rwc] Read")
	msg := make(chan []byte, 1)
	select {
	case m := <-c.Buffer:
		msg <- m
	case m := <-c.InputStream:
		msg <- m
	}
	message := <-msg

	if len(message) == 0 {
		logger.Info("[rws] Read: Empty message")
		return 0, nil
	}

	logger.Debug("[rws] Read: message:", message)

	for idx := 0; idx < len(p); idx++ {
		p[idx] = message[idx]
	}

	if len(message) > len(p) {
		c.Buffer <- message[len(p):]
	}

	return len(p), nil
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
