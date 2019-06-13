package network

type MsgParser interface {
	// parse body from msgdata
	Parse(data []byte) (interface{}, error)

	// package struct to msgdata
	Package(interface{}) ([]byte, error)
}
