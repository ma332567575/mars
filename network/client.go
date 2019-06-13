package network

import (
	"errors"
	"fmt"
	"net"
)

type Client struct {
	conn      net.Conn
	writeChan chan []byte
	parser    MsgParser
}

type ClientSet map[net.Conn]struct{}

func newClient(conn net.Conn, pendingWriteNum int, parser MsgParser) *Client {
	client := new(Client)
	client.conn = conn
	client.parser = parser
	client.writeChan = make(chan []byte, pendingWriteNum)

	// write data
	go func() {
		for b := range client.writeChan {
			if b == nil {
				break
			}

			_, err := client.conn.Write(b)
			if err != nil {
				break
			}
		}

		client.conn.Close()
	}()

	return client
}

func (client *Client) Read() ([]byte, error) {
	data := make([]byte, 1000)
	n, err := client.conn.Read(data)
	if err != nil {
		// read fail. close conn
		client.conn.Close()
		fmt.Printf("socket close")
		return nil, errors.New("socket close")
	}
	fmt.Printf("receive %d(byte)", n)

	return data, nil
}

func (client *Client) Write(data []byte) {
	// use channel send
	client.writeChan <- data
}

func (client *Client) ReadMsg() (interface{}, error) {
	if client.parser == nil {
		return nil, errors.New("client.parser no found")
	}
	data, err := client.Read()
	if err != nil {
		return nil, err
	}
	return client.parser.Parse(data)
}

func (client *Client) WriteMsg(msg interface{}) error {
	if client.parser == nil {
		return errors.New("client.parser not found")
	}
	data, err := client.parser.Package(msg)
	if err != nil {
		return err
	}
	client.Write(data)
	return nil
}

func (client *Client) Close() {
	client.Write(nil)
}
