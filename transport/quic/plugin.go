package quic

import (
	"github.com/stack-labs/stack/plugin"
	"github.com/stack-labs/stack/transport"
)

type quicTransportPlugin struct {
}

func (q *quicTransportPlugin) Name() string {
	return "quic"
}

func (q *quicTransportPlugin) Options() []transport.Option {
	return nil
}

func (q *quicTransportPlugin) New(opts ...transport.Option) transport.Transport {
	return NewTransport(opts...)
}

func init() {
	plugin.TransportPlugins["quic"] = &quicTransportPlugin{}
}
