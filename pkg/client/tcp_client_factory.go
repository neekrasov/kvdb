package client

import "github.com/neekrasov/kvdb/internal/delivery/tcp"

type TCPClientFactory struct {
}

func (tcf *TCPClientFactory) Make(address string, opts ...tcp.ClientOption) (NetClient, error) {
	return tcp.NewClient(address, opts...)
}
