/* header 3 items
typedef struct tagOGHEADER
{
	DWORD dwType;		//消息操作类型及消息值
	DWORD dwLength;		//消息头结构后跟的消息体字节长度，不包括头本身长度
	DWORD dwSeqID;		//请求方标示的消息序列号，接收方不允许修改
}
*/
package network

import (
	"encoding/binary"
	"errors"

	protobuf "github.com/golang/protobuf/proto"
)

type OGMsg struct {
	MsgID uint32
	SeqID uint32
	Body  []byte
}

type OGPBMsg struct {
	MsgID uint32
	SeqID uint32
	Msg   protobuf.Message
}

type OGMsgParser struct {
}

func (parser *OGMsgParser) Parse(data []byte) (interface{}, error) {
	// msg too short
	if len(data) < 12 {
		return nil, errors.New("too short")
	}

	// read header
	header := data[0:12]
	msgType := binary.LittleEndian.Uint32(header[0:4])
	msgLength := binary.LittleEndian.Uint32(header[4:8])
	msgSeqID := binary.LittleEndian.Uint32(header[8:12])

	// read body
	body := data[12 : 12+msgLength]
	if len(body) < int(msgLength) {
		return nil, errors.New("body too short")
	}
	// protobuf
	msg := new(OGMsg)
	msg.MsgID = msgType
	msg.SeqID = msgSeqID
	msg.Body = body
	return msg, nil
}

func (parser *OGMsgParser) Package(msgInter interface{}) ([]byte, error) {
	// header
	msg := msgInter.(*OGPBMsg)
	data, err := protobuf.Marshal(msg.Msg)
	if err != nil {
		return nil, errors.New("pb marshal fail")
	}
	buffer := make([]byte, 12+len(data))
	binary.LittleEndian.PutUint32(buffer, msg.MsgID)
	binary.LittleEndian.PutUint32(buffer, uint32(len(data)))
	binary.LittleEndian.PutUint32(buffer, msg.SeqID)
	// body
	copy(buffer[12:], data)
	return buffer, nil
}
