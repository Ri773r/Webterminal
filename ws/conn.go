package ws

import (
	"fmt"
	"io"

	"github.com/gorilla/websocket"
)

// WsWrapper Websocket wrapper
type WsWrapper struct {
	*websocket.Conn
	// resize chan
	ResizeCh chan map[string]int
	// heartbeat chan
	HeartbeatCh chan struct{}
	// before read hook
	PreRead func(io.Reader) error
	// after read hook
	// if return 0, eof return eof (eof will end the session) ;
	// if return 0, err enter the session as is ;
	// if return <=0, nil do nothing ;
	// if want to return the processed data
	// first copy(temP, byte(data)) (temP is third parameter)
	// then return len(byte(data)), nil ;
	PostRead func([]byte, int, []byte, chan<- map[string]int, chan<- struct{}, chan<- struct{}) (int, error)
	// quit chan
	QuitChan chan struct{}
	// timeout chan
	TimeoutChan chan struct{}
}

// Read Implement the Read interface
func (wsw *WsWrapper) Read(p []byte) (int, error) {
	for {
		msgType, reader, err := wsw.Conn.NextReader()
		if err != nil {
			fmt.Println(err)
			return 0, err
		}
		fmt.Println("msgType", msgType)
		// if msgType != websocket.TextMessage {
		// 	continue
		// }

		// Before reading hook
		if wsw.PreRead != nil {
			err = wsw.PreRead(reader)
			if err != nil {
				fmt.Println("ws pre read failed ...", err)
				return 0, io.EOF
			}
		}

		n, err := reader.Read(p)
		fmt.Println("read from ws ", string(p[:n]), "length ", n)
		if err != nil {
			fmt.Println("read from ws failed ...", err)
			return 0, err
		}

		// After read hook
		if wsw.PostRead != nil {
			tempP := make([]byte, n)
			n2, err2 := wsw.PostRead(p, n, tempP, wsw.ResizeCh, wsw.HeartbeatCh, wsw.TimeoutChan)
			if err2 != nil {
				fmt.Println(err)
				if err2 == io.EOF {
					return 0, io.EOF
				}
				// if other err do nothing
				return n, err
			}
			if n2 > 0 {
				copy(p, tempP)
				return n2, nil
			}
			return 0, nil
		}

		return n, err
	}
}

// Write Implement the write interface
func (wsw *WsWrapper) Write(p []byte) (int, error) {
	writer, err := wsw.Conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	defer writer.Close()
	return writer.Write(p)
}

// NewWsWrapper Return a WsWrapper object pointer
func NewWsWrapper(conn *websocket.Conn) *WsWrapper {
	return &WsWrapper{
		Conn:        conn,
		HeartbeatCh: make(chan struct{}),
		ResizeCh:    make(chan map[string]int),
		QuitChan:    make(chan struct{}),
		TimeoutChan: make(chan struct{}),
	}
}

// SetPreReadHandle Set pre before read
func (wsw *WsWrapper) SetPreReadHandle(f func(io.Reader) error) {
	wsw.PreRead = f
}

// SetPostReadHandle Set up after read
func (wsw *WsWrapper) SetPostReadHandle(f func([]byte, int, []byte, chan<- map[string]int, chan<- struct{}, chan<- struct{}) (int, error)) {
	wsw.PostRead = f
}
