package network

import (
	protobuf "github.com/golang/protobuf/proto"
)

type Agent interface {
	Run()
	OnClose()
	SendMsg(msgID uint32, seqID uint32, msg protobuf.Message)
}

type IClient interface {
	ReadMsg() (interface{}, error)
	WriteMsg(msg interface{}) error
	Close()
}
