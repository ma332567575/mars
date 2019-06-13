package cluster

import (
	"fmt"

	"github.com/ma332567575/mars/conf"
	"github.com/ma332567575/mars/network"

	protobuf "github.com/golang/protobuf/proto"
)

var (
	server   *network.Server
	wsServer *network.WSServer
	clter    *cluster
)

type cluster struct {
	onMsg func(a network.Agent, msgID uint32, seqID uint32, body []byte)
}

func Init(onMsg func(a network.Agent, msgID uint32, seqID uint32, body []byte)) {
	clter = new(cluster)
	clter.onMsg = onMsg
	// server listen
	server = new(network.Server)
	server.Addr = conf.Addr
	server.NewAgent = newTcpAgent
	server.Parser = new(network.OGMsgParser)
	server.Start()

	// wsserver listen
	wsServer = new(network.WSServer)
	wsServer.Addr = conf.WSAddr
	wsServer.NewAgent = newWSAgent
	wsServer.Parser = new(network.OGMsgParser)
	wsServer.Start()
}

func Destroy() {
	// close server
	if server != nil {
		server.Close()
	}
}

type agent struct {
	iclient network.IClient
}

func (a *agent) Run() {
	for {
		msg, err := a.iclient.ReadMsg()
		if err != nil {
			fmt.Print(err.Error())
			if err.Error() == "socket close" {
				break
			}
			continue
		}
		msgUse := msg.(*network.OGMsg)

		clter.onMsg(a, msgUse.MsgID, msgUse.SeqID, msgUse.Body)
	}
}

func (a *agent) OnClose() {

}

func (a *agent) SendMsg(msgID uint32, seqID uint32, msg protobuf.Message) {
	flymsg := new(network.OGPBMsg)
	flymsg.Msg = msg
	flymsg.MsgID = msgID
	flymsg.SeqID = seqID
	a.iclient.WriteMsg(flymsg)
}

func newTcpAgent(client *network.Client) network.Agent {
	a := new(agent)
	a.iclient = client
	return a
}

func newWSAgent(wsclient *network.WSClient) network.Agent {
	a := new(agent)
	a.iclient = wsclient
	return a
}
