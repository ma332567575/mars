package network

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/gorilla/websocket"
)

type WebsocketClientSet map[*websocket.Conn]struct{}

type WSClient struct {
	sync.Mutex
	conn      *websocket.Conn
	writeChan chan []byte
	maxMsgLen uint32
	closeFlag bool
	parser    MsgParser
}

func newWSClient(conn *websocket.Conn, pendingWriteNum int, maxMsgLen uint32, parser MsgParser) *WSClient {
	wsClient := new(WSClient)
	wsClient.conn = conn
	wsClient.writeChan = make(chan []byte, pendingWriteNum)
	wsClient.maxMsgLen = maxMsgLen
	wsClient.parser = parser

	go func() {
		for b := range wsClient.writeChan {
			if b == nil {
				break
			}

			err := conn.WriteMessage(websocket.BinaryMessage, b)
			if err != nil {
				break
			}
		}

		conn.Close()
		wsClient.Lock()
		wsClient.closeFlag = true
		wsClient.Unlock()
	}()

	return wsClient
}

func (wsClient *WSClient) doDestroy() {
	wsClient.conn.UnderlyingConn().(*net.TCPConn).SetLinger(0)
	wsClient.conn.Close()

	if !wsClient.closeFlag {
		close(wsClient.writeChan)
		wsClient.closeFlag = true
	}
}

func (wsClient *WSClient) Destroy() {
	wsClient.Lock()
	defer wsClient.Unlock()

	wsClient.doDestroy()
}

func (wsClient *WSClient) Close() {
	wsClient.Lock()
	defer wsClient.Unlock()
	if wsClient.closeFlag {
		return
	}

	wsClient.doWrite(nil)
	wsClient.closeFlag = true
}

func (wsClient *WSClient) doWrite(b []byte) {
	if len(wsClient.writeChan) == cap(wsClient.writeChan) {
		fmt.Printf("close conn: channel full")
		wsClient.doDestroy()
		return
	}

	wsClient.writeChan <- b
}

// goroutine not safe
func (wsClient *WSClient) ReadMsg() (interface{}, error) {
	_, b, err := wsClient.conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	return wsClient.parser.Parse(b)
}

// args must not be modified by the others goroutines
func (wsClient *WSClient) WriteMsg(msg interface{}) error {
	wsClient.Lock()
	defer wsClient.Unlock()
	if wsClient.closeFlag {
		return nil
	}

	data, err := wsClient.parser.Package(msg)
	if err != nil {
		return err
	}
	// get len
	msgLen := len(data)

	// check len
	if uint32(msgLen) > wsClient.maxMsgLen {
		return errors.New("message too long")
	} else if msgLen < 1 {
		return errors.New("message too short")
	}

	wsClient.doWrite(data)

	return nil
}
