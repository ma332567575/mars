package network

import (
	"log"
	"net"
	"sync"
)

type Server struct {
	Addr            string
	ln              net.Listener
	PendingWriteNum int
	NewAgent        func(agent *Client) Agent
	Parser          MsgParser
	mutexConns      sync.Mutex
	conns           ClientSet
	cwg             sync.WaitGroup
	rwg             sync.WaitGroup
}

func (server *Server) Start() {
	server.init()
	go server.run()
}

func (server *Server) Close() {
	server.rwg.Wait()
	server.ln.Close()
	// close all client
	server.mutexConns.Lock()
	for k, _ := range server.conns {
		k.Close()
	}
	server.conns = nil
	server.mutexConns.Unlock()
	server.cwg.Wait()
}

func (server *Server) init() {
	// listen
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatal("%v", err)
	}

	server.ln = ln
	server.conns = make(ClientSet)
}

func (server *Server) run() {
	server.rwg.Add(1)
	defer server.rwg.Done()
	// loop for accpet client
	for {
		conn, err := server.ln.Accept()
		if err != nil {
			continue
		}
		server.cwg.Add(1)
		server.mutexConns.Lock()
		server.conns[conn] = struct{}{}
		server.mutexConns.Unlock()
		// create client agent
		client := newClient(conn, server.PendingWriteNum, server.Parser)
		// notify a agent come
		agent := server.NewAgent(client)
		go func() {
			agent.Run()

			// cleanup
			client.Close()
			server.mutexConns.Lock()
			delete(server.conns, conn)
			server.mutexConns.Unlock()
			agent.OnClose()

			server.cwg.Done()
		}()
	}
}
